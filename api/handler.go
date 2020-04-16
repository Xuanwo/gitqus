package api

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
