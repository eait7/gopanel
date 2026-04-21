#!/bin/bash
# GoPanel — One-Command Installer
# Usage: curl -sL https://raw.githubusercontent.com/username/gopanel/main/install.sh | sudo bash

set -e

echo "🚀 Installing GoPanel Ecosystem..."

# Install dependencies if missing
if ! command -v git &> /dev/null; then
    echo "📦 Installing git..."
    apt-get update && apt-get install -y git
fi

if ! command -v docker &> /dev/null; then
    echo "📦 Installing docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
fi

if ! docker compose version &> /dev/null; then
    echo "📦 Installing docker-compose-plugin..."
    apt-get update && apt-get install -y docker-compose-plugin || apt-get install -y docker-compose-v2
fi

# Clone the repository
INSTALL_DIR="/opt/gopanel"
if [ -d "$INSTALL_DIR" ]; then
    echo "🔄 GoPanel is already installed. Pulling latest updates..."
    cd "$INSTALL_DIR"
    git pull
else
    echo "📥 Cloning repository..."
    git clone https://github.com/eait7/gopanel.git "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# Run setup
bash setup.sh
