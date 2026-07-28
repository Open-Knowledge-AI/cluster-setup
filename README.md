# Welcome to the Cluster, ${USER}!

## System Overview
This is a shared computing cluster with centralized tools and data storage. Here's everything you need to know:

### Shared Directories
- **/shared/projects** - Project workspace for collaborative development
- **/shared/models** - Shared ML models and weights
- **/shared/huggingface** - HuggingFace cache directory
- **/shared/datasets** - Shared datasets
- **/shared/tools** - Shared development tools

### Available Tools
All tools are available system-wide:

#### Container Runtimes
- **Docker** - Container management
- **Podman** - Rootless container runtime
- **kubectl** - Kubernetes cluster management

#### Development Tools
- **Python** - Miniconda3 installation at /shared/tools/miniconda3
- **Mamba** - Fast conda package manager
- **UV** - Python package installer
- **Node.js** - JavaScript runtime
- **Go** - Go programming language

### HuggingFace Configuration
The HuggingFace cache is automatically configured to use the shared directory:
- HF_HOME=/shared/huggingface
- HF_DATASETS_CACHE=/shared/huggingface/datasets
- TRANSFORMERS_CACHE=/shared/huggingface/cache
- HUGGINGFACE_HUB_CACHE=/shared/huggingface/cache

### Your Groups
You have been added to the following groups:
  - docker
  - podman
  - kubectl
  - devtools
  - ml-users
  - shared-users
  - sudo


### Storage Quotas
- **Home Directory**: 15GB maximum
- **Shared Directories**: No quota (but please be considerate)

### Best Practices
1. Store large files in /shared/ directories
2. Keep your home directory clean (under 15GB)
3. Use the HuggingFace cache for model files
4. When using containers, mount shared directories as volumes
5. Be mindful of resource usage

### Quick Commands
```bash
# Check your quota
quota -s

# View shared directories
ls -la /shared/

# Activate conda environment
source /shared/tools/miniconda3/etc/profile.d/conda.sh

# Use mamba
mamba create -n myenv python=3.11
mamba activate myenv

# Check Docker status
docker info

# Check Podman status
podman info
```

### Getting Help
- For tool issues: Check the tool's documentation
- For cluster issues: Contact the cluster administrator
- For quota issues: Use the shared directories for large files

### Important Notes
- ⚠️  You have sudo access - use it responsibly
- Log out and log back in for group changes to take effect
- Use shared directories for collaborative work
- Report any issues to the administrator

Welcome aboard and happy computing!
