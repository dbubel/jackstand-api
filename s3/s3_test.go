package s3

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

var sess *session.Session
var log *logrus.Logger

func init() {
	bs := true
	testEndpoint := "http://localhost:5002"
	sess, _ = session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         &testEndpoint,
		S3ForcePathStyle: &bs,
	})
	log = logrus.New()
	log.SetLevel(logrus.FatalLevel)

}

func TestS3(t *testing.T) {
	type testStruct struct {
		Name    string
		Address string
	}

	ts := testStruct{
		Name:    "Jon",
		Address: "Home",
	}

	ts2 := testStruct{
		Name:    "zach",
		Address: "space",
	}
	_, _ = ts, ts2

	t.Run("test upload object to s3", func(t *testing.T) {
		err := CreateCredential(log, sess, "jackstand-s3-test", "testStruct.json", ts)
		assert.NoError(t, err)
	})

	t.Run("test listing objects from s3", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		objects, err := List(ctx, log, sess, "jackstand-s3-test", "testStruct.json")
		assert.NoError(t, err)
		assert.Len(t, objects.Contents, 1)
	})

	t.Run("test get object from s3", func(t *testing.T) {
		resp, err := Get(log, sess, "jackstand-s3-test", "testStruct.json")
		_ = resp
		assert.NoError(t, err)
		assert.JSONEq(t, `{"Name":"Jon","Address":"Home"}`, string(resp))
	})

	t.Run("test get object from s3 does not exist", func(t *testing.T) {
		resp, err := Get(log, sess, "jackstand-s3-test", "dne.json")
		assert.Error(t, err)
		assert.Equal(t, []byte{}, resp)
	})

	t.Run("test get marshalled object from s3", func(t *testing.T) {
		var m testStruct
		err := GetCredential(log, sess, "jackstand-s3-test", "testStruct.json", &m)
		assert.NoError(t, err)
		assert.Equal(t, m.Name, ts.Name)
		assert.Equal(t, m.Address, ts.Address)
	})

	t.Run("test get marshalled object slice from s3", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		var m []testStruct
		err := CreateCredential(log, sess, "jackstand-s3-test", "testStruct.json", ts)
		assert.NoError(t, err)
		err = CreateCredential(log, sess, "jackstand-s3-test", "testStruct2.json", ts2)
		assert.NoError(t, err)
		err = GetCredentials(ctx, log, sess, "jackstand-s3-test", "testStruct", &m)
		assert.NoError(t, err)
		assert.Contains(t, m, ts)
		assert.Contains(t, m, ts2)
	})

	t.Run("test remove object from s3", func(t *testing.T) {
		err := DeleteCredential(log, sess, "jackstand-s3-test", "testStruct.json")
		assert.NoError(t, err)
	})

	t.Run("nuke bucket", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		DeleteAll(ctx, log, sess, "jackstand-s3-test", "users/")
	})
}
