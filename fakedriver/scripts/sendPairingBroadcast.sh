#!/usr/bin/env bash

payload()
{
    local duration=$1
    cat <<EOF
{"params":{"duration": $duration},"jsonrpc":"2.0","time":1413180973519}
EOF
}

topic()
{
    echo "\$device/fake-phone/channel/user-agent/event/pairing-requested"
}

transport()
{
    local topic=$1
    mosquitto_pub -h ${DEVKIT_HOST:-localhost} -t "${topic}" -s
}

pipeline()
{
    payload "$@" | transport $(topic)
}

pipeline "$@"