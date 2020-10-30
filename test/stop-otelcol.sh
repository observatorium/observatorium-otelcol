#!/bin/bash

pid=$(cat otelcol.pid)
kill "${pid}"
if [ $? != 0 ]; then
    echo "Failed to stop the running instance. Return code: $? . Skipping tests."
    exit 2
fi

echo "✅ Observatorium OpenTelemetry Collector stopped."
