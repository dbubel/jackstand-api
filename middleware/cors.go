package middleware

import (
	"github.com/dbubel/intake"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// Handle CORS for other requests
func Cors(next intake.Handler) intake.Handler {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		next(w, r, params)
	}
}
