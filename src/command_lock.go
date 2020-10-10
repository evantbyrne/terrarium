package src

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type CommandLock struct{}

func (this *CommandLock) Help() string {
	return "Usage: terrarium [-expires] [-retry-wait] [-s3-bucket] [-s3-region] [-timeout] lock <directory>"
}

func (this *CommandLock) Run(config *Config, args []string) error {
	if len(args) != 2 {
		return errors.New("Expected one positional argument for 'lock' command: <directory>")
	}

	config.Directory = strings.Trim(args[1], "/") + "/"
	lockPath := config.Directory + ".lock"
	statePath := config.Directory + "state/"
	timeoutExpires := time.Now().Add(time.Duration(int64(time.Second) * int64(config.Timeout))).Unix()

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		return err
	}

	// Wait for remote state to unlock.
	unlocked := false
	for !unlocked {
		unlocked, err = CheckS3Unlocked(config, lockPath, sess)
		if err != nil {
			fmt.Println("S3 error.")
			return err
		}

		if !unlocked {
			if config.Timeout == 0 {
				return errors.New("Could not unlock remote state.")
			}
			if time.Now().Unix() > timeoutExpires {
				return errors.New("Timeout exceeded for unlocking remote state.")
			}

			time.Sleep(time.Duration(int64(time.Second) * int64(config.RetriesWait)))
		}
	}

	// Upload the lock.
	lockExpires := time.Now().Add(time.Duration(int64(time.Second) * int64(config.Expires))).Unix()
	fr := bytes.NewReader([]byte(strconv.FormatInt(lockExpires, 10)))
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:   fr,
		Bucket: aws.String(config.S3Bucket),
		Key:    aws.String(lockPath),
	})
	fmt.Printf("Locked to: %d\n", lockExpires)

	// Create S3 service client.
	svc := s3.New(sess)

	// Get the list of files.
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

		// Create directory.
		if err := os.MkdirAll(path.Dir(*item.Key), 0770); err != nil {
			return err
		}

		// Download file.
		fw, err := os.Create(*item.Key)
		if err != nil {
			return err
		}
		defer fw.Close()

		downloader := s3manager.NewDownloader(sess)
		_, err = downloader.Download(fw,
			&s3.GetObjectInput{
				Bucket: aws.String(config.S3Bucket),
				Key:    aws.String(*item.Key),
			})
		if err != nil {
			return err
		}
	}

	return nil
}

func CheckS3Unlocked(config *Config, lockPath string, sess *session.Session) (bool, error) {
	fw := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloader(sess)
	_, err := downloader.Download(fw,
		&s3.GetObjectInput{
			Bucket: aws.String(config.S3Bucket),
			Key:    aws.String(lockPath),
		})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			// Lock file does not exist.
			return true, nil
		} else {
			// Some other S3 error.
			return false, err
		}
	}

	expires, err := strconv.ParseInt(string(fw.Bytes()), 10, 64)
	if err != nil {
		return false, err
	}

	if expires <= time.Now().Unix() {
		return true, nil
	}

	return false, nil
}
