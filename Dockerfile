FROM registry.access.redhat.com/ubi8/ubi

USER nobody
EXPOSE 55680

COPY observatorium-otelcol /bin/opentelemetry-collector
ENTRYPOINT ["/bin/opentelemetry-collector"]
