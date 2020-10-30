#!/bin/bash

# we send first the no auth, so that we get higher assurance that the non-existence of the
# span is because of the auth failure instead of network latency
tracegen -otlp-endpoint localhost:55685 -otlp-insecure -service e2e-test-agent-without-auth &>> ./test/tracegen.log
if [ $? != 0 ]; then
    echo "Failed to generate a trace to the agent that does not authenticate with the collector."
    exit 1
fi

tracegen -otlp-endpoint localhost:55680 -otlp-insecure -service e2e-test-agent-with-auth &>> ./test/tracegen.log
if [ $? != 0 ]; then
    echo "Failed to generate a trace to the agent that authenticates with the collector."
    exit 1
fi

echo "✅ Traces generated."
