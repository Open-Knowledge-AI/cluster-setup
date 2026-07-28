# Create audit rules file
sudo tee /etc/audit/rules.d/99-core-audit.rules <<EOF
# Clear existing rules
-D

# Buffer size
-b 8192

# Failure mode
-f 1

# Monitor sudo users file
-w /etc/sudoers -p wa -k sudo_users
-w /etc/sudoers.d/ -p wa -k sudo_users

# Monitor password changes
-w /etc/passwd -p wa -k password_changes
-w /etc/shadow -p wa -k password_changes
-w /etc/gshadow -p wa -k password_changes
-w /etc/group -p wa -k group_changes

# Monitor Docker
-w /usr/bin/docker -p x -k docker_exec
-w /var/run/docker.sock -p rw -k docker_socket
-w /etc/docker/ -p wa -k docker_config

# Monitor Podman
-w /usr/bin/podman -p x -k podman_exec
-w /etc/containers/ -p wa -k podman_config

# Monitor kubectl
-w /usr/local/bin/kubectl -p x -k kubectl_exec
-w /etc/kubernetes/ -p wa -k kube_config

# Monitor critical system files
-w /etc/ssh/sshd_config -p wa -k sshd_config
-w /etc/crontab -p wa -k crontab
-w /etc/cron.d/ -p wa -k cron_files

# Monitor user management
-w /usr/sbin/useradd -p x -k user_management
-w /usr/sbin/userdel -p x -k user_management
-w /usr/sbin/usermod -p x -k user_management
-w /usr/bin/passwd -p x -k user_management

# Monitor container events via syslog
-a always,exit -S connect -F path=/var/run/docker.sock -k docker_connect
-a always,exit -S connect -F path=/run/podman/podman.sock -k podman_connect

# Monitor package installations
-w /usr/bin/apt -p x -k package_install
-w /usr/bin/dpkg -p x -k package_install
EOF

# Apply audit rules
sudo augenrules --load

# Persist rules
sudo systemctl restart auditd
