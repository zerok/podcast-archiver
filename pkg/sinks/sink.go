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
)

type Configuration struct {
	GoogleProjectID         string `yaml:"google_project_id"`
	Bucket                  string `yaml:"bucket"`
	AccessKeyID             string `yaml:"access_key_id"`
	AccessKeySecret         string `yaml:"access_key_secret"`
	Region                  string `yaml:"region"`
	DropboxOAuthAccessToken string `yaml:"dropbox_oauth_access_token"`
}

type BucketObject struct {
	Name   string
	Prefix string
	Bucket string
}

type Sink interface {
	ListObjects(context.Context, string) ([]BucketObject, error)
	CreateObject(context.Context, string, io.Reader) error
}
