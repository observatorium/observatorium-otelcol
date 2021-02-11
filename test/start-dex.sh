#!/bin/bash

max_retries=50

podman run \
    --detach \
    --rm \
    --name dex \
    -p 5556:5556 \
    -v ./test/dex.yaml:/etc/config/dex.yaml:z \
    -v ./test/certs/cert.pem:/etc/config/cert.pem:z \
    -v ./test/certs/cert-key.pem:/etc/config/cert-key.pem:z \
    dexidp/dex:v2.25.0 \
    serve \
    /etc/config/dex.yaml > /dev/null

retries=0
while true
do
    curl -sL https://localhost:5556/dex/.well-known/openid-configuration --cacert ./test/certs/ca.pem > /dev/null 2>&1 && break
    sleep 0.1s
    let "retries++"
    if [ "$retries" -gt "$max_retries" ]; then
        echo "❌ ERROR: dex wasn't up after about 5s."
        failed=true

        podman stop dex
        if [ $? != 0 ]; then
            echo "Failed to stop the running container."
            exit 1
        fi
        exit 2
    fi
done

echo "✅ dex started."
