package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/romusking/chirpy/internal/database"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error connecting to database: %s", err)
	}
	platform := os.Getenv("PLATFORM")

	secret := os.Getenv("SECRET")

	mux := http.NewServeMux()
	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		MaxHeaderBytes: 1 << 20,
	}
	apiCfg := apiConfig{
		db:       database.New(db),
		platform: platform,
		secret:   secret,
	}

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("GET /admin/metrics", apiCfg.middlewareMetricsDsp)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetUserDB)

	mux.HandleFunc("POST /api/reset", apiCfg.middlewareMetricsRst)

	mux.HandleFunc("POST /api/users", apiCfg.createUser)

	mux.HandleFunc("POST /api/login", apiCfg.loginUser)

	mux.HandleFunc("POST /api/refresh", apiCfg.refreshToken)

	mux.HandleFunc("POST /api/revoke", apiCfg.revokeRefreshToken)

	mux.HandleFunc("POST /api/chirps", apiCfg.createChirp)

	mux.HandleFunc("GET /api/chirps", apiCfg.getAllChirps)

	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getOneChirp)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix(
			"/app", http.FileServer(
				http.Dir("public")))))

	log.Fatal(s.ListenAndServe())
}
