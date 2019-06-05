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
