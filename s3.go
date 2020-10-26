/*
usher is a tiny personal url shortener.

This file contains functions for pushing database mappings to Amazon S3.
*/

package usher

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (db *DB) pushS3Mapping(ctx context.Context, awsS3 *s3.S3, config *ConfigEntry, code, url string) error {
	_, err := awsS3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(db.Domain),
		ContentType: aws.String("text/plain"),
		Key:         aws.String(code),
		WebsiteRedirectLocation: aws.String(url),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			// If the SDK can determine the request or retry delay was canceled
			// by a context the CanceledErrorCode error code will be returned.
			return fmt.Errorf("push of %q cancelled due to timeout\n", code)
		} else {
			return fmt.Errorf("push of %q failed: %s\n", code, err)
		}
	}

	return nil
}

func (db *DB) pushS3(config *ConfigEntry) error {
	// Setup background context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create AWS session and S3 objects
	awsSession := session.Must(session.NewSession())
	awsCredentials := credentials.NewStaticCredentials(config.AWSKey, config.AWSSecret, "")
	awsS3 := s3.New(awsSession, &aws.Config{
		Credentials: awsCredentials,
		Region:      aws.String(config.AWSRegion),
	})

	// Read all mappings
	mappings, err := db.readDB()
	if err != nil {
		return err
	}

	// Push each code-url pair to s3
	for code, url := range mappings {
		//fmt.Printf("+ pushing %s => %s\n", code, url)
		err = db.pushS3Mapping(ctx, awsS3, config, code, url)
		if err != nil {
			return err
		}
	}

	return nil
}
