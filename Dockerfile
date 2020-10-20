FROM registry.access.redhat.com/ubi8/ubi

USER nobody
EXPOSE 55680

ENTRYPOINT ["/bin/observatorium-otelcol"]
COPY _build/observatorium-otelcol /bin/observatorium-otelcol
