#!/bin/bash

max_retries=50

podman run \
    --rm \
    --detach \
    --name jaeger \
    -p 14250 \
    -p 14269 \
    jaegertracing/all-in-one:1.20.0 > /dev/null

retries=0
while true
do
    curl -sL http://localhost:14269 > /dev/null 2>&1 && break
    sleep 0.1s
    let "retries++"
    if [ "$retries" -gt "$max_retries" ]; then
        echo "❌ ERROR ${test}. Jaeger wasn't up after about 5s."

        podman stop jaeger
        if [ $? != 0 ]; then
            echo "Failed to stop the running container."
            exit 1
        fi
        exit 2
    fi
done

echo "✅ Jaeger started."
