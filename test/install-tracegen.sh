#!/bin/bash

d=$(mktemp -d)
cd "$d"
go mod init temp >/dev/null 2>&1
go get "github.com/open-telemetry/opentelemetry-collector-contrib/tracegen@v0.22.0"
go install "github.com/open-telemetry/opentelemetry-collector-contrib/tracegen"
rm -r "$d"
