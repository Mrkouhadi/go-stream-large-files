package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

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
		if err := SendFile(filePath, progressChan, passPhrase); err != nil { // FIXME:
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
	fileName := filepath.Base(strings.TrimLeft(r.URL.Path, "/"))
	filePath := filepath.Join(uploadDir, fileName)

	fmt.Println(filePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		fmt.Println("error hapened heeere.")

		return
	}
	fmt.Println("reached here...")

	http.ServeFile(w, r, filePath)
}
