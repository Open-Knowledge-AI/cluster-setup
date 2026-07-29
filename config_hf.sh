# Set HuggingFace cache to shared directory
mkdir -p ~/.cache/huggingface
cat > ~/.cache/huggingface/.env <<EOF
HF_DATASETS_CACHE=/shared/huggingface/datasets
TRANSFORMERS_CACHE=/shared/huggingface/cache
HUGGINGFACE_HUB_CACHE=/shared/huggingface/cache
EOF

# Create global configuration
sudo tee /etc/profile.d/huggingface.sh <<EOF
export HF_DATASETS_CACHE=/shared/huggingface/datasets
export TRANSFORMERS_CACHE=/shared/huggingface/cache
export HUGGINGFACE_HUB_CACHE=/shared/huggingface/cache
EOF

# Make it executable
sudo chmod +x /etc/profile.d/huggingface.sh

# 6. Install huggingface-cli globally using UV tool
echo "Installing huggingface-cli globally..."
uv tool install huggingface-hub

# 7. Verify installation
echo "Verifying installation..."
huggingface-cli --version

# 8. Login (optional - uncomment if you want to authenticate)
# huggingface-cli login

echo "✅ HuggingFace setup complete!"
echo "CLI installed at: $(which huggingface-cli)"
echo "Cache directory: $HF_HOME"
