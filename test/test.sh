#!/bin/bash

# register the teardown function before we can use it in the trap
function teardown {
    ## tear down
    echo "Tearing down..."
    for st in ./test/stop-otelcol.sh ./test/stop-jaeger.sh ./test/stop-dex.sh
    do
        ./${st}
    done
}

## setup
echo "Setting up..."
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
echo "Starting Observatorium OpenTelemetry Collector distribution..."
./test/start-otelcol.sh
if [ $? != 0 ]; then
    exit $?
fi

## generate a trace
echo "Generating trace..."
./test/generate-trace.sh
if [ $? != 0 ]; then
    exit $?
fi

## check that a trace exists in Jaeger
echo "Checking for existence of a trace..."
./test/check-trace.sh
exit $?
