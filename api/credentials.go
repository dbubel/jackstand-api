package api

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	//cacher "github.com/dbubel/cacheflow"
	"github.com/dbubel/intake"
	"github.com/dbubel/jackstand-api/models"
	"github.com/dbubel/jackstand-api/s3"
	"github.com/dbubel/jackstand-api/subendpoints"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type Credentials struct {
	bucket string
	sess   *session.Session
	log    *logrus.Logger
	//cache  *cacher.Cacher
}

const NOT_FOUND = "error listing credentials list no results found"

func (c *Credentials) updateUsername(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		subendpoints.CredentialIdFromParams(w, r, params, func(credentialUid uuid.UUID) {
			attribute := struct {
				Username string `validate:"required"`
			}{}

			if err := intake.UnmarshalJSON(r.Body, &attribute); err != nil {
				intake.RespondError(w, r, err, http.StatusBadRequest)
				return
			}

			var existingCredential models.Credential
			objectKey := s3.GetKeyForSingleCredential(userId, credentialUid)

			if err := s3.GetCredential(c.log, c.sess, c.bucket, objectKey, &existingCredential); err != nil {
				intake.RespondError(w, r, err, http.StatusBadRequest)
				return
			}

			existingCredential.Username = attribute.Username
			existingCredential.UpdatedAt = models.CustomTime(time.Now())

			if err := s3.CreateCredential(c.log, c.sess, c.bucket, objectKey, existingCredential); err != nil {
				intake.RespondError(w, r, err, http.StatusInternalServerError)
				return
			}

			intake.RespondJSON(w, r, http.StatusOK, existingCredential)
			//c.cache.Delete(s3.GetKeyForAllCredentials(userId))
			//c.cache.Delete(objectKey)
		})
	})
}

func (c *Credentials) updatePassword(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		subendpoints.CredentialIdFromParams(w, r, params, func(credentialUid uuid.UUID) {
			attribute := struct {
				Password string `validate:"required"`
			}{}

			if err := intake.UnmarshalJSON(r.Body, &attribute); err != nil {
				intake.RespondError(w, r, err, http.StatusBadRequest)
				return
			}

			var existingCredential models.Credential
			objectKey := s3.GetKeyForSingleCredential(userId, credentialUid)

			if err := s3.GetCredential(c.log, c.sess, c.bucket, objectKey, &existingCredential); err != nil {
				intake.RespondError(w, r, err, http.StatusBadRequest)
				return
			}

			existingCredential.Password = attribute.Password
			existingCredential.UpdatedAt = models.CustomTime(time.Now())

			if err := s3.CreateCredential(c.log, c.sess, c.bucket, objectKey, existingCredential); err != nil {
				intake.RespondError(w, r, err, http.StatusInternalServerError)
				return
			}

			intake.RespondJSON(w, r, http.StatusOK, existingCredential)
			//c.cache.Delete(s3.GetKeyForAllCredentials(userId))
			//c.cache.Delete(objectKey)
		})
	})
}

func (c *Credentials) updateServiceName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		subendpoints.CredentialIdFromParams(w, r, params, func(credentialUid uuid.UUID) {
			attribute := struct {
				Service string `validate:"required,max=15,min=3"`
			}{}

			if err := intake.UnmarshalJSON(r.Body, &attribute); err != nil {
				intake.RespondError(w, r, err, http.StatusBadRequest, "service name is too long")
				return
			}

			var existingCredential models.Credential
			objectKey := s3.GetKeyForSingleCredential(userId, credentialUid)

			if err := s3.GetCredential(c.log, c.sess, c.bucket, objectKey, &existingCredential); err != nil {
				intake.RespondError(w, r, err, http.StatusBadRequest)
				return
			}

			existingCredential.Service = attribute.Service
			existingCredential.UpdatedAt = models.CustomTime(time.Now())

			if err := s3.CreateCredential(c.log, c.sess, c.bucket, objectKey, existingCredential); err != nil {
				intake.RespondError(w, r, err, http.StatusInternalServerError)
				return
			}

			intake.RespondJSON(w, r, http.StatusOK, existingCredential)
			//c.cache.Delete(s3.GetKeyForAllCredentials(userId))
			//c.cache.Delete(objectKey)
		})
	})
}

func (c *Credentials) createCredential(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		var credential models.Credential

		if err := intake.UnmarshalJSON(r.Body, &credential); err != nil {
			intake.RespondError(w, r, err, http.StatusBadRequest)
			return
		}

		credential.Uid = uuid.Must(uuid.NewV4())
		credential.CreatedAt = models.CustomTime(time.Now())
		credential.UpdatedAt = models.CustomTime(time.Now())
		objectKey := s3.GetKeyForSingleCredential(userId, credential.Uid)

		if err := s3.CreateCredential(c.log, c.sess, c.bucket, objectKey, credential); err != nil {
			intake.RespondError(w, r, err, http.StatusInternalServerError)
			return
		}

		intake.RespondJSON(w, r, http.StatusOK, credential)

		// insert a new object in the cache for this specific credential
		//err := c.cache.InsertObject(objectKey, credential)

		//if err != nil {
		//	c.log.WithError(err).Warn("error caching struct")
		//}

		// Blow the cache away for the all credentials object
		// so it will not contain the new object on re cache
		//c.cache.Delete(s3.GetKeyForAllCredentials(userId))
	})
}

func (c *Credentials) getCredential(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		subendpoints.CredentialIdFromParams(w, r, params, func(credentialUid uuid.UUID) {
			var data models.Credential
			objectKey := s3.GetKeyForSingleCredential(userId, credentialUid)

			// Check the cache for existing value
			//err := c.cache.GetObject(objectKey, &data)
			//if err == nil { // If no error then we got a cache hit
			//	intake.RespondJSON(w, r, http.StatusOK, data)
			//	return
			//}

			// Check s3 for credential
			err := s3.GetCredential(c.log, c.sess, c.bucket, objectKey, &data)
			if err != nil {
				intake.RespondError(w, r, err, http.StatusBadRequest)
				return
			}
			intake.RespondJSON(w, r, http.StatusOK, data)
			//c.cache.InsertObject(objectKey, data)
		})
	})
}

func (c *Credentials) getCredentials(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		duration := 5000 * time.Millisecond

		// Create a context that is both manually cancellable and will signal
		// cancel at the specified duration.
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()
		objectKey := s3.GetKeyForAllCredentials(userId)
		var ts []models.Credential

		// Check the cache for existing value
		//if err := c.cache.GetObject(objectKey, &ts); err == nil {
		//	intake.RespondJSON(w, r, http.StatusOK, ts)
		//	return
		//}

		// Check S3 for the credentials
		if err := s3.GetCredentials(ctx, c.log, c.sess, c.bucket, objectKey, &ts); err != nil {
			if err.Error() == NOT_FOUND {
				intake.Respond(w, r, http.StatusNoContent, nil)
				return
			}
			intake.RespondError(w, r, err, http.StatusInternalServerError)
			return
		}

		intake.RespondJSON(w, r, http.StatusOK, ts)
		//c.cache.InsertObject(objectKey, ts)
	})
}

func (c *Credentials) deleteCredential(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		subendpoints.CredentialIdFromParams(w, r, params, func(credentialUid uuid.UUID) {
			objectKey := s3.GetKeyForSingleCredential(userId, credentialUid)

			if err := s3.DeleteCredential(c.log, c.sess, c.bucket, objectKey); err != nil {
				intake.RespondError(w, r, err, http.StatusInternalServerError)
				return
			}

			intake.RespondJSON(w, r, http.StatusOK, map[string]string{
				"status":      "deleted",
				"description": "credential deleted OK",
			})
			//c.cache.Delete(objectKey)
		})
	})
}
