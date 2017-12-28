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
	"os"
	"strings"

	"github.com/mmcdole/gofeed"
)

func loadFeed(u string) (*gofeed.Feed, error) {
	parser := gofeed.NewParser()
	if strings.HasPrefix(u, ".") {
		fp, err := os.Open(u)
		if err != nil {
			return nil, err
		}
		feed, err := parser.Parse(fp)
		fp.Close()
		return feed, err
	}
	return parser.ParseURL(u)
}
