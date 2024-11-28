package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	apiCfg := apiConfig{}

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("GET /admin/metrics", apiCfg.middlewareMetricsDsp)

	mux.HandleFunc("POST /admin/reset", apiCfg.middlewareMetricsRst)

	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix(
			"/app", http.FileServer(
				http.Dir("public")))))

	log.Fatal(s.ListenAndServe())
}
