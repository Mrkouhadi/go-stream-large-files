package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

// FilServer struct with passphrase field
type FilServer struct {
	passphrase string
}

// HandleError is a helper function to handle errors
func HandleError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// Start initializes the TCP server
func (fs *FilServer) Start() {
	ln, err := net.Listen("tcp", ":9000")
	HandleError(err)
	for {
		conn, err := ln.Accept()
		HandleError(err)
		go fs.ReadLoop(conn)
	}
}

// ReadLoop reads data from the connection and saves it to a file
func (fs *FilServer) ReadLoop(conn net.Conn) {
	defer conn.Close()

	var fileNameSize int64
	err := binary.Read(conn, binary.LittleEndian, &fileNameSize)
	HandleError(err)

	fileName := make([]byte, fileNameSize)
	_, err = io.ReadFull(conn, fileName)
	HandleError(err)

	filePath := filepath.Join(uploadDir, string(fileName))

	var offset int64
	if _, err := os.Stat(filePath); err == nil {
		fileInfo, _ := os.Stat(filePath)
		offset = fileInfo.Size()
	}

	err = binary.Write(conn, binary.LittleEndian, offset)
	HandleError(err)

	outFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	HandleError(err)
	defer outFile.Close()

	_, err = outFile.Seek(offset, 0)
	HandleError(err)

	var size int64
	err = binary.Read(conn, binary.LittleEndian, &size)
	HandleError(err)

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

		decryptedData, err := decrypt(buffer[:n], fs.passphrase)
		if err != nil {
			log.Println("Error decrypting data:", err)
			return
		}

		_, err = outFile.Write(decryptedData)
		if err != nil {
			log.Println("Error writing to file:", err)
			return
		}
		receivedBytes += int64(len(decryptedData))
	}

	fmt.Printf("Received %d bytes over the network, saved as %s\n", receivedBytes, filePath)
}
func SendFile(filePath string, progressChan chan<- float64, passphrase string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	conn, err := net.Dial("tcp", ":9000")
	if err != nil {
		return err
	}
	defer conn.Close()

	fileName := filepath.Base(filePath)
	fileNameSize := int64(len(fileName))

	err = binary.Write(conn, binary.LittleEndian, fileNameSize)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(fileName))
	if err != nil {
		return err
	}

	var offset int64
	err = binary.Read(conn, binary.LittleEndian, &offset)
	if err != nil {
		return err
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	err = binary.Write(conn, binary.LittleEndian, fileSize)
	if err != nil {
		return err
	}

	buffer := make([]byte, 1024*1024) // 1MB buffer
	var sentBytes int64
	for {
		// Read from file into buffer
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		// If no bytes were read and it's not the end of the file, continue reading
		if n == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		// Encrypt the data read from the file
		encryptedData, err := encrypt(buffer[:n], passphrase)
		if err != nil {
			return err
		}

		// Write encrypted data to the connection
		_, err = conn.Write(encryptedData)
		if err != nil {
			return err
		}

		// Update progress and sentBytes
		sentBytes += int64(n)
		progressChan <- float64(sentBytes+offset) / float64(fileSize+offset) * 100
	}

	fmt.Printf("File %s sent successfully\n", filePath)
	return nil
}
