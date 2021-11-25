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
package sinks

import (
	"context"
	"io"

	gcstorage "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type GCSSink struct {
	bucket string
	client *gcstorage.Client
}

func NewGCSSink(ctx context.Context, cfg Configuration) (*GCSSink, error) {
	client, err := gcstorage.NewClient(ctx)
	return &GCSSink{
		client: client,
		bucket: cfg.Bucket,
	}, err
}

func (s *GCSSink) CreateObject(ctx context.Context, path string, input io.Reader) error {
	output := s.client.Bucket(s.bucket).Object(path).NewWriter(ctx)
	_, err := io.Copy(output, input)
	if err != nil {
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	return nil
}

func (s *GCSSink) ListObjects(ctx context.Context, prefix string) ([]BucketObject, error) {
	q := gcstorage.Query{
		Delimiter: "/",
		Prefix:    prefix,
	}
	it := s.client.Bucket(s.bucket).Objects(ctx, &q)
	result := make([]BucketObject, 0, 10)
	for {
		obj, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		result = append(result, BucketObject{
			Name:   obj.Name,
			Prefix: obj.Prefix,
			Bucket: obj.Bucket,
		})

	}
	return result, nil
}
