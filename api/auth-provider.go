package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dbubel/intake"
	"github.com/julienschmidt/httprouter"
)

// Todo routes
type Create struct {
	Email             string `json:"email" validate:"required"`
	Password          string `json:"password" validate:"required"`
	ReturnSecureToken bool   `json:"returnSecureToken" validate:"required"`
}

type Delete struct {
	IDToken string `json:"idToken" validate:"required"`
}

type UpdatePassword struct {
	IDToken           string `json:"idToken" validate:"required"`
	Password          string `json:"password" validate:"required"`
	ReturnSecureToken bool   `json:"returnSecureToken" validate:"required"`
}

type Verify struct {
	RequestType string `json:"requestType" validate:"required"`
	IDToken     string `json:"idToken" validate:"required"`
}

type singin struct {
	Email             string `json:"email" validate:"required"`
	Password          string `json:"password" validate:"required"`
	ReturnSecureToken bool   `json:"returnSecureToken" validate:"required"`
}

type FireBaseAuth struct {
	ApiKey          string
	FirebaseBaseURL string
}

func (c *FireBaseAuth) Signin(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	singinURL := fmt.Sprintf("%s/verifyPassword?key=%s", c.FirebaseBaseURL, c.ApiKey)
	var signinReq singin
	if err := intake.UnmarshalJSON(r.Body, &signinReq); err != nil {
		intake.RespondError(w, r, err, http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	signinJSON, err := json.Marshal(signinReq)
	if err != nil {
		intake.RespondError(w, r, err, http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest("POST", singinURL, bytes.NewReader(signinJSON))
	if err != nil {
		intake.RespondError(w, r, err, http.StatusBadRequest)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		intake.RespondError(w, r, err, http.StatusBadRequest)
		return
	}

	var firebaseResp interface{}
	err = json.NewDecoder(res.Body).Decode(&firebaseResp)
	if err != nil {
		intake.RespondError(w, r, err, http.StatusBadRequest)
		return
	}
	intake.RespondJSON(w, r, res.StatusCode, firebaseResp)

}
