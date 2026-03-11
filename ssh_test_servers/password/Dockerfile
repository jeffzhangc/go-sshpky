# ssh_test_servers/password/Dockerfile
FROM ubuntu:22.04

# Install OpenSSH server and sudo
RUN apt-get update && apt-get install -y openssh-server sudo

# Create a user for testing
RUN useradd -rm -d /home/testuser -s /bin/bash -g root -G sudo -u 1000 testuser
RUN echo 'testuser:testpass' | chpasswd

# SSH server configuration
COPY sshd_config /etc/ssh/sshd_config

# Create sshd host keys
RUN ssh-keygen -A

RUN mkdir -p /run/sshd

# Expose port 22
EXPOSE 22

# Start SSH server
CMD ["/usr/sbin/sshd", "-D"]
