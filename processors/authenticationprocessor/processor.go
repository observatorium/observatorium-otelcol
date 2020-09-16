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

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
)

type authenticationProcessor struct {
	nextConsumer  consumer.TraceConsumer
	config        Config
	logger        *zap.Logger
	authenticator authenticator
}

var (
	_ component.TraceProcessor = (*authenticationProcessor)(nil)

	errAuthenticationRequired = errors.New("authentication required")
)

func newAuthenticationProcessor(l *zap.Logger, nextConsumer consumer.TraceConsumer, config Config) (*authenticationProcessor, error) {
	logger := l.With(zap.String("processor", "authentication"))
	if nextConsumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}

	if config.OIDC == nil {
		return nil, errNoOIDCProvided
	}

	p := &authenticationProcessor{
		logger:        logger,
		nextConsumer:  nextConsumer,
		config:        config,
		authenticator: newOIDCAuthenticator(logger, *config.OIDC),
	}

	return p, nil
}

func (p *authenticationProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	authenticated, err := p.authenticator.authenticate(ctx)
	if err != nil {
		return err
	}

	if !authenticated {
		return errAuthenticationRequired
	}

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *authenticationProcessor) GetCapabilities() component.ProcessorCapabilities {
	return component.ProcessorCapabilities{MutatesConsumedData: false}
}

// Start is invoked during service startup.
func (p *authenticationProcessor) Start(ctx context.Context, host component.Host) error {
	return p.authenticator.start(ctx, host)
}

// Shutdown is invoked during service shutdown.
func (p *authenticationProcessor) Shutdown(ctx context.Context) error {
	return p.authenticator.shutdown(ctx)
}
