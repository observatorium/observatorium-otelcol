// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

module github.com/observatorium/observatorium-otelcol/processors/authenticationprocessor

go 1.14

require (
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/pquerna/cachecontrol v0.0.0-20200819021114-67c6ae64274f // indirect
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.10.0
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.31.1
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
)
