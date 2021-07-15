package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

func GetKeyForAllCredentials(userID string) string {
	return fmt.Sprintf("users/%s", userID)
}

func GetKeyForSingleCredential(userId string, credentialUid uuid.UUID) string {
	return fmt.Sprintf("users/%s/%s", userId, credentialUid.String())
}

func CreateCredential(log *logrus.Logger, sess *session.Session, bucket, s3ObjectKey string, v interface{}) error {
	log.WithFields(logrus.Fields{"bucket": bucket, "objectKey": s3ObjectKey}).Debug("s3 create")
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}

	svc := s3.New(sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String("" + s3ObjectKey),
		Body:                 bytes.NewReader(buf),
		ServerSideEncryption: aws.String(s3.ServerSideEncryptionAes256),
	})

	return err
}

func List(ctx context.Context, log *logrus.Logger, sess *session.Session, bucket, prefix string) (*s3.ListObjectsOutput, error) {
	log.WithFields(logrus.Fields{"bucket": bucket, "objectPrefix": prefix}).Debug("s3 List")
	svc := s3.New(sess)
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	result, err := svc.ListObjectsWithContext(ctx, input)
	//if aerr, ok := err.(awserr.Error); ok {
	//	fmt.Println(aerr.Code())
	//}

	if err != nil {
		return &s3.ListObjectsOutput{}, fmt.Errorf("error listing credentials %w", err)
	}

	for i := range result.Contents {
		log.WithFields(logrus.Fields{"objectKey": *result.Contents[i].Key}).Debug("found object in s3 list")
	}
	return result, nil
}

func Get(log *logrus.Logger, sess *session.Session, bucket, s3ObjectKey string) ([]byte, error) {
	log.WithFields(logrus.Fields{"bucket": bucket, "objectKey": s3ObjectKey}).Debug("s3 get")
	svc := s3.New(sess)
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3ObjectKey),
	})

	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func GetCredential(log *logrus.Logger, sess *session.Session, bucket, s3ObjectKey string, v interface{}) error {
	// TODO: refactor err on return to break out awserr fields
	resp, err := Get(log, sess, bucket, s3ObjectKey)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, v)
}

func GetCredentials(ctx context.Context, log *logrus.Logger, sess *session.Session, bucket, s3ObjectKey string, v interface{}) error {
	objects, err := List(ctx, log, sess, bucket, s3ObjectKey)
	if err != nil {
		return fmt.Errorf("error listing credentials list %w", err)
	}

	var objectBlobs []json.RawMessage
	for i := range objects.Contents {
		objectBlob, err := Get(log, sess, bucket, *objects.Contents[i].Key)
		if err != nil {
			return fmt.Errorf("error getting credential in list %w", err)
		}

		objectBlobs = append(objectBlobs, objectBlob)
	}

	jsonBlobs, err := json.Marshal(objectBlobs)
	if err != nil {
		return fmt.Errorf("error marshalling credential list %w", err)
	}

	return json.Unmarshal(jsonBlobs, v)
}

func DeleteCredential(log *logrus.Logger, sess *session.Session, bucket, s3ObjectKey string) error {
	log.WithFields(logrus.Fields{"bucket": bucket, "objectKey": s3ObjectKey}).Debug("s3 delete")
	svc := s3.New(sess)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3ObjectKey),
	}

	_, err := svc.DeleteObject(input)
	return err
}

func DeleteAll(ctx context.Context, log *logrus.Logger, sess *session.Session, bucket, s3ObjectKey string) error {
	objects, err := List(ctx, log, sess, bucket, s3ObjectKey)
	if err != nil {
		return err
	}

	for i := range objects.Contents {
		if err := DeleteCredential(log, sess, bucket, *objects.Contents[i].Key); err != nil {
			return err
		}
	}

	return nil
}
