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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func main() {
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

	sess, err := session.NewSession(aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(cfg.Sink.AccessKeyID, cfg.Sink.AccessKeySecret, "")).
		WithRegion(cfg.Sink.Region))
	if err != nil {
		log.WithError(err).Fatalf("Failed to open AWS session")
	}

	uploader := s3manager.NewUploader(sess)
	svc := s3.New(sess)

	for _, feed := range cfg.Feeds {
		log.Infof("Downloading items from %s", feed.URL)
		f, err := loadFeed(feed.URL)
		if err != nil {
			log.WithError(err).Fatalf("Failed to parse %s", feed.URL)
		}
		knownFiles := make(map[string]struct{})
		bucketPrefix := fmt.Sprintf("%s/", feed.Folder)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		// Let's retrieve a list of items we already have in the bucket so that
		// we don't upload something twice.
		bucketList, err := svc.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
			Bucket: &cfg.Sink.Bucket,
			Prefix: &bucketPrefix,
		})
		if err != nil {
			cancel()
			log.WithError(err).Fatalf("Failed to list objects in %s", cfg.Sink.Bucket)
		}
		for _, key := range bucketList.Contents {
			knownFiles[strings.TrimPrefix(*key.Key, bucketPrefix)] = struct{}{}
		}
		for _, item := range f.Items {
			for _, enc := range item.Enclosures {
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

				input := s3manager.UploadInput{
					Bucket: &cfg.Sink.Bucket,
					Body:   resp.Body,
					Key:    &key,
				}

				_, err = uploader.Upload(&input)
				if err != nil {
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
