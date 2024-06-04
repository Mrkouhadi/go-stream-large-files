package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

////// server

type FilServer struct{}

func HandleError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func (fs *FilServer) Start() {
	ln, err := net.Listen("tcp", ":9000")
	HandleError(err)
	for {
		conn, err := ln.Accept()
		HandleError(err)
		go fs.ReadLoop(conn)
	}
}

func (fs *FilServer) ReadLoop(conn net.Conn) {
	for {
		var size int64
		err := binary.Read(conn, binary.LittleEndian, &size) // get the size that was written into conn
		if err != nil {
			log.Println("Connection closed or error:", err)
			return
		}

		filePath := fmt.Sprintf("./received/file_%d", time.Now().Unix())
		outFile, err := os.Create(filePath)
		if err != nil {
			log.Println("Failed to create file:", err)
			return
		}
		defer outFile.Close()

		// Stream the file in chunks
		buffer := make([]byte, 1024*1024) // 1MB buffer
		var receivedBytes int64
		for receivedBytes < size {
			n, err := conn.Read(buffer)
			if err != nil && err != io.EOF {
				log.Println("Error during file transfer:", err)
				return
			}
			if n == 0 {
				break
			}
			_, err = outFile.Write(buffer[:n])
			if err != nil {
				log.Println("Error writing to file:", err)
				return
			}
			receivedBytes += int64(n)
		}

		fmt.Printf("Received %d bytes over the network, saved as %s\n", receivedBytes, filePath)
	}
}

func SendFile(filePath string, progressChan chan<- float64) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", ":9000")
	if err != nil {
		return err
	}
	defer conn.Close()

	fileSize := fileInfo.Size()
	err = binary.Write(conn, binary.LittleEndian, int64(fileSize)) // write the size into conn because we need it in readLoop func
	if err != nil {
		return err
	}

	// Stream the file in chunks
	buffer := make([]byte, 1024*1024) // 1MB buffer
	var sentBytes int64
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		_, err = conn.Write(buffer[:n])
		if err != nil {
			return err
		}
		sentBytes += int64(n)
		progressChan <- float64(sentBytes) / float64(fileSize) * 100
	}

	fmt.Printf("File %s sent successfully\n", filePath)
	return nil
}
