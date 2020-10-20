# Observatorium OpenTelemetry Collector distribution

This is the OpenTelemetry Collector distribution for Observatorium. It's composed of a manifest that is used to build the actual distribution using the [OpenTelemetry Collector Builder](https://github.com/observatorium/opentelemetry-collector-builder), plus a few modules that aren't available in any upstream repository.

Right now, the following modules are part of the distribution:

* All modules (extensions, receivers, processors and exporters) from the core [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector/) distribution
* [`k8sprocessor`](github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor) from OpenTelemetry Collector Contrib
* [`resourcedetectionprocessor`](github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor) from OpenTelemetry Collector Contrib
* [`routingprocessor`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/processor/routingprocessor) from [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)


## Building

To build this project, you'll need the OpenTelemetry Collector Builder, which can be installed using:

```console
$ go get github.com/observatorium/opentelemetry-collector-builder
```

Then, from the root of this project, run:

```console
$ opentelemetry-collector-builder --config manifest.yaml
```

This is the expected outcome:

```
$ opentelemetry-collector-builder --config manifest.yaml
2020-09-16T12:15:17.089+0200	INFO	cmd/root.go:83	Using config file	{"path": "manifest.yaml"}
2020-09-16T12:15:17.092+0200	INFO	builder/main.go:80	Sources created	{"path": "./_build"}
2020-09-16T12:15:17.165+0200	INFO	builder/main.go:91	Compiling
2020-09-16T12:15:33.100+0200	INFO	builder/main.go:97	Compiled	{"binary": "./_build/observatorium-otelcol"}
```

The resulting binary is located at `_build/observatorium-otelcol` and can be started with:

```
$ _build/observatorium-otelcol --config collector.yaml
```

The provided OpenTelemetry Collector configuration example (`collector.yaml`) has a regular OTLP gRPC receiver and exports data to a Jaeger collector via gRPC (port 14250) that is expected to be running on localhost.

All extra processors are referenced in the configuration, making sure that the distribution knows about them. Not all of the processors are part of the pipeline.
