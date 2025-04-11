# Certs Service

Issues certificates for clients. `Certs` service can create certificates to be used when `Mitras` is deployed to support mTLS.
Certificate service can create certificates using PKI mode - where certificates issued by PKI, when you deploy `Vault` as PKI certificate management `cert` service will proxy requests to `Vault` previously checking access rights and saving info on successfully created certificate.
