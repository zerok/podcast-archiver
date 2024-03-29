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
package notifications

import (
	"context"
	"fmt"
	"os"

	"github.com/matrix-org/gomatrix"
)

type Notifications struct {
	username   string
	password   string
	room       string
	homeServer string
	matrix     *gomatrix.Client
}

func NewFromEnv() (*Notifications, error) {
	homeserver := os.Getenv("MATRIX_HOMESERVER")
	username := os.Getenv("MATRIX_USERNAME")
	password := os.Getenv("MATRIX_PASSWORD")
	room := os.Getenv("MATRIX_ROOM")
	if homeserver == "" || username == "" || password == "" || room == "" {
		return nil, nil
	}
	return New(homeserver, username, password, room)
}

func New(homeServer, username, password, room string) (*Notifications, error) {
	n := &Notifications{
		username:   username,
		password:   password,
		homeServer: homeServer,
		room:       room,
	}

	client, err := gomatrix.NewClient(n.homeServer, "", "")
	if err != nil {
		return nil, err
	}
	n.matrix = client
	return n, nil
}

func (n *Notifications) Login(ctx context.Context) error {
	resp, err := n.matrix.Login(&gomatrix.ReqLogin{
		Type:     "m.login.password",
		User:     n.username,
		Password: n.password,
	})
	if err != nil {
		return err
	}
	n.matrix.SetCredentials(resp.UserID, resp.AccessToken)
	return nil
}

func (n *Notifications) JoinRoom(ctx context.Context) error {
	joined, err := n.matrix.JoinedRooms()
	if err != nil {
		return fmt.Errorf("failed to fetch joined rooms: %w", err)
	}
	for _, rid := range joined.JoinedRooms {
		if rid == n.room {
			return nil
		}
	}
	if _, err := n.matrix.JoinRoom(n.room, "", nil); err != nil {
		return err
	}
	return nil
}

func (n *Notifications) Send(ctx context.Context, msg string) error {
	_, err := n.matrix.SendText(n.room, msg)
	return err
}
