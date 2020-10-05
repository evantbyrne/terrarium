package src

import (
	"errors"
)

type Config struct {
	Directory   string
	Expires     uint64
	S3AccessKey string
	S3Bucket    string
	S3Region    string
	S3SecretKey string
}

func (this *Config) SetS3AccessKey(value string) error {
	if value == "" {
		return errors.New("A value is required for the 'AWS_ACCESS_KEY' environment variable.")
	}
	this.S3AccessKey = value
	return nil
}

func (this *Config) SetExpires(value uint64) error {
	if value < 1 {
		return errors.New("The value of the '-expires' flag must be non-zero.")
	}
	this.Expires = value
	return nil
}

func (this *Config) SetS3Bucket(value string) error {
	if value == "" {
		return errors.New("A value is required for the '-s3-bucket' flag.")
	}
	this.S3Bucket = value
	return nil
}

func (this *Config) SetS3Region(value string) error {
	if value == "" {
		return errors.New("A value is required for the '-s3-region' flag.")
	}
	this.S3Region = value
	return nil
}

func (this *Config) SetS3SecretKey(value string) error {
	if value == "" {
		return errors.New("A value is required for the 'AWS_SECRET_KEY' environment variable.")
	}
	this.S3SecretKey = value
	return nil
}
