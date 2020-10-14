package src

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type CommandCommit struct{}

func (this *CommandCommit) Help() string {
	return "Usage: terrarium [-s3-bucket] [-s3-region] commit <directory>"
}

func (this *CommandCommit) Run(config *Config, args []string) error {
	if len(args) != 2 {
		return errors.New("Expected one positional argument for 'commit' command: <directory>")
	}

	now := time.Now()
	config.Directory = strings.Trim(args[1], "/") + "/"
	backupPath := config.Directory + fmt.Sprintf("backups/%d/", now.Unix())
	lockPath := config.Directory + ".lock"
	statePath := config.Directory + "state/"

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		return err
	}

	// Create S3 service client.
	svc := s3.New(sess)

	// Create backup.
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(config.S3Bucket),
		Delimiter: aws.String(""),
		Prefix:    aws.String(statePath),
	})
	if err != nil {
		return err
	}
	for _, item := range resp.Contents {
		// Skip directories.
		if strings.HasSuffix(*item.Key, "/") {
			continue
		}
		backupKey := strings.TrimPrefix(*item.Key, statePath)
		_, err = svc.CopyObject(&s3.CopyObjectInput{
			Bucket:     aws.String(config.S3Bucket),
			CopySource: aws.String(fmt.Sprintf("/%s/%s", config.S3Bucket, *item.Key)),
			Key:        aws.String(fmt.Sprintf("%s/%s", backupPath, backupKey)),
		})
		if err != nil {
			return err
		}
	}

	// Delete old state.
	if len(resp.Contents) > 0 {
		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(config.S3Bucket),
			Delete: &s3.Delete{
				Objects: []*s3.ObjectIdentifier{},
			},
		}
		for _, item := range resp.Contents {
			deleteInput.Delete.Objects = append(deleteInput.Delete.Objects, &s3.ObjectIdentifier{
				Key: aws.String(*item.Key),
			})
		}
		_, err = svc.DeleteObjects(deleteInput)
		if err != nil {
			return err
		}
	}

	// Upload new state.
	di := NewDirectoryIterator(config.S3Bucket, statePath)
	uploader := s3manager.NewUploader(sess)
	if err := uploader.UploadWithIterator(aws.BackgroundContext(), di); err != nil {
		return err
	}

	// Delete lock file.
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.S3Bucket),
		Key:    aws.String(lockPath),
	})
	if err != nil {
		return err
	}

	return nil
}
