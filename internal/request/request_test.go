package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
		numBytesPerRead: 1,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data: "GET /coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
		numBytesPerRead: 2,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Good POST Request with path
	r, err = RequestFromReader(strings.NewReader(
		"POST /coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader(
		"/coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n"))
	require.Error(t, err)
	assert.Equal(t, err.Error(), "request line is invalid format")

	// Invalid method (out of order) Request line
	_, err = RequestFromReader(strings.NewReader(
		"/coffee GET HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n"))
	require.Error(t, err)
	assert.Equal(t, err.Error(), "request method invalid")

	_, err = RequestFromReader(strings.NewReader(
		"/coffee HTTP/1.1 GET\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n"))
	require.Error(t, err)
	assert.Equal(t, err.Error(), "request method invalid")

	// Invalid version in Request line
	r, err = RequestFromReader(strings.NewReader(
		"GET /coffee HTTP/2\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n"))
	require.Error(t, err)
	assert.Equal(t, err.Error(), "only HTTP/1.1 supported")
}

func TestParseHeaders(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// Test: Malformed Header
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 4,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Empty Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"\r\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	// Duplicate Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"Some-State: Green\r\n" +
			"Some-State: Blue\r\n" +
			"\r\n",
		numBytesPerRead: 6,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "Green, Blue", r.Headers.Get("some-state"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// Case Insensitive Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"host: localhost:42069\r\n" +
			"UsEr-AgEnT: curl/7.81.0\r\n" +
			"ACCEPT: */*\r\n" +
			"\r\n",
		numBytesPerRead: 7,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("HOST"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("uSeR-aGeNt"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// Missing End of Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"host: localhost:42069\r\n" +
			"UsEr-AgEnT: curl/7.81.0\r\n" +
			"ACCEPT: */*\r\n",
		numBytesPerRead: 8,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestParseBody(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 9,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 10,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Empty Body, 0 reported content length (valid)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 11,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(r.Body))

	// Empty Body, no reported content length (valid)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 12,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(r.Body))

	// No Content-Length but Body Exists (shouldn't error, we're assuming Content-Length will be present if a body exists)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"just gonna sneak this in here",
		numBytesPerRead: 13,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(r.Body))
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}
