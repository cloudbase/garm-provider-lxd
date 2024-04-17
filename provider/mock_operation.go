// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright 2023 Cloudbase Solutions SRL
//
// Licensed under the AGPLv3, see LICENCE file for details

package provider

import (
	"context"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
)

type MockOperation struct {
	mock.Mock
}

func (m *MockOperation) Wait() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOperation) AddHandler(handler func(api.Operation)) (target *lxd.EventTarget, err error) {
	args := m.Called(handler)
	return args.Get(0).(*lxd.EventTarget), args.Error(1)
}

func (m *MockOperation) Cancel() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOperation) Get() api.Operation {
	args := m.Called()
	return args.Get(0).(api.Operation)
}

func (m *MockOperation) GetWebsocket(secret string) (conn *websocket.Conn, err error) {
	args := m.Called(secret)
	return args.Get(0).(*websocket.Conn), args.Error(1)
}

func (m *MockOperation) RemoveHandler(target *lxd.EventTarget) error {
	args := m.Called(target)
	return args.Error(0)
}

func (m *MockOperation) Refresh() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOperation) WaitContext(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
