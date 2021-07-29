package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/dbubel/intake"
	"github.com/julienschmidt/httprouter"
)

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

var BuildTime = time.Now().Format(time.RFC1123)

func (c *Credentials) status(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	intake.RespondJSON(w, r, http.StatusOK, struct {
		Alloc      uint64
		TotalAlloc uint64
		CacheItems int
		BuildTime  string
	}{
		Alloc:      bToMb(m.Alloc),
		TotalAlloc: bToMb(m.TotalAlloc),
		//CacheItems: c.cache.NumElements(),
		BuildTime: BuildTime,
	})
}
