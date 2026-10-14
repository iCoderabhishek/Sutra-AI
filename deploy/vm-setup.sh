#!/usr/bin/env bash
# One-time setup for a fresh Debian or Ubuntu VM. Run it once over SSH:
#   curl -fsSL https://raw.githubusercontent.com/iCoderabhishek/Sutra-AI/production/deploy/vm-setup.sh | bash
# Then log out and back in so the docker group applies.
set -euo pipefail

REPO="https://github.com/iCoderabhishek/Sutra-AI.git"
DIR="/opt/sutra"

echo "==> Installing Docker"
if ! command -v docker >/dev/null; then
  curl -fsSL https://get.docker.com | sh
fi
sudo usermod -aG docker "$USER"

echo "==> Adding 2 GB swap (image builds can run out of memory on small VMs)"
if ! swapon --show | grep -q /swapfile; then
  sudo fallocate -l 2G /swapfile
  sudo chmod 600 /swapfile
  sudo mkswap /swapfile
  sudo swapon /swapfile
  echo "/swapfile none swap sw 0 0" | sudo tee -a /etc/fstab >/dev/null
fi

echo "==> Cloning the repository to $DIR"
sudo apt-get update -qq
sudo apt-get install -y git >/dev/null
sudo mkdir -p "$DIR"
sudo chown "$USER": "$DIR"
if [ ! -d "$DIR/.git" ]; then
  git clone "$REPO" "$DIR"
fi

echo "Done. Log out and back in, then run the Deploy workflow from GitHub."
