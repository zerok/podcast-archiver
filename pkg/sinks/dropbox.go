package sinks

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	dfiles "github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

type DropboxSink struct {
	dropboxCfg dropbox.Config
}

func NewDropboxSink(ctx context.Context, cfg Configuration) (*DropboxSink, error) {
	dbxCfg := dropbox.Config{
		Token: cfg.DropboxOAuthAccessToken,
	}
	return &DropboxSink{
		dropboxCfg: dbxCfg,
	}, nil
}

func (s *DropboxSink) ListObjects(ctx context.Context, path string) ([]BucketObject, error) {
	normalizedPath := "/" + path
	dbf := dfiles.New(s.dropboxCfg)
	result := make([]BucketObject, 0, 10)
	arg := dfiles.NewListFolderArg(normalizedPath)
	arg.Recursive = true
	resp, err := dbf.ListFolder(arg)
	for {
		if err != nil {
			if strings.Contains(err.Error(), "path/not_found/") {
				break
			}
			return nil, err
		}
		for _, e := range resp.Entries {
			switch entry := e.(type) {
			case *dfiles.FileMetadata:
				o := BucketObject{
					Prefix: normalizedPath,
					Name:   entry.Name,
				}
				result = append(result, o)
			}
		}
		if resp.HasMore {
			carg := dfiles.NewListFolderContinueArg(resp.Cursor)
			resp, err = dbf.ListFolderContinue(carg)
		} else {
			break
		}
	}
	return result, nil
}

func (s *DropboxSink) CreateObject(ctx context.Context, path string, data io.Reader) error {
	dbf := dfiles.New(s.dropboxCfg)
	normalizedPath := "/" + path
	folder := filepath.Dir(normalizedPath)
	farg := dfiles.NewCreateFolderArg(folder)
	if _, err := dbf.CreateFolderV2(farg); err != nil {
		if !strings.Contains(err.Error(), "conflict") {
			return err
		}
	}
	arg := dfiles.NewCommitInfo(normalizedPath)
	_, err := dbf.Upload(arg, data)
	return err
}
