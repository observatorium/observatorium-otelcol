# Releasing the Observatorium OpenTelemetry Collector distribution

This project uses [`goreleaser`](https://github.com/goreleaser/goreleaser) to manage the release of new versions.

To release a new version, simply add a tag named `vX.Y.Z`, like:

```
git tag -a v0.2.1 -m "Release v0.2.1"
git push upstream v0.2.1
```

A new GitHub workflow should be started, and at the end, a GitHub release should have been created, similar with the ["Release v0.2.1"](https://github.com/observatorium/observatorium-otelcol/releases/tag/v0.2.1).