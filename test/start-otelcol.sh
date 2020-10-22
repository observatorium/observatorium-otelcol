#!/bin/bash

max_retries=50

# start the distribution
./_build/observatorium-otelcol --config ./test/collector.yaml > ./test/otelcol.log 2>&1 &
pid=$!

retries=0
while true
do
    kill -0 "${pid}" >/dev/null 2>&1
    if [ $? != 0 ]; then
        echo "❌ FAIL. The Observatorium OpenTelemetry Collector isn't running. Startup log:"
        cat ./test/otelcol.log
        failed=true
        exit 1
    fi

    curl -s localhost:13133 | grep "Server available" > /dev/null
    if [ $? == 0 ]; then
        echo "✅ Observatorium OpenTelemetry Collector started."

        kill "${pid}"
        if [ $? != 0 ]; then
            echo "Failed to stop the running instance. Return code: $? . Skipping tests."
            exit 2
        fi
        break
    fi

    echo "Server still unavailable" >> ./test/test.log

    let "retries++"
    if [ "$retries" -gt "$max_retries" ]; then
        echo "❌ FAIL. Server wasn't up after about 5s."

        kill "${pid}"
        if [ $? != 0 ]; then
            echo "Failed to stop the running instance. Return code: $? . Skipping tests."
            exit 8
        fi
        exit 16
    fi
    sleep 0.1s
done

echo "Startup log"
cat ./test/otelcol.log
