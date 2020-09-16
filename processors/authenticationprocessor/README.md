# Authentication Processor

* Supported pipeline types: traces

This processor authenticates the incoming traces by extracting the authentication data from the context
and verifying the bearer token with the specified provider.

Currently, only bearer token authentication is supported, added as part of `PerRPC` gRPC authentication.
It requires the gRPC client to send a header named `authorization` in line with the equivalent HTTP/2 header.

Examples:
```yaml
processors:
  authentication:
    oidc:
      issuer_url: https://auth.example.com/
      issuer_ca_path: /etc/pki/tls/cert.pem
      client_id: my-oidc-client
      username_claim: email
```

Refer to [config.yaml](./testdata/config.yaml) for detailed examples on using the processor.

## Usage with the OpenTelemetry Collector Builder

This module can be used with the [OpenTelemetry Collector Builder](https://github.com/observatorium/opentelemetry-collector-builder) by adding this to the manifest:

```yaml
processors:
  - gomod: github.com/observatorium/observatorium-otelcol/processors/authenticationprocessor
```
