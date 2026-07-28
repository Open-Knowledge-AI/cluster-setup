# Create groups
sudo groupadd -r docker
sudo groupadd -r podman
sudo groupadd -r kubectl
sudo groupadd -r devtools
sudo groupadd -r ml-users
sudo groupadd -r shared-users

# Add your user to groups (replace $USER with your actual username)
sudo usermod -aG docker,devtools,ml-users,shared-users $USER

# Verify groups
groups $USER
