package headers

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

// Takes a case insensitive key and returns the value if key in headers.
// Returns empty string if key does not exist.
func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

// Takes a key and value pair to add to the headers. The key will be treated
// as case insensitive. If key already exists, the value will be appended to
// to the previous value, delimited by a comma.
func (h Headers) Set(key, value string) {
	// NOTE: Perhaps de-dupe values using a set?
	keyLowered := strings.ToLower(key)
	if v, ok := h[keyLowered]; ok {
		value = v + ", " + value
	}
	h[keyLowered] = value
}

// Takes a key and value pair to add to the headers. The key will be treated
// as case insensitive. If the key already exists, previous value will be
// overwritten.
func (h Headers) Replace(key, value string) {
	keyLowered := strings.ToLower(key)
	h[keyLowered] = value
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Look for CRLF
	idx := bytes.Index(data, []byte(crlf))

	// If not found, assume we need more data
	if idx == -1 {
		return 0, false, nil
	}

	// If data starts with CLRF, return done with headers
	if idx == 0 {
		return 2, true, nil
	}

	// Split data into key->value pair on the ':' separator
	keyBuffer, valueBuffer, found := bytes.Cut(data[:idx], []byte(":"))
	if !found || len(keyBuffer) == 0 || len(valueBuffer) == 0 {
		// Missing key or value
		msg := fmt.Sprintf("invalid header: %s", data[:idx])
		return 0, false, errors.New(msg)
	}

	key := string(keyBuffer)
	// Whitespace is allowed before the key, but not after
	key = strings.TrimLeft(key, " ")

	// Only allow letters, digits, and specific special characters.
	validKey := regexp.MustCompile(`^[a-zA-Z0-9!@#$%&'\*+-~]+$`).MatchString
	if !validKey(key) {
		// To many names on left side of separator
		msg := fmt.Sprintf("invalid header: %s", data[:idx])
		return 0, false, errors.New(msg)
	}

	value := string(valueBuffer)
	value = strings.TrimSpace(value)

	// Add key->value to headers
	h.Set(key, value)

	// Report how many bytes were parsed
	return idx + 2, false, nil
}
