package meta

import (
	"bytes"
	"io"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestClient(t *testing.T) {
	t.Run("Buffer", func(t *testing.T) {
		client := GraphAPIClient{}
		expected := "hello buffer"
		input := []byte(expected)

		buf := client.Buffer(input)

		assert.Equal(t, expected, buf.String())
	})

	t.Run("Reader", func(t *testing.T) {
		client := GraphAPIClient{}
		expected := "hello reader"
		input := []byte(expected)

		reader := client.Reader(input)
		readBytes, err := io.ReadAll(reader)

		assert.NoError(t, err)
		assert.Equal(t, expected, string(readBytes))
	})

	t.Run("Writer", func(t *testing.T) {
		client := GraphAPIClient{}
		initialData := []byte("hello ")
		newData := []byte("writer")
		expected := "hello writer"

		writer := client.Writer(initialData)
		n, err := writer.Write(newData)

		assert.NoError(t, err)
		assert.Equal(t, len(newData), n)

		buf, ok := writer.(*bytes.Buffer)
		assert.True(t, ok)
		assert.Equal(t, expected, buf.String())
	})
}
