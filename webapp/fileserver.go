package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

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
