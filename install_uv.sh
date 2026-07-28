# Install UV
curl -LsSf https://astral.sh/uv/install.sh | sudo sh
sudo mv /root/.local/bin/uv /usr/local/bin/
sudo chown root:devtools /usr/local/bin/uv
sudo chmod 775 /usr/local/bin/uv
