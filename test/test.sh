#!/bin/bash

# register the teardown function before we can use it in the trap
function teardown {
    podman logs dex > ./test/dex.log
    podman logs jaeger > ./test/jaeger.log

    ## tear down
    echo "🔧 Tearing down..."
    for st in ./test/stop-otelcol.sh ./test/stop-jaeger.sh ./test/stop-dex.sh
    do
        ./${st}
    done

    echo "🪵 dex logs"
    cat ./test/dex.log

    echo "🪵 Jaeger logs"
    cat ./test/jaeger.log

    echo "🪵 tracegen with auth logs"
    cat ./test/tracegen-auth.log

    echo "🪵 tracegen without auth logs"
    cat ./test/tracegen-noauth.log

    echo "🪵 Observatorium OpenTelemetry Collector distribution logs"
    cat ./test/otelcol.log

    echo "🪵 Test logs"
    cat ./test/test.log
}

## setup
echo "🔧 Setting up..."
for st in ./test/start-jaeger.sh ./test/start-dex.sh ./test/install-tracegen.sh
do
    ./${st}
    if [ $? != 0 ]; then
        exit $?
    fi
done

# from this point and on, we run the teardown before we exit
trap teardown EXIT

## test
echo "🔧 Starting Observatorium OpenTelemetry Collector distribution..."
./test/start-otelcol.sh
if [ $? != 0 ]; then
    exit $?
fi

## generate a trace
echo "🔧 Generating trace..."
./test/generate-trace.sh
if [ $? != 0 ]; then
    exit $?
fi

## check that a trace exists in Jaeger
echo "🔧 Checking for existence of a trace..."
./test/check-trace.sh
exit $?
