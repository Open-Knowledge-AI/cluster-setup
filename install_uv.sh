# Clear uv's internal cache
uv cache clean || true

# Delete uv-managed Python versions
rm -r "$(uv python dir)" || true

# Delete uv-managed global tools
rm -r "$(uv tool dir)" || true

rm -rf ~/.config/uv || true
rm -rf ~/.cache/uv || true

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
cache-dir = "/shared/tools/python-packages/uv-cache"
concurrent-downloads = 8
concurrent-builds = 4
concurrent-installs = 4
index-url = "https://pypi.org/simple"
resolution = "highest"
prerelease = "if-necessary"
compile-bytecode = false
link-mode = "clone"
EOF
