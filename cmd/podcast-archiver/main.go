// Copyright 2017, Horst Gutmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/zerok/podcast-archiver/pkg/sinks"
)

func main() {
	ctx := context.Background()
	log := logrus.New()
	var configPath string
	var verbose bool
	pflag.StringVar(&configPath, "config", "-", "Path to a config file. (Default: stdin)")
	pflag.BoolVar(&verbose, "verbose", false, "Verbose logging")
	pflag.Parse()

	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.WithError(err).Fatalf("Failed to load configuration from '%s'", configPath)
	}

	var sink sinks.Sink
	if cfg.Sink.GoogleProjectID != "" {
		sink, err = sinks.NewGCSSink(ctx, cfg.Sink)
	} else if cfg.Sink.DropboxOAuthAccessToken != "" {
		sink, err = sinks.NewDropboxSink(ctx, cfg.Sink)
	} else if cfg.Sink.FileSystemFolder != "" {
		sink, err = sinks.NewFSSink(ctx, cfg.Sink)
	} else {
		sink, err = sinks.NewS3Sink(ctx, cfg.Sink)
	}

	if err != nil {
		log.WithError(err).Fatal("Failed to setup sink.")
	}

	for _, feed := range cfg.Feeds {
		log.Infof("Downloading items from %s", feed.URL)
		f, err := loadFeed(feed.URL)
		if err != nil {
			log.WithError(err).Fatalf("Failed to parse %s", feed.URL)
		}
		knownFiles := make(map[string]struct{})
		bucketPrefix := fmt.Sprintf("%s/", feed.Folder)
		timedctx, cancel := context.WithTimeout(ctx, time.Second*5)
		// Let's retrieve a list of items we already have in the bucket so that
		// we don't upload something twice.
		objects, err := sink.ListObjects(timedctx, bucketPrefix)
		if err != nil {
			cancel()
			log.WithError(err).Fatalf("Failed to list objects in %s", cfg.Sink.Bucket)
		}
		for _, obj := range objects {
			knownFiles[strings.TrimPrefix(obj.Name, bucketPrefix)] = struct{}{}
		}
		for _, item := range f.Items {
			for _, enc := range item.Enclosures {
				if enc.URL == "" {
					continue
				}
				filename := getFilename(enc.URL)
				if _, found := knownFiles[filename]; found {
					log.Debugf("%s already uploaded", enc.URL)
					continue
				}
				log.Infof("Archiving %s", enc.URL)
				resp, err := http.Get(enc.URL)
				if err != nil {
					cancel()
					log.WithError(err).Fatalf("Failed to download '%s'", enc.URL)
				}
				if resp.StatusCode != http.StatusOK {
					cancel()
					log.Fatalf("%s resulted in status code %v", enc.URL, resp.StatusCode)
				}
				key := fmt.Sprintf("%s/%s", feed.Folder, filename)

				if err := sink.CreateObject(ctx, key, resp.Body); err != nil {
					resp.Body.Close()
					log.WithError(err).Fatalf("Failed to upload %s", key)
				}
				resp.Body.Close()
			}
		}
		cancel()
	}
}

func getFilename(u string) string {
	segments := strings.Split(u, "/")
	return segments[len(segments)-1]
}
