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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exportertest"
)

func TestCreateTestProcessor(t *testing.T) {
	// prepare
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	f := NewFactory()
	c := f.CreateDefaultConfig().(*Config)

	c.OIDC = &OIDC{
		Audience:  "unit-tests",
		IssuerURL: "http://example.com/",
	}
	params := component.ProcessorCreateParams{Logger: logger}

	// test
	p, err := f.CreateTraceProcessor(context.Background(), params, exportertest.NewNopTraceExporter(), c)

	// verify
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestMissingOIDCParameters(t *testing.T) {
	for _, tt := range []struct {
		casename    string
		config      *Config
		expectedErr error
	}{
		{
			"no-oidc",
			NewFactory().CreateDefaultConfig().(*Config),
			errNoOIDCProvided,
		},
		{
			"no-client-id",
			&Config{
				OIDC: &OIDC{
					IssuerURL: "http://example.com/",
				},
			},
			errNoClientIDProvided,
		},
		{
			"no-issuer-url",
			&Config{
				OIDC: &OIDC{
					Audience: "unit-tests",
				},
			},
			errNoIssuerURL,
		},
	} {
		t.Run(tt.casename, func(t *testing.T) {
			// prepare
			f := NewFactory()
			params := component.ProcessorCreateParams{}

			// test
			p, err := f.CreateTraceProcessor(context.Background(), params, nil, tt.config)

			// verify
			assert.Equal(t, err, tt.expectedErr)
			assert.Nil(t, p)
		})
	}
}
