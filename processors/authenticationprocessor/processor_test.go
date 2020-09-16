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
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

var (
	logger, _ = zap.NewDevelopment()
)

func TestProcessorCapabilities(t *testing.T) {
	// prepare
	config := Config{
		OIDC: &OIDC{
			IssuerURL: "http://example.com/",
			Audience:  "unit-test",
		},
	}

	// test
	p, err := newAuthenticationProcessor(logger, exportertest.NewNopTraceExporter(), config)
	caps := p.GetCapabilities()

	// verify
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, false, caps.MutatesConsumedData)
}

func TestInvalidNextConsumer(t *testing.T) {
	// prepare
	config := Config{
		OIDC: &OIDC{},
	}

	// test
	p, err := newAuthenticationProcessor(logger, nil, config)

	// verify
	assert.Equal(t, componenterror.ErrNilNextConsumer, err)
	assert.Nil(t, p)
}

func TestInvalidProcessorConfiguration(t *testing.T) {
	// prepare
	config := Config{}

	// test
	p, err := newAuthenticationProcessor(logger, exportertest.NewNopTraceExporter(), config)

	// verify
	assert.Equal(t, errNoOIDCProvided, err)
	assert.Nil(t, p)
}

func TestNotAuthenticated(t *testing.T) {
	// prepare
	config := Config{
		OIDC: &OIDC{},
	}
	p, err := newAuthenticationProcessor(logger, exportertest.NewNopTraceExporter(), config)
	require.NoError(t, err)

	// test
	err = p.ConsumeTraces(context.Background(), pdata.NewTraces())

	// verify
	assert.Equal(t, errAuthenticationRequired, err)
}

func TestAuthenticatorFails(t *testing.T) {
	// prepare
	theErr := errors.New("the error")
	p := &authenticationProcessor{
		logger: logger,
		config: Config{
			OIDC: &OIDC{},
		},
		authenticator: &mockAuthenticator{
			succeed: false,
			err:     theErr,
		},
	}

	// test
	err := p.ConsumeTraces(context.Background(), pdata.NewTraces())

	// verify
	assert.Equal(t, theErr, err)
}

func TestStartsAuthenticator(t *testing.T) {
	// prepare
	theErr := errors.New("the error")
	p := &authenticationProcessor{
		logger: logger,
		config: Config{
			OIDC: &OIDC{},
		},
		authenticator: &mockAuthenticator{
			// we test by making the underlying authenticator fail, and checking that the processor failed
			startErr: theErr,
		},
	}

	// test
	err := p.Start(context.Background(), nil)

	// verify
	assert.Equal(t, theErr, err)
}

func TestShutdownAuthenticator(t *testing.T) {
	// prepare
	theErr := errors.New("the error")
	p := &authenticationProcessor{
		logger: logger,
		config: Config{
			OIDC: &OIDC{},
		},
		authenticator: &mockAuthenticator{
			// we test by making the underlying authenticator fail, and checking that the processor failed
			shutdownErr: theErr,
		},
	}

	// test
	err := p.Shutdown(context.Background())

	// verify
	assert.Equal(t, theErr, err)
}

func TestNextConsumerOnAuthSuccess(t *testing.T) {
	// prepare
	wg := &sync.WaitGroup{}
	wg.Add(1)

	next, err := processorhelper.NewTraceProcessor(&Config{}, exportertest.NewNopTraceExporter(), &mockProcessor{
		onTraces: func(context.Context, pdata.Traces) (pdata.Traces, error) {
			wg.Done()
			return pdata.NewTraces(), nil
		},
	})
	require.NoError(t, err)

	p := &authenticationProcessor{
		logger: logger,
		config: Config{
			OIDC: &OIDC{},
		},
		nextConsumer: next,
		authenticator: &mockAuthenticator{
			succeed: true,
		},
	}

	// test
	err = p.ConsumeTraces(context.Background(), pdata.NewTraces())

	// verify
	wg.Wait()
	assert.NoError(t, err)
}

type mockProcessor struct {
	onTraces func(context.Context, pdata.Traces) (pdata.Traces, error)
}

func (mp *mockProcessor) ProcessTraces(ctx context.Context, traces pdata.Traces) (pdata.Traces, error) {
	if mp.onTraces != nil {
		return mp.onTraces(ctx, traces)
	}
	return pdata.NewTraces(), nil
}
