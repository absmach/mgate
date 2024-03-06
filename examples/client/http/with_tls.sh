#!/bin/bash
protocol=https
host=localhost
port=8087
path="messages"
content="application/json"
message="{\"message\": \"Hello mProxy\"}"
invalidPath="invalid_path"
cafile=ssl/certs/ca.crt
certfile=ssl/certs/thing.crt
keyfile=ssl/certs/thing.key
reovokedcertfile=ssl/certs/thing_revoked.crt
reovokedkeyfile=ssl/certs/thing_revoked.key
unknowncertfile=ssl/certs/thing_unknown.crt
unknownkeyfile=ssl/certs/thing_unknown.key

echo "Posting message to ${protocol}://${host}:${port}/${path} with tls, Authorization header, ca certificate ${cafile}..."
curl -sSiX POST "${protocol}://${host}:${port}/${path}" -H "content-type:${content}" -H "Authorization:TOKEN" -d "${message}" --cacert $cafile


echo -e "\nPosting message to ${protocol}://${host}:${port}/${path} with tls, basic authentication ca certificate ${cafile}...."
curl -sSi -u username:password -X  POST "${protocol}://${host}:${port}/${path}" -H "content-type:${content}" -d "${message}"  --cacert $cafile

echo -e "\nPosting message to invalid path ${protocol}://${host}:${port}/${invalidPath} with tls, Authorization header, ca certificate ${cafile}..."
curl -sSiX POST "${protocol}://${host}:${port}/${invalidPath}" -H "content-type:${content}" -H "Authorization:TOKEN" -d "${message}" --cacert $cafile

echo -e "\nPosting message to ${protocol}://${host}:${port}/${path} with tls, Authorization header, and without ca certificate.."
curl -sSiX POST "${protocol}://${host}:${port}/${invalidPath}" -H "content-type:${content}" -H "Authorization:TOKEN" -d "${message}" 2>&1

