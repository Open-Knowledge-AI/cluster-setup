# Install Node.js (LTS)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# Install Go
wget https://go.dev/dl/go1.22.5.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.22.5.linux-arm64.tar.gz
rm go1.22.5.linux-arm64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh

# Add to setup-cluster.sh after Go installation
# Create shared Go workspace
sudo mkdir -p /shared/tools/go/bin
sudo chown -R root:devtools /shared/tools/go
sudo chmod -R 775 /shared/tools/go

# Set Go environment variables globally
sudo tee /etc/profile.d/go-shared.sh <<'EOF'
export GOROOT=/usr/local/go
export GOPATH=/shared/tools/go
export GOBIN=/shared/tools/go/bin
export PATH=$PATH:$GOROOT/bin:$GOBIN
EOF

# Make it executable
sudo chmod +x /etc/profile.d/go-shared.sh

# Set permissions for the go directory
sudo chmod g+s /shared/tools/go
