#!/bin/bash

max_retries=50
retries=0
while true
do
    traces=$(curl -s "http://localhost:16686/api/traces?service=e2e-test-agent-with-auth")
    numTraces=$(echo ${traces} | jq -r '.data' | jq length)
    numSpans=$(echo ${traces}  | jq -r '.data[0].spans' | jq length)

    # we expect at least one trace and at least two spans in the first trace
    if (( numTraces > 0 )); then
        echo "found at least one trace for the e2e-test service" >> ./test/test.log
        if (( numSpans > 1 )); then
            echo "found at least two spans in the trace" >> ./test/test.log
            echo "✅ Traces found in the backend for the authenticated agent."
            break
        else
            echo "not enough spans in the trace: found ${numSpans}, expected at least 2" >> ./test/test.log
        fi
    else
        echo "no traces found for the e2e-test service" >> ./test/test.log
    fi

    let "retries++"
    if [ "$retries" -gt "$max_retries" ]; then
        echo "❌ FAIL. Could not find the traces/spans within a reasonable time."
        exit 1
    fi
    sleep 0.1s
done

traces=$(curl -s "http://localhost:16686/api/traces?service=e2e-test-agent-without-auth")
numTraces=$(echo ${traces} | jq -r '.data' | jq length)
if (( numTraces > 0 )); then
    echo "❌ FAIL. Expected to not find traces for non-authenticated agents, but found ${numTraces}."
    exit 1
fi
