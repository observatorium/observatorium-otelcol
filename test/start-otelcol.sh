#!/bin/bash

max_retries=50

token=$(curl \
    --cacert ./test/certs/ca.pem \
    --silent \
    --request POST \
    --url https://127.0.0.1:5556/dex/token \
    --header 'content-type: application/x-www-form-urlencoded' \
    --data grant_type=password \
    --data username=admin@example.com \
    --data password=password \
    --data client_id=test \
    --data client_secret=ZXhhbXBsZS1hcHAtc2VjcmV0 \
    --data scope="openid email" | sed 's/^{.*"id_token":[^"]*"\([^"]*\)".*}/\1/'
)

sed "s/bearer_token\:.*/bearer_token: ${token}/" test/collector.example.yaml > test/collector.yaml
if [ $? != 0 ]; then
    echo "❌ FAIL. Could not set the bearer token in the collector's configuration"
    failed=true
    exit 2
fi

# start the distribution
./_build/observatorium-otelcol --config ./test/collector.yaml > ./test/otelcol.log 2>&1 &
pid=$!

retries=0
while true
do
    kill -0 "${pid}" >/dev/null 2>&1
    if [ $? != 0 ]; then
        echo "❌ FAIL. The Observatorium OpenTelemetry Collector isn't running. Startup log:"
        failed=true
        exit 1
    fi

    curl -s localhost:13133 | grep "Server available" > /dev/null
    if [ $? == 0 ]; then
        echo "✅ Observatorium OpenTelemetry Collector started."
        echo "${pid}" > otelcol.pid
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
