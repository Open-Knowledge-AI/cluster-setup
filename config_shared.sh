# Create shared directories
sudo mkdir -p /shared/{projects,models,huggingface,datasets,tools}
sudo mkdir -p /shared/huggingface/{cache,datasets}

# Set appropriate permissions
sudo chown root:shared-users /shared
sudo chmod 775 /shared
sudo chown root:ml-users /shared/models /shared/huggingface /shared/datasets
sudo chmod 775 /shared/models /shared/huggingface /shared/datasets
sudo chown root:devtools /shared/tools
sudo chmod 775 /shared/tools

# Set SGID to ensure new files inherit group ownership
sudo chmod g+s /shared /shared/models /shared/huggingface /shared/datasets /shared/tools
