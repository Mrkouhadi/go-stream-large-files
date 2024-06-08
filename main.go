package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/pkg/profile"
)

const uploadDir = "./uploads"
const uploadDirPerm = 0755

var passPhrase = goDotEnvVariable("AES_PASSPHRASE")

func main() {

	// Profiling our program
	defer profile.Start(profile.MemProfileRate(1), profile.MemProfile, profile.ProfilePath(".")).Stop()

	// Create upload directory if not exists
	err := os.MkdirAll(uploadDir, uploadDirPerm)
	if err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Initialize the FilServer with a passphrase
	server := &FilServer{
		passphrase: passPhrase,
	}

	// Start the TCP server in a goroutine
	go server.Start()

	// build an HTTP server
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*", "http://localhost:3000"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
		Debug:            true,
	}))
	// render uploads files
	fileServer := http.FileServer(http.Dir("./uploads/"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads", fileServer))

	r.Post("/upload", uploadHandler)
	r.Get("/download/", downloadHandler) // http://localhost:8080/download/FileName.extension

	// Start the HTTP server
	port := ":8080"
	fmt.Printf("Starting server on %s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	// if we have specific file names like this: system.env & others.env we can just do : godotenv.Load("system.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
