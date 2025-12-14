package request

import (
	"fmt"
	"io"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var ERROR_BAD_REQUEST_HEADER = fmt.Errorf("Bad Request Header")

func parseRequestLine(b []byte) (*RequestLine, string, error) {

}

func RequestFromReader(reader io.Reader) (*Request, error) {

}
