package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFileName(t *testing.T) {
	// Make sure that the filename is normalized:
	t.Run("only-filepath", func(t *testing.T) {
		output, err := getFilename("https://domain.com/prefix/path.mp3?some-more-stuff")
		require.NoError(t, err)
		expected := "path.mp3"
		require.Equal(t, expected, output)
	})
}
