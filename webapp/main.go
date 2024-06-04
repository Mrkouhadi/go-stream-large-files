package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/profile"
)

const uploadDir = "./uploads"

func main() {
	// Profiling our program
	defer profile.Start(profile.MemProfileRate(1), profile.MemProfile, profile.ProfilePath(".")).Stop()

	// Create upload directory if not exists
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Start the TCP server in a goroutine
	server := &FilServer{}
	go server.Start()

	// Define HTTP routes
	/*
			<form Method="POST" action="http://localhost:8080/upload"  enctype="multipart/form-data" >
		        <input type="file" id="file" name="file"/>
		        <input type="submit" value="UPLOAD"/>
		    </form>
	*/
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download/", downloadHandler) // http://localhost:8080/download/uploads/FileName

	// Start the HTTP server
	port := ":8080"
	fmt.Printf("Starting server on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// uploadHandler handles file uploads
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	defer file.Close()

	filePath := filepath.Join(uploadDir, header.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Use SendFile to send the file to the TCP server with progress reporting
	progressChan := make(chan float64)
	go func() {
		if err := SendFile(filePath, progressChan); err != nil {
			log.Printf("Failed to send file: %v", err)
		}
		close(progressChan)
	}()

	// Stream progress updates to the client
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for progress := range progressChan {
		fmt.Fprintf(w, "data: %.2f\n\n", progress)
		w.(http.Flusher).Flush()
	}

	fmt.Fprintf(w, "File uploaded and sent successfully: %s\n", header.Filename)
}

// downloadHandler handles file downloads
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := filepath.Base(r.URL.Path)
	filePath := filepath.Join(uploadDir, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}
