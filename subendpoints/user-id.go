package subendpoints

import (
	"net/http"

	"github.com/bf-dbubel/intake"
)

func UserIdFromClaims(w http.ResponseWriter, r *http.Request, next func(userId string)) {
	use, ok := r.Context().Value("userId").(string)
	if !ok {
		intake.RespondJSON(w, r, http.StatusBadRequest, map[string]string{"error": "no userId present in claims"})
		return
	}
	next(use)
}
