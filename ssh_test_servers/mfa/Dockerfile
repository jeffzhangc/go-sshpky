# ssh_test_servers/mfa/Dockerfile
FROM ubuntu:22.04

# Install OpenSSH server, sudo, and Google Authenticator PAM module
RUN apt-get update && apt-get install -y openssh-server sudo libpam-google-authenticator

# Create a user for testing
RUN useradd -rm -d /home/testuser -s /bin/bash -g root -G sudo -u 1000 testuser
RUN echo 'testuser:testpass' | chpasswd

# Setup Google Authenticator for the user
COPY .google_authenticator /home/testuser/.google_authenticator
RUN chown testuser:root /home/testuser/.google_authenticator
RUN chmod 400 /home/testuser/.google_authenticator

# SSH server configuration
COPY sshd_config /etc/ssh/sshd_config

# PAM configuration for SSHD
RUN echo "@include common-auth" > /etc/pam.d/sshd
RUN echo "auth required pam_google_authenticator.so" >> /etc/pam.d/sshd

# Create sshd host keys
RUN ssh-keygen -A

RUN mkdir -p /run/sshd

# Expose port 22
EXPOSE 22

# Start SSH server
CMD ["/usr/sbin/sshd", "-D"]
