package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

const minimalRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>http://example.com</link>
    <description>A minimal test feed</description>
  </channel>
</rss>`

func TestLoadFeed(t *testing.T) {
	// If the given file doesn't exist, an error should be returned
	t.Run("local-file-not-found", func(t *testing.T) {
		_, err := loadFeed("./nonexistent-feed.xml", nil)
		require.Error(t, err)
	})

	// Support loading feeds via HTTP
	t.Run("http-feed", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(minimalRSS))
		}))
		defer srv.Close()
		feed, err := loadFeed(srv.URL, srv.Client())
		require.NoError(t, err)
		require.NotNil(t, feed)
	})

	t.Run("http-feed-with-retry", func(t *testing.T) {
		requestCount := 0
		rc := retryablehttp.NewClient()
		rc.RetryWaitMax = time.Minute * 1
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if requestCount == 1 {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(minimalRSS))
		}))
		defer srv.Close()
		rc.HTTPClient = srv.Client()
		feed, err := loadFeed(srv.URL, rc.StandardClient())
		require.NoError(t, err)
		require.NotNil(t, feed)
		require.Equal(t, 2, requestCount)
	})
}
