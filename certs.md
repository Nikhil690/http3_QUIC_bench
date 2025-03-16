This is a guide that explains how to generate self-signed certificates and test TLS communication between two machines (server and client) using OpenSSL. Adjust any details or file paths for your environment.

---

# TLS Communication Test using OpenSSL

This guide demonstrates how to generate and use self-signed certificates for a simple TLS test between two machines:

- **Server**: 192.168.6.239  
- **Client**: 192.168.6.232  

We will:

1. Create a Certificate Authority (CA).  
2. Create and sign the server certificate.  
3. Create and sign the client certificate.  
4. Test TLS communication using OpenSSL’s built-in testing tools.  

---

## Prerequisites

- **OpenSSL** installed on each machine.  
- Basic knowledge of Linux command-line usage.  
- Proper network connectivity between the two machines on the desired port.  

---

## Step 1: Create a Certificate Authority (CA)

1. **Generate the CA private key** (e.g., 2048 bits):

   ```bash
   openssl genrsa -out ca.key 2048
   ```

2. **Generate a self-signed CA certificate** (valid for 365 days, adjust as needed):

   ```bash
   openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 \
       -out ca.crt \
       -subj "/C=US/ST=State/L=City/O=MyOrg/OU=MyOrgUnit/CN=MyTestCA"
   ```

   - **ca.key**: Keep this private and secure.
   - **ca.crt**: Public certificate for the CA.

---

## Step 2: Create and Sign the Server Certificate

Perform these steps on (or for) the **server** (192.168.6.239).

1. **Generate the server key**:

   ```bash
   openssl genrsa -out server.key 2048
   ```

2. **Generate a CSR (Certificate Signing Request)**:

   ```bash
   openssl req -new -key server.key -out server.csr \
       -subj "/C=US/ST=State/L=City/O=MyOrg/OU=MyOrgUnit/CN=192.168.6.239"
   ```

   - The `CN` should match the server’s hostname or IP. Here we use its IP address.

3. **Sign the server CSR with the CA**:

   ```bash
   openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
       -out server.crt -days 365 -sha256
   ```

   - **server.crt** is now a valid certificate signed by your CA.

   Copy `server.key`, `server.crt`, and `ca.crt` to the server (192.168.6.239).

---

## Step 3: Create and Sign the Client Certificate

Perform these steps on (or for) the **client** (192.168.6.232).

1. **Generate the client key**:

   ```bash
   openssl genrsa -out client.key 2048
   ```

2. **Generate a CSR**:

   ```bash
   openssl req -new -key client.key -out client.csr \
       -subj "/C=US/ST=State/L=City/O=MyOrg/OU=MyOrgUnit/CN=192.168.6.232"
   ```

3. **Sign the client CSR with the CA**:

   ```bash
   openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
       -out client.crt -days 365 -sha256
   ```

   Copy `client.key`, `client.crt`, and `ca.crt` to the client (192.168.6.232).

---

## Step 4: Test the TLS Connection

### 4.1 Start the Server

On **192.168.6.239**, run:

```bash
openssl s_server \
  -cert server.crt \
  -key server.key \
  -CAfile ca.crt \
  -Verify 1 \
  -accept 4433
```

- `-cert server.crt` / `-key server.key`: The server’s TLS certificate and key.  
- `-CAfile ca.crt`: Tells the server to trust this CA.  
- `-Verify 1`: Requests (and requires) the client certificate for mutual TLS.  
- `-accept 4433`: Listen on port 4433 (pick any unused port).

### 4.2 Connect from the Client

On **192.168.6.232**, run:

```bash
openssl s_client \
  -connect 192.168.6.239:4433 \
  -cert client.crt \
  -key client.key \
  -CAfile ca.crt
```

- `-connect 192.168.6.239:4433`: Points to the server’s IP and the port used in `s_server`.  
- `-cert client.crt` / `-key client.key`: The client’s TLS credentials.  
- `-CAfile ca.crt`: The CA certificate to trust.

If the handshake succeeds, you’ll see output indicating a successful TLS connection. You can type text in the client terminal to send it to the server.

---

## Troubleshooting

1. **Check Firewall**: Ensure port 4433 (or whichever you used) is open and accessible.  
2. **File Permissions**: Your private keys should be accessible only by their owner (e.g., `chmod 600 server.key`).  
3. **IP vs Hostname**: If you used an IP address for the Common Name (`CN`), ensure you connect using that IP.  
4. **Ensure CA is Trusted**: Both server and client need the `ca.crt` to establish trust.  

---

## Summary

- **CA Generation**: Create `ca.key` and `ca.crt`.  
- **Server**: Generate `server.key`, create `server.csr`, sign it with `ca.crt` → produces `server.crt`.  
- **Client**: Generate `client.key`, create `client.csr`, sign it with `ca.crt` → produces `client.crt`.  
- **Run a Test**:  
  - `openssl s_server` on the server with `server.key`, `server.crt`, and `ca.crt`.  
  - `openssl s_client` on the client with `client.key`, `client.crt`, and `ca.crt`.  

This allows you to confirm that the certificates work and that TLS is established securely between client and server.