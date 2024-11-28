package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

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
	dbQueries := database.New(db)
	dbQueries.CreateUser(context.TODO(), sql.NullString{"romus@duck.com", true})
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

	mux.HandleFunc("POST /api/users", handlerChirpsValidate)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix(
			"/app", http.FileServer(
				http.Dir("public")))))

	log.Fatal(s.ListenAndServe())
}
