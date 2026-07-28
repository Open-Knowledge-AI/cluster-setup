# Set HuggingFace cache to shared directory
mkdir -p ~/.cache/huggingface
cat > ~/.cache/huggingface/.env <<EOF
HF_HOME=/shared/huggingface
HF_DATASETS_CACHE=/shared/huggingface/datasets
TRANSFORMERS_CACHE=/shared/huggingface/cache
HUGGINGFACE_HUB_CACHE=/shared/huggingface/cache
EOF

# Create global configuration
sudo tee /etc/profile.d/huggingface.sh <<EOF
export HF_HOME=/shared/huggingface
export HF_DATASETS_CACHE=/shared/huggingface/datasets
export TRANSFORMERS_CACHE=/shared/huggingface/cache
export HUGGINGFACE_HUB_CACHE=/shared/huggingface/cache
EOF

# Make it executable
sudo chmod +x /etc/profile.d/huggingface.sh
