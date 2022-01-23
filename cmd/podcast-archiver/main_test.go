package main

import (
	"context"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/require"
)

func TestFileNameGenerator(t *testing.T) {
	t.Run("without-template", func(t *testing.T) {
		t.Run("only-filepath", func(t *testing.T) {
			g, err := newFileNameGenerator(&feedConfiguration{})
			require.NoError(t, err)
			feed := &gofeed.Feed{}
			item := &gofeed.Item{}
			enc := &gofeed.Enclosure{
				URL: "https://domain.com/prefix/path.mp3?some-more-stuff",
			}
			output, err := g.GenerateFileName(context.Background(), feed, item, enc)
			require.NoError(t, err)
			expected := "path.mp3"
			require.Equal(t, expected, output)
		})
	})
	t.Run("with-template", func(t *testing.T) {
		t.Run("url-filename", func(t *testing.T) {
			g, err := newFileNameGenerator(&feedConfiguration{
				FileNameTemplate: "{{ (fileName .Enclosure.URL) }}",
			})
			require.NoError(t, err)
			feed := &gofeed.Feed{}
			item := &gofeed.Item{}
			enc := &gofeed.Enclosure{
				URL: "https://domain.com/prefix/path.mp3?some-more-stuff",
			}
			output, err := g.GenerateFileName(context.Background(), feed, item, enc)
			require.NoError(t, err)
			expected := "path.mp3"
			require.Equal(t, expected, output)
		})
		t.Run("with-seasons", func(t *testing.T) {
			g, err := newFileNameGenerator(&feedConfiguration{
				FileNameTemplate: "{{ (formatDate .Item.PublishedParsed \"2006-01-02\") }}-{{ .Item.Title|slugify }}.mp3",
			})
			require.NoError(t, err)
			feed := &gofeed.Feed{}
			item := &gofeed.Item{}
			item.Title = "Hello World"
			item.Published = "Tue, 19 Oct 2021 07:05:00 +0000"
			parsed, err := time.Parse(time.RFC1123Z, item.Published)
			require.NoError(t, err)
			item.PublishedParsed = &parsed
			enc := &gofeed.Enclosure{
				URL: "https://domain.com/prefix/path.mp3?some-more-stuff",
			}
			output, err := g.GenerateFileName(context.Background(), feed, item, enc)
			require.NoError(t, err)
			expected := "2021-10-19-hello-world.mp3"
			require.Equal(t, expected, output)
		})
	})
}
