package request

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	state       state
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type state int

const (
	requestStateInitialized state = iota
	requestStateDone
)

const (
	clrf       = "\r\n"
	bufferSize = 8
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{
		state: requestStateInitialized,
	}
	for req.state != requestStateDone {
		// If buffer is too small, double it
		if readToIndex >= len(buf) {
			tempBuf := make([]byte, 2*len(buf), 2*len(buf))
			copy(tempBuf, buf)
			buf = tempBuf
		}

		// Read new bytes AFTER what has been read
		nBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += nBytesRead

		nBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		// Keep buffer size small by shifting out already parsed data
		if nBytesParsed > 0 {
			copy(buf, buf[nBytesParsed:])
			readToIndex -= nBytesParsed
		}
	}
	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(clrf))
	if idx == -1 {
		return nil, 0, nil
	}

	// Get all three parts of the request line
	parts := strings.Split(string(data[:idx]), " ")
	if len(parts) != 3 {
		return nil, idx, errors.New("request line is invalid format")
	}
	for _, part := range parts {
		if part == "" {
			return nil, idx, errors.New("request line is invalid format")
		}
	}

	isAllCaps := regexp.MustCompile(`^[A-Z]+$`).MatchString
	if !isAllCaps(parts[0]) {
		return nil, idx, errors.New("request method invalid")
	}
	if parts[2] != "HTTP/1.1" {
		return nil, idx, errors.New("only HTTP/1.1 supported")
	}

	rl := &RequestLine{
		HttpVersion:   "1.1",
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	return rl, idx + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	requestLine, bytesParsed, err := parseRequestLine(data)
	if err != nil {
		return bytesParsed, err
	}

	// If more data needed
	if bytesParsed == 0 {
		return 0, nil
	}

	// Update Request
	r.state = requestStateDone
	r.RequestLine = *requestLine

	// Return the number of bytes parsed successfully
	return bytesParsed, nil
}
