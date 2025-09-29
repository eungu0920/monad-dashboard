#!/bin/bash

# Ubuntu Server Setup Script for Monad Dashboard
# Run this on your Ubuntu server

set -e

echo "🔧 Setting up Monad Dashboard on Ubuntu Server"

# Update system
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo "📦 Installing Go..."
    cd /tmp
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    echo "✅ Go installed successfully"
fi

# Install Node.js if not present
if ! command -v node &> /dev/null; then
    echo "📦 Installing Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    sudo apt-get install -y nodejs
    echo "✅ Node.js installed successfully"
fi

# Install git and make if not present
sudo apt install -y git make wget curl

# Create monad directory
sudo mkdir -p /opt/monad
sudo chown $USER:$USER /opt/monad
cd /opt/monad

echo "✅ Server setup completed!"
echo ""
echo "Next steps:"
echo "1. Clone or transfer your monad-dashboard code to /opt/monad/"
echo "2. Run: make install && make build"
echo "3. Configure dashboard to connect to your Monad nodes"