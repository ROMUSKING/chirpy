package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
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

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", apiCfg.middlewareMetricsDsp)

	mux.HandleFunc("POST /admin/reset", apiCfg.middlewareMetricsRst)

	mux.HandleFunc("POST /api/validate_chirp", validator)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix(
			"/app", http.FileServer(
				http.Dir("public")))))

	log.Fatal(s.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsDsp(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))

}

func (cfg *apiConfig) middlewareMetricsRst(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Counter reset to zero"))

}

func validator(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type errorVal struct {
		Error string `json:"error"`
	}
	type returnVals struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		Valid bool `json:"valid"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {

		respBody := errorVal{
			Error: fmt.Sprintf("Error decoding parameters: %s", err),
		}
		dat, err := json.Marshal(respBody)
		if err != nil {

			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		w.Write(dat)
		return
	}

	if len(params.Body) > 140 {

		respBody := errorVal{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {

			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Chirp is too long")
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	respBody := returnVals{
		Valid: true,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}
