package api

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
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
	})
}

func (c *Credentials) getCredential(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		subendpoints.CredentialIdFromParams(w, r, params, func(credentialUid uuid.UUID) {
			var data models.Credential
			objectKey := s3.GetKeyForSingleCredential(userId, credentialUid)

			// Check s3 for credential
			err := s3.GetCredential(c.log, c.sess, c.bucket, objectKey, &data)
			if err != nil {
				intake.RespondError(w, r, err, http.StatusBadRequest)
				return
			}
			intake.RespondJSON(w, r, http.StatusOK, data)
		})
	})
}

func (c *Credentials) getCredentials(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	subendpoints.UserIdFromClaims(w, r, func(userId string) {
		objectKey := s3.GetKeyForAllCredentials(userId)
		var ts []models.Credential

		// Check S3 for the credentials
		if err := s3.GetCredentials(r.Context(), c.log, c.sess, c.bucket, objectKey, &ts); err != nil {
			if err.Error() == NOT_FOUND {
				intake.Respond(w, r, http.StatusNoContent, nil)
				return
			}
			intake.RespondError(w, r, err, http.StatusInternalServerError)
			return
		}

		intake.RespondJSON(w, r, http.StatusOK, ts)
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
		})
	})
}
