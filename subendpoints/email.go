package subendpoints

import (
	"net/http"

	"github.com/bf-dbubel/intake"
)

func EmailFromClaims(w http.ResponseWriter, r *http.Request, next func(email string)) {
	email, ok := r.Context().Value("email").(string)
	if !ok {
		intake.RespondJSON(w, r, http.StatusBadRequest, map[string]string{"error": "no email present in claims"})
		return
	}
	next(email)
}
