# Copyright (c) Abstract Machines
# SPDX-License-Identifier: Apache-2.0

CRT_LOCATION = certs
O = Magistrala
OU_CA = magistrala_ca
OU_CRT = magistrala_crt
EA = info@magistrala.com
CN_CA = Magistrala_Self_Signed_CA
CN_SRV = localhost
CLINET_CN = mProxy_client
CRT_FILE_NAME = thing

define CERT_CONFIG
[req]
req_extensions = v3_req
distinguished_name = dn
prompt = no

[dn]
CN = <<COMMON_NAME>>
C  = RS
ST = RS
L  = BELGRADE
O  = MAGISTRALA
OU = MAGISTRALA

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = <<COMMON_NAME>>
endef

define ANNOUNCE_BODY
Version $(VERSION) of $(PACKAGE_NAME) has been released.

It can be downloaded from $(DOWNLOAD_URL).

etc, etc.
endef
all: clean_certs ca server_cert thing_cert thing_cert_revoked thing_cert_unknown

# CA name and key is "ca".
ca:
	openssl req -newkey rsa:2048 -x509 -nodes -sha512 -days 1095 \
				-keyout $(CRT_LOCATION)/ca.key -out $(CRT_LOCATION)/ca.crt -subj "/CN=$(CN_CA)/O=$(O)/OU=$(OU_CA)/emailAddress=$(EA)"

# Server cert and key name is "server".
server_cert:
	# Create Conf file
	$(file > $(CRT_LOCATION)/server.conf,$(subst <<COMMON_NAME>>,$(CN_SRV),$(CERT_CONFIG)) )

	# Create magistrala server key and CSR.
	openssl req -new -sha256 -newkey rsa:4096 -nodes \
				-keyout $(CRT_LOCATION)/server.key \
				-out $(CRT_LOCATION)/server.csr \
				-subj "/CN=$(CN_SRV)/O=$(O)/OU=$(OU_CRT)/emailAddress=$(EA)" \
				-config $(CRT_LOCATION)/server.conf \
				-extensions v3_req

	# Sign server CSR.
	openssl x509 -req -days 1000 -in $(CRT_LOCATION)/server.csr  \
				-CA $(CRT_LOCATION)/ca.crt \
				-CAkey $(CRT_LOCATION)/ca.key \
				-CAcreateserial \
				-out $(CRT_LOCATION)/server.crt \
				-extfile $(CRT_LOCATION)/server.conf \
				-extensions v3_req

	# Remove CSR.
	rm $(CRT_LOCATION)/server.csr $(CRT_LOCATION)/server.conf


thing_cert:
	# Create magistrala server key and CSR.
	openssl req -new -sha256 -newkey rsa:4096 -nodes -keyout $(CRT_LOCATION)/$(CRT_FILE_NAME).key \
				-out $(CRT_LOCATION)/$(CRT_FILE_NAME).csr -subj "/CN=$(CLINET_CN)/O=$(O)/OU=$(OU_CRT)/emailAddress=$(EA)"

	# Sign client CSR.
	openssl x509 -req -days 730 -in $(CRT_LOCATION)/$(CRT_FILE_NAME).csr -CA $(CRT_LOCATION)/ca.crt -CAkey $(CRT_LOCATION)/ca.key -CAcreateserial -out $(CRT_LOCATION)/$(CRT_FILE_NAME).crt

	# Remove CSR.
	rm $(CRT_LOCATION)/$(CRT_FILE_NAME).csr

thing_cert_revoked:
	# Create magistrala server key and CSR.
	openssl req -new -sha256 -newkey rsa:4096 -nodes -keyout $(CRT_LOCATION)/$(CRT_FILE_NAME)_revoked.key \
				-out $(CRT_LOCATION)/$(CRT_FILE_NAME)_revoked.csr -subj "/CN=$(CLINET_CN)/O=$(O)/OU=$(OU_CRT)/emailAddress=$(EA)"

	# Sign client CSR.
	openssl x509 -req -days 730 -in $(CRT_LOCATION)/$(CRT_FILE_NAME)_revoked.csr -CA $(CRT_LOCATION)/ca.crt -CAkey $(CRT_LOCATION)/ca.key -CAcreateserial -out $(CRT_LOCATION)/$(CRT_FILE_NAME)_revoked.crt

	# Remove CSR.
	rm $(CRT_LOCATION)/$(CRT_FILE_NAME)_revoked.csr

thing_cert_unknown:
	# Create magistrala server key and CSR.
	openssl req -new -sha256 -newkey rsa:4096 -nodes -keyout $(CRT_LOCATION)/$(CRT_FILE_NAME)_unknown.key \
				-out $(CRT_LOCATION)/$(CRT_FILE_NAME)_unknown.csr -subj "/CN=$(CLINET_CN)/O=$(O)/OU=$(OU_CRT)/emailAddress=$(EA)"

	# Sign client CSR.
	openssl x509 -req -days 730 -in $(CRT_LOCATION)/$(CRT_FILE_NAME)_unknown.csr -CA $(CRT_LOCATION)/ca.crt -CAkey $(CRT_LOCATION)/ca.key -CAcreateserial -out $(CRT_LOCATION)/$(CRT_FILE_NAME)_unknown.crt

	# Remove CSR.
	rm $(CRT_LOCATION)/$(CRT_FILE_NAME)_unknown.csr

clean_certs:
	rm -r $(CRT_LOCATION)/*.crt
	rm -r $(CRT_LOCATION)/*.key
