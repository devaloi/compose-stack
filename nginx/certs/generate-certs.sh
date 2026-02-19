#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="${SCRIPT_DIR}"

if [ -f "${CERT_DIR}/server.crt" ] && [ -f "${CERT_DIR}/server.key" ]; then
    echo "Certificates already exist in ${CERT_DIR}"
    echo "Remove them first if you want to regenerate."
    exit 0
fi

echo "Generating self-signed certificate..."

openssl req -x509 -nodes -days 365 \
    -newkey rsa:2048 \
    -keyout "${CERT_DIR}/server.key" \
    -out "${CERT_DIR}/server.crt" \
    -subj "/C=US/ST=State/L=City/O=Development/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

chmod 600 "${CERT_DIR}/server.key"
chmod 644 "${CERT_DIR}/server.crt"

echo "Certificate generated:"
echo "  Key:  ${CERT_DIR}/server.key"
echo "  Cert: ${CERT_DIR}/server.crt"
