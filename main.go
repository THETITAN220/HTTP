package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	file, err := os.Open("./messages.txt")

	if err != nil {
		log.Fatal("Error : ", err)
	}

	defer file.Close()

	for {
		data := make([]byte, 8)
		n, err := file.Read(data)

		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}

		fmt.Printf("Read : %s\n", string(data[:n]))
	}
}
