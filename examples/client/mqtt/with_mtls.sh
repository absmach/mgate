#!/bin/bash

topic="test/topic"
message="Hello mProxy"
port=8884
host=localhost
cafile=ssl/certs/ca.crt
certfile=ssl/certs/thing.crt
keyfile=ssl/certs/thing.key
reovokedcertfile=ssl/certs/thing_revoked.crt
reovokedkeyfile=ssl/certs/thing_revoked.key
unknowncertfile=ssl/certs/thing_unknown.crt
unknownkeyfile=ssl/certs/thing_unknown.key

echo "Subscribing to topic ${topic} with mTLS certificate ${cafile} ${certfile} ${keyfile}..."
mosquitto_sub -h $host -p $port -t $topic --cafile $cafile --cert $certfile --key $keyfile &
sub_pid=$!
sleep 1

cleanup() {
    echo "Cleaning up..."
    kill $sub_pid
}

trap cleanup EXIT

echo "Publishing to topic ${topic} with mTLS, with ca certificate ${cafile} and with client certificate ${certfile} ${keyfile}..."
mosquitto_pub -h $host -p $port -t $topic  -m "${message}"  --cafile $cafile --cert $certfile  --key $keyfile
sleep 1

echo "Publishing to topic ${topic} with mTLS, with ca certificate ${cafile} and with client revoked certificate ${reovokedcertfile} ${reovokedkeyfile}..."
mosquitto_pub -h $host -p $port -t $topic  -m "${message}" --cafile $cafile --cert $reovokedcertfile --key $reovokedkeyfile  2>&1
sleep 1

echo "Publishing to topic ${topic} with mTLS, with ca certificate ${cafile} and with client unknown certificate ${unknowncertfile} ${unknownkeyfile}..."
mosquitto_pub -h $host -p $port -t $topic  -m "${message}"  --cafile $cafile --cert $unknowncertfile --key $unknownkeyfile  2>&1
sleep 1

echo "Publishing to topic ${topic} with mTLS, with ca certificate ${cafile} and without any clinet certificate ...."
mosquitto_pub -h $host -p $port -t $topic  -m "${message}" --cafile $cafile   2>&1
sleep 1

echo "Publishing to topic ${topic} without mTLS, without any certificate ...."
mosquitto_pub -h $host -p $port -t $topic  -m "${message}"  2>&1
sleep 1
