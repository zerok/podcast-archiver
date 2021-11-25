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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Sink struct {
	bucket   string
	session  *session.Session
	uploader *s3manager.Uploader
	svc      *s3.S3
}

func NewS3Sink(ctx context.Context, cfg Configuration) (*S3Sink, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.AccessKeySecret, "")).WithRegion(cfg.Region))
	if err != nil {
		return nil, err
	}

	uploader := s3manager.NewUploader(sess)
	svc := s3.New(sess)

	return &S3Sink{
		session:  sess,
		bucket:   cfg.Bucket,
		uploader: uploader,
		svc:      svc,
	}, nil
}

func (s *S3Sink) CreateObject(ctx context.Context, path string, input io.Reader) error {
	up := s3manager.UploadInput{
		Bucket: &s.bucket,
		Body:   input,
		Key:    &path,
	}

	_, err := s.uploader.Upload(&up)
	return err
}

func (s *S3Sink) ListObjects(ctx context.Context, prefix string) ([]BucketObject, error) {
	bucketList, err := s.svc.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: &s.bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, err
	}
	result := make([]BucketObject, 0, len(bucketList.Contents))
	for _, obj := range bucketList.Contents {
		result = append(result, BucketObject{
			Name:   *obj.Key,
			Prefix: *bucketList.Prefix,
			Bucket: s.bucket,
		})
	}
	return nil, nil
}
