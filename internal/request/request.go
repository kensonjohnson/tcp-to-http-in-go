package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"tcp-to-http/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
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
	requestStateParsingHeaders
	requestStateParsingBody
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
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
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
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, reached end of file at %d, bytes read: %d", req.state, nBytesRead)
				}
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
	totalBytesParsed := 0
	// Buffer could have many lines of data in it, attempt to parse as
	// many as possible.
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, bytesParsed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		// If more data needed
		if bytesParsed == 0 {
			return 0, nil
		}

		// Update Request
		r.state = requestStateParsingHeaders
		r.RequestLine = *requestLine

		// Return the number of bytes parsed successfully
		return bytesParsed, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, fmt.Errorf("error when parsing headers: %v", err)
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLengthString := r.Headers.Get("Content-Length")
		if contentLengthString == "" {
			r.state = requestStateDone
			return 0, nil
		}
		contentLength, err := strconv.Atoi(contentLengthString)
		if err != nil {
			return 0, fmt.Errorf("invalid content length value: %s", contentLengthString)
		}
		r.Body = append(r.Body, data...)
		bodyLength := len(r.Body)
		if bodyLength > contentLength {
			return 0, fmt.Errorf("invalid content length, body too long")
		}
		if bodyLength == contentLength {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, errors.New("trying to parse after 'done' state")
	default:
		return 0, errors.New("parser encountered unknown state")
	}
}
