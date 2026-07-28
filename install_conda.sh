# Create symlinks for system-wide access
sudo rm -rf /shared/tools/miniconda3/bin/conda /usr/local/bin/conda
sudo rm -rf /shared/tools/miniconda3/bin/mamba /usr/local/bin/mamba

sudo rm -rf /shared/tools/miniconda3

# Install Miniconda for aarch64
sudo wget https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-aarch64.sh
sudo bash Miniconda3-latest-Linux-aarch64.sh -u -b -p /shared/tools/miniconda3
sudo rm Miniconda3-latest-Linux-aarch64.sh

sudo /shared/tools/miniconda3/bin/conda tos accept --override-channels --channel https://repo.anaconda.com/pkgs/r
sudo /shared/tools/miniconda3/bin/conda tos accept --override-channels --channel https://repo.anaconda.com/pkgs/main

# Install Mamba
sudo /shared/tools/miniconda3/bin/conda install -n base -c conda-forge mamba -y

# Create symlinks for system-wide access
sudo ln -s /shared/tools/miniconda3/bin/conda /usr/local/bin/conda
sudo ln -s /shared/tools/miniconda3/bin/mamba /usr/local/bin/mamba

# Set permissions
sudo chown -R root:devtools /shared/tools/miniconda3
sudo chmod -R 775 /shared/tools/miniconda3

# add to env on login
sudo ln -s /shared/tools/miniconda3/etc/profile.d/conda.sh /etc/profile.d/conda.sh
