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
	"bytes"
	"context"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/gosimple/slug"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/zerok/podcast-archiver/internal/notifications"
	"github.com/zerok/podcast-archiver/internal/sinks"
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

	var n *notifications.Notifications

	n, err = notifications.NewFromEnv()
	if err != nil {
		log.WithError(err).Fatal("Failed to create notification channel")
	}

	if n != nil {
		if err := n.Login(ctx); err != nil {
			log.WithError(err).Fatal("Failed to log into Matrix")
		}
		if err := n.JoinRoom(ctx); err != nil {
			log.WithError(err).Fatal("Failed to join Matrix room")
		}
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
		fngen, err := newFileNameGenerator(&feed)
		if err != nil {
			log.WithError(err).Fatalf("Failed to generate filename generator: %s", err.Error())
		}
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
				filename, err := fngen.GenerateFileName(ctx, f, item, enc)
				if err != nil {
					log.Warnf("%s could not be parsed", enc.URL)
					continue
				}
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
				if err := resp.Body.Close(); err != nil {
					log.WithError(err).Fatalf("Failed close %s", key)
				}
				if n != nil {
					if err := n.Send(ctx, fmt.Sprintf("%s/%s (%s) archived", feed.Folder, filename, html.EscapeString(item.Title))); err != nil {
						log.WithError(err).Errorf("Failed to send notification for %s", key)
					}
				}
			}
		}
		cancel()
	}
}

func getFileNameFromURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	elems := strings.Split(parsed.Path, "/")
	if len(elems) == 0 {
		return "", fmt.Errorf("failed to split path")
	}
	return elems[len(elems)-1], nil
}

type fileNameGenerator struct {
	tpl *template.Template
	cfg *feedConfiguration
}

func newFileNameGenerator(feed *feedConfiguration) (*fileNameGenerator, error) {
	g := &fileNameGenerator{}
	g.cfg = feed
	if feed.FileNameTemplate != "" {
		tpl, err := template.New("root").Funcs(map[string]interface{}{
			"fileName": getFileNameFromURL,
			"formatDate": func(v time.Time, fmt string) string {
				return v.Format(fmt)
			},
			"slugify": func(v string) string {
				return slug.Make(v)
			},
		}).Parse(feed.FileNameTemplate)
		if err != nil {
			return nil, err
		}
		g.tpl = tpl
	}
	return g, nil
}

func (g *fileNameGenerator) GenerateFileName(ctx context.Context, feed *gofeed.Feed, item *gofeed.Item, enc *gofeed.Enclosure) (string, error) {
	if g.tpl == nil {
		return getFileNameFromURL(enc.URL)
	}
	out := bytes.Buffer{}
	if err := g.tpl.ExecuteTemplate(&out, "root", fileNameGeneratorContent{
		Feed:      feed,
		Item:      item,
		Enclosure: enc,
	}); err != nil {
		return "", err
	}
	return out.String(), nil
}

type fileNameGeneratorContent struct {
	Feed      *gofeed.Feed
	Item      *gofeed.Item
	Enclosure *gofeed.Enclosure
}
