// Copyright The OpenTelemetry Authors
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

package authenticationprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
)

var _ authenticator = (*mockAuthenticator)(nil)

type mockAuthenticator struct {
	succeed     bool
	err         error
	startErr    error
	shutdownErr error
}

func (m *mockAuthenticator) authenticate(context.Context) (bool, error) {
	return m.succeed, m.err
}
func (m *mockAuthenticator) start(context.Context, component.Host) error {
	return m.startErr
}
func (m *mockAuthenticator) shutdown(context.Context) error {
	return m.shutdownErr
}
