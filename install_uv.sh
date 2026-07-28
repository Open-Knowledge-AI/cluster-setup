# Install UV
curl -LsSf https://astral.sh/uv/install.sh | sudo sh

sudo mv /root/.local/bin/uv /usr/local/bin/
sudo chown root:devtools /usr/local/bin/uv
sudo chmod 775 /usr/local/bin/uv

# Create shared Python packages directory
sudo mkdir -p /shared/tools/python-packages
sudo chown -R root:devtools /shared/tools/python-packages
sudo chmod -R 775 /shared/tools/python-packages

# Configure UV to use shared cache and package directory
sudo tee /etc/profile.d/uv-shared.sh <<'EOF'
export UV_CACHE_DIR=/shared/tools/python-packages/uv-cache
export UV_TOOL_DIR=/shared/tools/python-packages/uv-tools
export UV_PYTHON_INSTALL_DIR=/shared/tools/python-packages/python-versions
export PIP_CACHE_DIR=/shared/tools/python-packages/pip-cache
export PIP_TARGET=/shared/tools/python-packages/site-packages
export PYTHONUSERBASE=/shared/tools/python-packages/user-base
export PATH=$PATH:$PYTHONUSERBASE/bin
EOF

sudo chmod +x /etc/profile.d/uv-shared.sh

# Create UV configuration file
mkdir -p /etc/uv
sudo tee /etc/uv/uv.toml <<'EOF'
[cache]
dir = "/shared/tools/python-packages/uv-cache"

[python]
install-dir = "/shared/tools/python-packages/python-versions"

[tool]
dir = "/shared/tools/python-packages/uv-tools"

[install]
target = "/shared/tools/python-packages/site-packages"
EOF
