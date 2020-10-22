#!/bin/bash

## setup
echo "Setting up..."
for st in ./test/start-jaeger.sh ./test/start-dex.sh
do
    ./${st}
    if [ $? != 0 ]; then
        exit $?
    fi
done

## test
echo "Testing..."
./test/start-otelcol.sh
if [ $? != 0 ]; then
    echo "failed: $?"
    exit $?
fi

## tear down
echo "Tearing down..."
for st in ./test/stop-jaeger.sh ./test/stop-dex.sh
do
    ./${st}
    if [ $? != 0 ]; then
        exit $?
    fi
done
