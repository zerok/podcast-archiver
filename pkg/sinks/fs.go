package sinks

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FSSink struct {
	rootFolder string
}

func NewFSSink(ctx context.Context, cfg Configuration) (*FSSink, error) {
	folder := cfg.FileSystemFolder
	if folder == "" {
		return nil, fmt.Errorf("no root folder specified")
	}
	return &FSSink{rootFolder: folder}, nil
}

func (s *FSSink) CreateObject(ctx context.Context, path string, input io.Reader) error {
	fullpath := filepath.Join(s.rootFolder, path)
	dir := filepath.Dir(fullpath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	fp, err := os.OpenFile(fullpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = io.Copy(fp, input)
	return err
}

func (s *FSSink) ListObjects(ctx context.Context, prefix string) ([]BucketObject, error) {
	fullpath := filepath.Join(s.rootFolder, prefix)
	entries, err := os.ReadDir(fullpath)
	result := make([]BucketObject, 0, len(entries))
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	for _, e := range entries {
		result = append(result, BucketObject{
			Name:   e.Name(),
			Prefix: prefix,
		})
	}
	return result, nil
}
