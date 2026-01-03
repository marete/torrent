#!/bin/bash
#
# Set up CI container environment.
# This script installs dependencies needed in the Ubuntu 24.04 container.
#
# Usage: ./scripts/ci-setup.sh
#

set -e

export DEBIAN_FRONTEND=noninteractive

GO_VERSION="${GO_VERSION:-1.24.0}"

echo "=== Updating package lists ==="
apt-get update

echo "=== Installing base dependencies ==="
apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    g++ \
    gcc \
    git \
    libc6-dev

echo "=== Installing Go ${GO_VERSION} ==="
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz
export PATH="/usr/local/go/bin:${HOME}/go/bin:${PATH}"

echo "=== Go version: $(go version) ==="
echo "=== Container setup complete ==="
