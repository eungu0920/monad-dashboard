#!/bin/bash

# Monad Dashboard Deployment Script
# Usage: ./deploy.sh <server-ip> <username>

set -e

SERVER_IP=${1:-"your-server-ip"}
USERNAME=${2:-"ubuntu"}
REMOTE_PATH="/opt/monad"

echo "🚀 Deploying Monad Dashboard to ${USERNAME}@${SERVER_IP}"

# Build locally
echo "📦 Building dashboard..."
make clean
make build

if [ ! -f "./monad-dashboard" ]; then
    echo "❌ Build failed - monad-dashboard binary not found"
    exit 1
fi

echo "✅ Build completed successfully"

# Create remote directory
echo "📁 Setting up remote directory..."
ssh ${USERNAME}@${SERVER_IP} "sudo mkdir -p ${REMOTE_PATH} && sudo chown ${USERNAME}:${USERNAME} ${REMOTE_PATH}"

# Transfer binary
echo "📤 Transferring binary..."
scp ./monad-dashboard ${USERNAME}@${SERVER_IP}:${REMOTE_PATH}/

# Transfer config if exists
if [ -f "./config.toml" ]; then
    echo "📤 Transferring config..."
    scp ./config.toml ${USERNAME}@${SERVER_IP}:${REMOTE_PATH}/
fi

# Make binary executable
ssh ${USERNAME}@${SERVER_IP} "chmod +x ${REMOTE_PATH}/monad-dashboard"

echo "✅ Deployment completed!"
echo ""
echo "To start the dashboard on the server:"
echo "ssh ${USERNAME}@${SERVER_IP}"
echo "cd ${REMOTE_PATH}"
echo "./monad-dashboard"
echo ""
echo "Dashboard will be available at: http://${SERVER_IP}:8080"