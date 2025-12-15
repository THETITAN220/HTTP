package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
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
var ERROR_UNSUPPORTED_HTTP = fmt.Errorf("Invalid HTTP in header")
var ERROR_UNSUPPORTED_HTTP_VER = fmt.Errorf("Invalid HTTP version in header")
var SEPARATOR = "\r\n"

func parseRequestLine(b string) (*RequestLine, string, error) {
	index := strings.Index(b, SEPARATOR)
	if index != -1 {
		return nil, b, nil
	}

	startLine := b[:index]
	restOfMsg := b[index+len(SEPARATOR):]

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return nil, restOfMsg, ERROR_BAD_REQUEST_HEADER
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, restOfMsg, ERROR_UNSUPPORTED_HTTP_VER
	}

	reqLine := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpParts[1],
	}

	return reqLine, restOfMsg, nil

}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("Unable to read io.RealAll()\n"),
			err,
		)
	}

	str := string(data)

	reqLine, str, err := parseRequestLine(str)

	return &Request{
		RequestLine: *reqLine,
	}, nil
}
