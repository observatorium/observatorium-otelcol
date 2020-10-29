#!/bin/bash

tracegen -otlp-endpoint localhost:55680 -otlp-insecure -service e2e-test
if [ $? != 0 ]; then
    echo "Failed to generate a trace."
    exit 1
fi

echo "✅ Trace generated."
