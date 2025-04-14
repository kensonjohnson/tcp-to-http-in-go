package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("     Content-Type: application/json        \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, 45, n)
	assert.False(t, done)

	// Valid 2 headers with existing headers
	headers = NewHeaders()
	headers.Set("Host", "localhost:42069")
	headers.Set("Content-Type", "application/json")
	data = []byte("Authorization: Bearer ABCD\r\nAccept: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "Bearer ABCD", headers.Get("Authorization"))
	assert.Equal(t, 28, n)
	assert.False(t, done)
	copy(data, data[28:]) // clear parsed data
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "application/json", headers.Get("Accept"))
	assert.Equal(t, 26, n)
	assert.False(t, done)

	// Valid done
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 23, n)
	assert.False(t, done)
	copy(data, data[23:]) // clear parsed data
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Valid 2 headers with same keys
	headers = NewHeaders()
	headers.Set("Set-Message", "Blue")
	data = []byte("Set-Message: Green\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "Blue, Green", headers.Get("Set-Message"))
	assert.Equal(t, 20, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header key (field-name)
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
}
