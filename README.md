# KCAdm

Kubernetes Admission controller

## Test locally

First create the required certificate files

**Note-1: Please update the domains to a domain you prefer**
**Note-2: If you want to use the code as is, then add 127.0.0.1 sysiops.com to your /etc/hosts**

### Create Selfsigned Certificates

Create certs directory

```bash
mkdir -vp certs
```

Create CA root key
```bash
openssl genrsa -out certs/ca.key 2048
```

Create CA certificate
```bash
openssl req -new -x509 \
    -key certs/ca.key \
    -days 365000 \
    -subj "/CN=sysiops.com/C=GR/ST=Attiki/L=Athens/O=sysiops/OU=sysiops" \
    -out certs/ca.crt
```

Create Server's (Admission controller server) private key
```bash
openssl genrsa -out certs/server-key.pem 2048
```

Create a server certificate signing request
```bash
openssl req -new \
    -key certs/server-key.pem \
    -subj "/CN=kcadm.sysiops.svc/C=GR/ST=Attiki/L=Athens/O=sysiops/OU=sysiops" \
    -addext "subjectAltName = DNS:kcadm.sysiops.com,DNS:sysiops.com,DNS:kcadm,DNS:kcadm.sysiops,DNS:kcadm.sysiops.svc,DNS:kcadm.sysiops.svc.cluster.local" \
    -out certs/server.csr 
```

Create the Server's certificate and sign it with the CA we created
```bash
openssl x509 -req \
    -extfile <(printf "subjectAltName=DNS:kcadm.sysiops.com,DNS:sysiops.com,DNS:kcadm,DNS:kcadm.sysiops,DNS:kcadm.sysiops.svc,DNS:kcadm.sysiops.svc.cluster.local") \
    -in certs/server.csr \
    -days 365000 \
    -CA certs/ca.crt \
    -CAkey certs/ca.key \
    -CAcreateserial \
    -out certs/server-cert.pem
```

### Run Server

```bash
go run main.go
```

### Test namespace resource

Issue an invalid request

```bash
curl -s  https://sysiops.com:8080/namespace-admition -d @invalid-namespace-name.json --cacert certs/ca.crt | jq
{
  "kind": "Namespace",
  "apiVersion": "v1",
  "response": {
    "uid": "705ab4f5-6393-11e8-b7cc-42010a800002",
    "allowed": false,
    "status": {
      "metadata": {},
      "message": "Namespace Name must start with user-",
      "code": 403
    }
  }
}
```

Issue a valid request
```bash
curl -s  https://sysiops.com:8080/namespace-admition -d @valid-namespace-name.json --cacert certs/ca.crt | jq
{
  "kind": "Namespace",
  "apiVersion": "v1",
  "response": {
    "uid": "705ab4f5-6393-11e8-b7cc-42010a800002",
    "allowed": true,
    "status": {
      "metadata": {}
    }
  }
}
```
