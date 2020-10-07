package src

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type CommandUnlock struct{}

func (this *CommandUnlock) Help() string {
	return "Usage: terrarium [-s3-bucket] [-s3-region] unlock <directory>"
}

func (this *CommandUnlock) Run(config *Config, args []string) error {
	if len(args) != 2 {
		return errors.New("Expected one positional argument for 'unlock' command: <directory>")
	}

	config.Directory = strings.Trim(args[1], "/") + "/"
	lockPath := config.Directory + ".lock"

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		return err
	}

	// Create S3 service client.
	svc := s3.New(sess)

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
