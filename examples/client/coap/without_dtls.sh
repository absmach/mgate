#!/bin/bash
protocol=coap
host=localhost
port=5682
path="test"
content=0x32
message="{\"message\": \"Hello mProxy\"}"
auth="TOKEN"

echo "Posting message to ${protocol}://${host}:${port}/${path} without tls ..."
coap-client -m post coap://${host}:${port}/${path} -e "${message}" -O 12,${content} -O 15,auth=${auth}

echo "Getting message from ${protocol}://${host}:${port}/${path} without tls ..."
coap-client -m get coap://${host}:${port}/${path} -O 6,0x00 -O 15,auth=${auth} -
