# Node OPC UA Test Server

A Node.js OPC UA server built with [node-opcua](https://github.com/node-opcua/node-opcua), used for development and integration testing.

**Difference from MS OPC-PLC:** This server supports the `Aes256_Sha256_RsaPss` security policy, which the MS OPC-PLC containers do not. Use this server when testing that specific policy.

**Supported security policies:** None, Basic256Sha256, Aes128_Sha256_RsaOaep, Aes256_Sha256_RsaPss

**Supported security modes:** None, Sign, SignAndEncrypt

**Authentication:** Anonymous; Username/Password (`user1/password1`, `admin/admin123`)

Run via `docker-compose.full.yaml` as the `opcua-node` service on port 4840.
