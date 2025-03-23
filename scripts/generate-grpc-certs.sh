mkdir -p docker/ssl/certs
cd docker

if [ ! -f ssl/certs/ca.key ] || [ ! -f ssl/certs/ca.crt ]; then
    echo "Generating new CA certificates..."
    openssl genrsa -out ssl/certs/ca.key 4096
    openssl req -new -x509 -days 365 -key ssl/certs/ca.key -out ssl/certs/ca.crt -subj "/C=VN/ST=Hanoi/L=Hanoi/O=Mitras/OU=Mitras/CN=Mitras Root CA"
else
    echo "Using existing CA certificates..."
fi

generate_cert() {
    local name=$1
    local type=$2
    local cn="$3"

    openssl genrsa -out "ssl/certs/${name}-grpc-${type}.key" 4096

    openssl req -new \
        -key "ssl/certs/${name}-grpc-${type}.key" \
        -out "ssl/certs/${name}-grpc-${type}.csr" \
        -subj "/C=VN/ST=Hanoi/L=Hanoi/O=Mitras/OU=Mitras/CN=${cn}"

    cat > "ssl/certs/${name}-grpc-${type}.ext" << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = ${name}
EOF

    openssl x509 -req \
        -in "ssl/certs/${name}-grpc-${type}.csr" \
        -CA ssl/certs/ca.crt \
        -CAkey ssl/certs/ca.key \
        -CAcreateserial \
        -out "ssl/certs/${name}-grpc-${type}.crt" \
        -days 365 \
        -extfile "ssl/certs/${name}-grpc-${type}.ext"

    rm "ssl/certs/${name}-grpc-${type}.csr" "ssl/certs/${name}-grpc-${type}.ext"
}

# Generate server certificates
generate_cert "auth" "server" "auth.mitras.local"
generate_cert "groups" "server" "groups.mitras.local"
generate_cert "channels" "server" "channels.mitras.local"
generate_cert "clients" "server" "clients.mitras.local"

# Generate client certificates
generate_cert "auth" "client" "auth-client.mitras.local"
generate_cert "domains" "client" "domains-client.mitras.local"
generate_cert "groups" "client" "groups-client.mitras.local"
generate_cert "channels" "client" "channels-client.mitras.local"
generate_cert "clients" "client" "clients-client.mitras.local"

cd ssl/certs
chmod 644 *.crt
chmod 600 *.key

for service in auth groups channels clients domains; do
    ln -sf ca.crt "${service}-grpc-server-ca.crt"
    ln -sf ca.crt "${service}-grpc-client-ca.crt"
done

echo "Certificates generated successfully in docker/ssl/certs/"