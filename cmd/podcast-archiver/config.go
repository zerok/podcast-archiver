// Copyright 2017, Horst Gutmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"io/ioutil"
	"os"

	"github.com/zerok/podcast-archiver/internal/sinks"
	yaml "gopkg.in/yaml.v3"
)

type feedConfiguration struct {
	URL              string `yaml:"url"`
	Folder           string `yaml:"folder"`
	FileNameTemplate string `yaml:"filename_template"`
}

type configuration struct {
	Sink  sinks.Configuration `yaml:"sink"`
	Feeds []feedConfiguration `yaml:"feeds"`
}

func loadConfig(path string) (*configuration, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	var cfg configuration
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
