package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	cacher "github.com/dbubel/cacheflow"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/brianvoe/gofakeit"
	"github.com/dbubel/intake"
	"github.com/dbubel/models"
	"github.com/dbubel/s3"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var sess *session.Session
var log *logrus.Logger
var testBucket string = "jackstand-s3-test"

func init() {
	gofakeit.Seed(rand.Int63())
	bs := true
	testEndpoint := "http://localhost:5002"
	sess, _ = session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         &testEndpoint,
		S3ForcePathStyle: &bs,
	})
	log = logrus.New()
	log.SetLevel(logrus.ErrorLevel)
}
func randomCredential() models.Credential {
	return models.Credential{
		Uid:         uuid.Must(uuid.NewV4()),
		Service:     gofakeit.HackerVerb(),
		Username:    gofakeit.BeerAlcohol(),
		Password:    gofakeit.JobLevel(),
		Description: gofakeit.CarMaker(),
		Metadata:    nil,
	}
}

func TestGetCredential(t *testing.T) {
	testCredential1 := randomCredential()
	testCredential2 := randomCredential()

	app := intake.New(log)
	credsApi := Credentials{
		bucket: testBucket,
		sess:   sess,
		log:    log,
		cache:  cacher.NewCacherDefault(),
	}

	credentialsSlice := append([]models.Credential{}, testCredential1, testCredential2)
	app.AddEndpoints(endpoints(credsApi))
	userIdFromClaims := gofakeit.Username()
	userIdFromClaims2 := gofakeit.Username()
	_ = userIdFromClaims2
	objectKey := s3.GetKeyForSingleCredential(userIdFromClaims, testCredential1.Uid)
	err := s3.CreateCredential(log, sess, credsApi.bucket, objectKey, testCredential1)
	assert.NoError(t, err)

	objectKey = s3.GetKeyForSingleCredential(userIdFromClaims, testCredential2.Uid)
	err = s3.CreateCredential(log, sess, credsApi.bucket, objectKey, testCredential2)
	assert.NoError(t, err)

	t.Run("test getting a specific credential for a user", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/users/credentials/"+testCredential1.Uid.String(), nil)
		ctx := context.WithValue(r.Context(), "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var c models.Credential
		err := json.Unmarshal(body, &c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testCredential1, c)
	})

	t.Run("test getting all credentials for a user", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/users/credentials", nil)
		ctx := context.WithValue(r.Context(), "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var c []models.Credential
		err := json.Unmarshal(body, &c)
		assert.NoError(t, err)
		assert.Len(t, c, 2)
		assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("resp:%s", string(body)))
		assert.Contains(t, credentialsSlice, c[0])
		//assert.Contains(t, credentialsSlice, c[1])
	})

	t.Run("test getting all credentials for a user", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/users/credentials", nil)
		ctx := context.WithValue(r.Context(), "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var c []models.Credential
		err := json.Unmarshal(body, &c)
		assert.NoError(t, err)
		assert.Len(t, c, 2)
		assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("resp:%s", string(body)))
		assert.Contains(t, credentialsSlice, c[0])
		assert.Contains(t, credentialsSlice, c[1])
	})

	t.Run("test get credentialId does not exist", func(t *testing.T) {
		u := uuid.Must(uuid.NewV4())
		r := httptest.NewRequest(http.MethodGet, "/users/credentials/"+u.String(), nil)
		ctx := context.WithValue(r.Context(), "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var c models.Credential
		err := json.Unmarshal(body, &c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("test get credential does not belong to user", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/users/credentials/"+testCredential1.Uid.String(), nil)
		ctx := context.WithValue(r.Context(), "userId", userIdFromClaims2)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("test get non uuid credentialId", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/users/credentials/notAUUID", nil)
		ctx := context.WithValue(r.Context(), "userId", userIdFromClaims2)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var c models.Credential
		err := json.Unmarshal(body, &c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, ` {"description":null,"error":"uuid: incorrect UUID length 8 in string \"notAUUID\""}`, string(body))
	})

	t.Run("nuke bucket", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		err := s3.DeleteAll(ctx, log, sess, testBucket, "users/")
		assert.NoError(t, err)
	})
}

func TestCreateCredential(t *testing.T) {
	testCredential1 := randomCredential()
	testCredential2 := randomCredential()
	_, _ = testCredential1, testCredential2
	userIdFromClaims := gofakeit.Username()
	userIdFromClaims2 := gofakeit.Username()
	_, _ = userIdFromClaims, userIdFromClaims2

	app := intake.New(log)
	credsApi := Credentials{
		bucket: testBucket,
		sess:   sess,
		log:    log,
		cache:  cacher.NewCacherDefault(),
	}
	app.AddEndpoints(endpoints(credsApi))

	t.Run("test creating a new credential for a user", func(t *testing.T) {
		requestBody, _ := json.Marshal(testCredential1)
		r := httptest.NewRequest(http.MethodPost, "/users/credentials", bytes.NewReader(requestBody))
		ctx := context.WithValue(r.Context(), "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var c models.Credential
		err := json.Unmarshal(body, &c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		objectKey := s3.GetKeyForSingleCredential(userIdFromClaims, c.Uid)
		var objectFroms3 models.Credential
		err = s3.GetCredential(log, sess, testBucket, objectKey, &objectFroms3)
		assert.NoError(t, err)
		assert.Equal(t, testCredential1.Username, c.Username)
		assert.Equal(t, testCredential1.Username, objectFroms3.Username)

		assert.Equal(t, testCredential1.Password, c.Password)
		assert.Equal(t, testCredential1.Password, objectFroms3.Password)

		assert.Equal(t, testCredential1.Service, c.Service)
		assert.Equal(t, testCredential1.Service, objectFroms3.Service)
	})

	t.Run("test create credential missing username", func(t *testing.T) {
		// check how many objects exist
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		countOld, err := s3.List(ctx, log, sess, testBucket, "users/")
		assert.NoError(t, err)
		requestBody := []byte(` { "Password": "coffee", "Service":"github" }`)
		r := httptest.NewRequest(http.MethodPost, "/users/credentials", bytes.NewReader(requestBody))
		ctx = context.WithValue(ctx, "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t,
			`{"description":null,"error":",{Username:required}"}`,
			string(body),
		)

		countNew, err := s3.List(ctx, log, sess, testBucket, "users/")
		// ensure no new objects were created
		assert.Equal(t, len(countOld.Contents), len(countNew.Contents))
	})

	t.Run("test create credential missing password", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		// check how many objects exist
		countOld, err := s3.List(ctx, log, sess, testBucket, "users/")
		assert.NoError(t, err)
		requestBody := []byte(` { "Username": "coffee", "Service":"github" }`)
		r := httptest.NewRequest(http.MethodPost, "/users/credentials", bytes.NewReader(requestBody))
		ctx = context.WithValue(ctx, "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t,
			`{"description":null,"error":",{Password:required}"}`,
			string(body),
		)

		countNew, err := s3.List(ctx, log, sess, testBucket, "users/")
		// ensure no new objects were created
		assert.Equal(t, len(countOld.Contents), len(countNew.Contents))
	})

	t.Run("test create credential missing service", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		// check how many objects exist
		countOld, err := s3.List(ctx, log, sess, testBucket, "users/")
		assert.NoError(t, err)
		requestBody := []byte(` { "Username": "coffee", "Password":"github" }`)
		r := httptest.NewRequest(http.MethodPost, "/users/credentials", bytes.NewReader(requestBody))
		ctx = context.WithValue(ctx, "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t,
			`{"description":null,"error":",{Service:required}"}`,
			string(body),
		)

		countNew, err := s3.List(ctx, log, sess, testBucket, "users/")
		// ensure no new objects were created
		assert.Equal(t, len(countOld.Contents), len(countNew.Contents))
	})

	t.Run("test create credential bad json", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		// check how many objects exist
		countOld, err := s3.List(ctx, log, sess, testBucket, "users/")
		assert.NoError(t, err)
		requestBody := []byte(` { "Username": "coff`)
		r := httptest.NewRequest(http.MethodPost, "/users/credentials", bytes.NewReader(requestBody))
		ctx = context.WithValue(ctx, "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t,
			`{"description":null,"error":"unexpected EOF"}`,
			string(body),
		)

		countNew, err := s3.List(ctx, log, sess, testBucket, "users/")
		// ensure no new objects were created
		assert.Equal(t, len(countOld.Contents), len(countNew.Contents))
	})

	t.Run("nuke bucket", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		err := s3.DeleteAll(ctx, log, sess, testBucket, "users/")
		assert.NoError(t, err)
	})
}

func TestUpdateCredential(t *testing.T) {
	testCredential1 := randomCredential()
	testCredential2 := randomCredential()
	_, _ = testCredential1, testCredential2
	userIdFromClaims := gofakeit.Username()
	userIdFromClaims2 := gofakeit.Username()
	_, _ = userIdFromClaims, userIdFromClaims2

	app := intake.New(log)
	credsApi := Credentials{
		bucket: testBucket,
		sess:   sess,
		log:    log,
		cache:  cacher.NewCacherDefault(),
	}

	app.AddEndpoints(endpoints(credsApi))

	var credentialToModify models.Credential
	s3.CreateCredential(log, sess, testBucket, s3.GetKeyForSingleCredential(userIdFromClaims, testCredential1.Uid), testCredential1)
	s3.GetCredential(log, sess, testBucket, s3.GetKeyForSingleCredential(userIdFromClaims, testCredential1.Uid), &credentialToModify)

	t.Run("test updating username for a credential", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		countOld, err := s3.List(ctx, log, sess, testBucket, "users/")
		requestBody := []byte(` { "Username": "coffee" }`)
		r := httptest.NewRequest(http.MethodPut, "/users/credentials/"+credentialToModify.Uid.String()+"/username", bytes.NewReader(requestBody))
		ctx = context.WithValue(ctx, "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var recievedCred models.Credential
		err = json.Unmarshal(body, &recievedCred)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var objectFroms3 models.Credential
		err = s3.GetCredential(log, sess, testBucket, s3.GetKeyForSingleCredential(userIdFromClaims, recievedCred.Uid), &objectFroms3)
		assert.NoError(t, err)
		assert.NotEqual(t, testCredential1.Username, recievedCred.Username)
		assert.NotEqual(t, testCredential1.Username, objectFroms3.Username)

		assert.Equal(t, recievedCred.Username, "coffee")
		assert.Equal(t, objectFroms3.Username, "coffee")

		assert.Equal(t, testCredential1.Password, recievedCred.Password)
		assert.Equal(t, testCredential1.Password, objectFroms3.Password)

		assert.Equal(t, testCredential1.Service, recievedCred.Service)
		assert.Equal(t, testCredential1.Service, objectFroms3.Service)

		countNew, err := s3.List(ctx, log, sess, testBucket, "users/")
		// ensure no new objects were created
		assert.Equal(t, len(countOld.Contents), len(countNew.Contents))
	})

	t.Run("test updating password for a credential", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		countOld, err := s3.List(ctx, log, sess, testBucket, "users/")
		requestBody := []byte(` { "Password": "passwordisweak" }`)
		r := httptest.NewRequest(http.MethodPut, "/users/credentials/"+credentialToModify.Uid.String()+"/password", bytes.NewReader(requestBody))
		ctx = context.WithValue(r.Context(), "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var recievedCred models.Credential
		err = json.Unmarshal(body, &recievedCred)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var objectFroms3 models.Credential
		err = s3.GetCredential(log, sess, testBucket, s3.GetKeyForSingleCredential(userIdFromClaims, recievedCred.Uid), &objectFroms3)
		assert.NoError(t, err)

		assert.Equal(t, recievedCred.Password, "passwordisweak")
		assert.Equal(t, objectFroms3.Password, "passwordisweak")

		assert.Equal(t, recievedCred.Username, "coffee")
		assert.Equal(t, objectFroms3.Username, "coffee")

		assert.Equal(t, testCredential1.Service, recievedCred.Service)
		assert.Equal(t, testCredential1.Service, objectFroms3.Service)

		countNew, err := s3.List(ctx, log, sess, testBucket, "users/")
		// ensure no new objects were created
		assert.Equal(t, len(countOld.Contents), len(countNew.Contents))
	})

	t.Run("test updating service for a credential", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		countOld, err := s3.List(ctx, log, sess, testBucket, "users/")
		requestBody := []byte(` { "Service": "serviceisweak" }`)
		r := httptest.NewRequest(http.MethodPut, "/users/credentials/"+credentialToModify.Uid.String()+"/service", bytes.NewReader(requestBody))
		ctx = context.WithValue(ctx, "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		var recievedCred models.Credential
		err = json.Unmarshal(body, &recievedCred)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var objectFroms3 models.Credential
		err = s3.GetCredential(log, sess, testBucket, s3.GetKeyForSingleCredential(userIdFromClaims, recievedCred.Uid), &objectFroms3)
		assert.NoError(t, err)

		assert.Equal(t, recievedCred.Password, "passwordisweak")
		assert.Equal(t, objectFroms3.Password, "passwordisweak")

		assert.Equal(t, recievedCred.Username, "coffee")
		assert.Equal(t, objectFroms3.Username, "coffee")

		assert.Equal(t, recievedCred.Service, "serviceisweak")
		assert.Equal(t, objectFroms3.Service, "serviceisweak")

		countNew, err := s3.List(ctx, log, sess, testBucket, "users/")
		// ensure no new objects were created
		assert.Equal(t, len(countOld.Contents), len(countNew.Contents))
	})

	t.Run("test updating service for a credential that is too long", func(t *testing.T) {
		//countOld, err := s3.List(log, sess, testBucket, "users/")
		requestBody := []byte(` { "Service": "serviceisweakand is way too long" }`)
		r := httptest.NewRequest(http.MethodPut, "/users/credentials/"+credentialToModify.Uid.String()+"/service", bytes.NewReader(requestBody))
		ctx := context.WithValue(r.Context(), "userId", userIdFromClaims)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, r.WithContext(ctx))
		body, _ := ioutil.ReadAll(w.Body)
		//var recievedCred models.Credential
		//err := json.Unmarshal(body, &recievedCred)
		//assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code, string(body))
		//
		//var objectFroms3 models.Credential
		//err = s3.GetCredential(log, sess, testBucket, s3.GetKeyForSingleCredential(userIdFromClaims, recievedCred.Uid), &objectFroms3)
		//assert.NoError(t, err)
		//
		//assert.Equal(t, recievedCred.Password, "passwordisweak")
		//assert.Equal(t, objectFroms3.Password, "passwordisweak")
		//
		//assert.Equal(t, recievedCred.Username, "coffee")
		//assert.Equal(t, objectFroms3.Username, "coffee")
		//
		//assert.Equal(t, recievedCred.Service, "serviceisweak")
		//assert.Equal(t, objectFroms3.Service, "serviceisweak")
		//
		//countNew, err := s3.List(log, sess, testBucket, "users/")
		//// ensure no new objects were created
		//assert.Equal(t, len(countOld.Contents), len(countNew.Contents))
	})

	t.Run("nuke bucket", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4000*time.Millisecond)
		defer cancel()
		err := s3.DeleteAll(ctx, log, sess, testBucket, "users/")
		assert.NoError(t, err)
	})

}
