package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"net"
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
	buff := new(bytes.Buffer)
	for {
		var size int64
		binary.Read(conn, binary.LittleEndian, &size) // get the size that was written into conn
		n, err := io.CopyN(buff, conn, 4000)          // copy bytes from conn to buff
		HandleError(err)
		fmt.Println(buff.Bytes())
		fmt.Printf("Received %d bytes over the netwrok\n", n)
	}
}

func SendFile(size int) error {
	file := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		return err
	}
	conn, err := net.Dial("tcp", ":9000")
	if err != nil {
		return err
	}
	binary.Write(conn, binary.LittleEndian, int64(size))         // write the size into conn cuz we need it in readLopp func
	n, err := io.CopyN(conn, bytes.NewReader(file), int64(size)) // CopyN means copy only a specific size not like Copy which sin't limited...
	if err != nil {
		return err
	}
	fmt.Printf("Written %d bytes over the network\n", n)
	return nil
}
