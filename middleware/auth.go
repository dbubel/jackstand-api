package middleware

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bf-dbubel/intake"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

const (
	publicKeyUrl string = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
)

var publicCerts map[string]string
var keyDownloadedAt time.Time

func init() {
	getKey()
}

func getKey() error {
	resp, err := http.Get(publicKeyUrl)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &publicCerts)
	if err != nil {
		return err
	}
	keyDownloadedAt = time.Now()
	return nil
}

// AuthHandler validates a JWT present in the request.
func Auth(next intake.Handler) intake.Handler {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		var err error
		if !keyDownloadedAt.After(time.Now().Add(time.Hour * -1)) {
			getKey()
		}

		token := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(token) < 2 {
			intake.Respond(w, r, http.StatusUnauthorized, []byte("unauthorized"))
			return
		}

		t := strings.Replace(token[1], " ", "", -1) // replace white space

		var tok *jwt.Token
		for _, cert := range publicCerts {
			tok, err = jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
				return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			})

			if err != nil {
				continue
			} else {
				break
			}
		}

		// No valid jwt was found
		if err != nil {
			intake.RespondError(w, r, err, http.StatusUnauthorized, "invalid token")
			//intake.Respond(w, r, http.StatusUnauthorized, []byte("unauthorized no jwt found"))
			return
		}

		if !tok.Valid {
			//intake.RespondError(w, r, err, http.StatusUnauthorized, "invalid token")
			intake.Respond(w, r, http.StatusUnauthorized, []byte("unauthorized invalid token"))
			return
		}

		email, ok := tok.Claims.(jwt.MapClaims)["email"].(string)
		if !ok {
			intake.Respond(w, r, http.StatusUnauthorized, []byte("unauthorized no email in claims"))
			return
		}

		//emailVerified, ok := tok.Claims.(jwt.MapClaims)["email_verified"].(bool)
		//if emailVerified != true || !ok {
		//	intake.RespondError(w, r, fmt.Errorf("email not verified"),http.StatusUnauthorized)
		//	return
		//}

		//
		//iss, ok := tok.Claims.(jwt.MapClaims)["iss"].(string)
		//if iss != JWT_ISSUER || !ok {
		//	return errors.New("Invalid ISS")
		//}
		//
		//aud, ok := tok.Claims.(jwt.MapClaims)["aud"].(string)
		//if aud != JWT_AUD || !ok {
		//	return errors.New("Invalid AUD")
		//}
		//

		localId, ok := tok.Claims.(jwt.MapClaims)["user_id"].(string)
		if !ok {
			intake.Respond(w, r, http.StatusUnauthorized, []byte("unauthorized no user in claims"))
			return
		}

		ctx := context.WithValue(r.Context(), "userId", localId)
		ctx = context.WithValue(ctx, "email", email)
		next(w, r.WithContext(ctx), params)
	}
}
