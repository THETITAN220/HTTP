package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/THETITAN220/HTTP/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateDone    parserState = "done"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateError   parserState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
	Body        string
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

var ERROR_BAD_REQUEST_HEADER = fmt.Errorf("Bad Request Header")
var ERROR_UNSUPPORTED_HTTP_VER = fmt.Errorf("Invalid HTTP version in header")
var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("Invalid Request in Error State")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	index := bytes.Index(b, SEPARATOR)
	if index == -1 {
		return nil, 0, nil
	}

	startLine := b[:index]
	read := index + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_BAD_REQUEST_HEADER
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_VER
	}

	reqLine := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return reqLine, read, nil

}

func (r *Request) hasBody() bool {
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE

		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			lengthStr := getInt(r.Headers, "content-length", 0)
			if lengthStr == 0 {
				panic("CHUNKED NOT IMPLEMENTED")
			}

			remaining := min(lengthStr-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])

			if len(r.Body) == lengthStr {
				r.state = StateDone
			}

			read += remaining

		case StateDone:
			break outer

		default:
			panic("We were not supposed to reach here ")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	// NOTE : Buffer could be overrun if >1024 bytes received
	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufLen])
		bufLen -= readN

	}

	return request, nil
}
