package subendpoints

import (
	"net/http"

	"github.com/bf-dbubel/intake"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

func CredentialIdFromParams(w http.ResponseWriter, r *http.Request, params httprouter.Params, next func(credentialUid uuid.UUID)) {
	uid, err := uuid.FromString(params.ByName("credentialUid"))
	if err != nil {
		intake.RespondError(w, r, err, http.StatusBadRequest)
		return
	}
	next(uid)
}
