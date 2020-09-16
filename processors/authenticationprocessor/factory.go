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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// typeStr is the value of "type" for this processor in the configuration.
	typeStr configmodels.Type = "authentication"
)

var (
	errNoOIDCProvided     = errors.New("no OIDC information provided")
	errNoClientIDProvided = errors.New("no ClientID provided for the OIDC configuration")
	errNoIssuerURL        = errors.New("no IssuerURL provided for the OIDC configuration")
)

// NewFactory creates a factory for the routing processor.
func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		processorhelper.WithTraces(createTraceProcessor),
	)
}

func createDefaultConfig() configmodels.Processor {
	return &Config{
		ProcessorSettings: configmodels.ProcessorSettings{
			TypeVal: typeStr,
			NameVal: string(typeStr),
		},
	}
}

func createTraceProcessor(_ context.Context, params component.ProcessorCreateParams, cfg configmodels.Processor, nextConsumer consumer.TraceConsumer) (component.TraceProcessor, error) {
	oCfg := cfg.(*Config)
	if oCfg.OIDC == nil {
		return nil, errNoOIDCProvided
	}
	if oCfg.OIDC.Audience == "" {
		return nil, errNoClientIDProvided
	}
	if oCfg.OIDC.IssuerURL == "" {
		return nil, errNoIssuerURL
	}

	return newAuthenticationProcessor(params.Logger, nextConsumer, *oCfg)
}
