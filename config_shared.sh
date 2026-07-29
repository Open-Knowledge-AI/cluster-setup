# Create shared directories
# sudo mkdir -p /shared/{projects,models,huggingface,datasets,tools}
# sudo mkdir -p /shared/huggingface/{cache,datasets}

sudo chown -R root:ml-users /shared/huggingface
sudo chown -R root:ml-users /shared/datasets
sudo chown -R root:devtools /shared/tools
sudo chown -R root:shared-users /shared/projects

# Set SGID to ensure new files inherit group ownership
sudo chmod g+s /shared /shared/models /shared/huggingface /shared/datasets /shared/tools

sudo find /shared/huggingface /shared/datasets -type d -exec chmod 2775 {} \;
sudo find /shared/tools /shared/projects -type d -exec chmod 2775 {} \;

sudo find /shared/huggingface /shared/datasets /shared/tools /shared/projects -type f -exec chmod 664 {} \;
