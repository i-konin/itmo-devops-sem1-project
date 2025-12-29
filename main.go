package main

import (
	"log"
	"net/http"
	"os"

	"project_sem/internal/api"
	"project_sem/internal/db"
)

func main() {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	if dbHost == "" || dbUser == "" || dbPass == "" || dbName == ""{
		log.Fatal("Error: Environment variables POSTGRES_HOST, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, must be set")
	}

	cfg := db.PGConfig{
		Host:     dbHost,
		Port:     5432,   
		User:     dbUser,
		Password: dbPass,
		DBName:   dbName,
		SSLMode:  "disable",
	}

	pg, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("Critical DB connection error: %v", err)
	}
	defer pg.Close()

	log.Println("Database connection established successfully")

	router := api.NewRouter(pg)

	port := ":8080"
	log.Printf("Starting HTTP server on %s", port)
	
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}