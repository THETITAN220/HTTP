package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(file io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer file.Close()
		defer close(out)

		str := ""
		for {
			data := make([]byte, 8)
			n, err := file.Read(data)

			if err != nil {
				if err == io.EOF {
					break
				}
				break
			}
			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				out <- str
				str = ""
			}
			str += string(data)
		}
		if len(str) != 0 {
			out <- str
		}
	}()
	return out
}

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal("Error : ", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error : ", err)
		}
		for line := range getLinesChannel(conn) {
			fmt.Printf("Read : %s\n", line)
		}
	}
}
