package main

import (
	"log"
	"net/http"

	"github.com/Xuanwo/gitqus/api"
)

func main() {
	err := http.ListenAndServe(":8080", http.HandlerFunc(api.Handler))
	if err != nil {
		log.Fatal(err)
	}
}
