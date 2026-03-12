# Testing with Docker

This document provides instructions on how to set up and use the Docker-based testing environment for `go-sshpky`. This allows you to test the three main authentication methods against live SSH servers.

## Prerequisites

*   [Docker](https://www.docker.com/get-started)
*   [Docker Compose](https://docs.docker.com/compose/install/)

## 1. Start the Test Environment

The testing environment consists of three pre-configured SSH servers, each running in a Docker container.

Navigate to the `ssh_test_servers` directory and start the services in detached mode:

```bash
cd ssh_test_servers
docker-compose up -d
```

This will start three SSH servers on your local machine:
*   **Password Authentication**: `localhost:2222`
*   **Public Key Authentication**: `localhost:2223`
*   **MFA (Password + TOTP) Authentication**: `localhost:2224`

## 2. Running Test Scenarios

For the following scenarios, it's recommended to create a dedicated group in `go-sshpky` to keep your test configurations separate.

```bash
# Create and use a 'docker-test' group
sshpky mg add docker-test
sshpky mg use docker-test
```

### Scenario 1: Password Authentication

This server accepts a simple username and password.

1.  **Add the host using `go-sshpky`**:
    ```bash
    sshpky ms add
    ```

2.  **Enter the following details when prompted**:
    *   **Host Name**: `t_password`
    *   **User@Host**: `testuser@127.0.0.1`
    *   **Port**: `2222`
    *   **Authentication Method**: Choose `Password`.
    *   **Password**: `testpass`

3.  **Connect to the host**:
    ```bash
    sshpky conn t_password
    ```

### Scenario 2: Public Key Authentication

This server is configured to accept the pre-generated private key located in the `ssh_test_servers` directory.

1.  **Add the host using `go-sshpky`**:
    ```bash
    sshpky ms add
    ```

2.  **Enter the following details when prompted**:
    *   **Host Name**: `t_publickey`
    *   **User@Host**: `testuser@127.0.0.1`
    *   **Port**: `2223`
    *   **Authentication Method**: Choose `Private Key`.
    *   **IdentityFile Path**: Provide the absolute path to `ssh_test_servers/test_id_rsa`.

3.  **Connect to the host**:
    ```bash
    sshpky conn t_publickey
    ```

### Scenario 3: MFA (Password + TOTP) Authentication

This server requires both a password and a Time-based One-Time Password (TOTP).

1.  **Add the host using `go-sshpky`**:
    ```bash
    sshpky ms add
    ```

2.  **Enter the following details when prompted**:
    *   **Host Name**: `t_mfa`
    *   **User@Host**: `testuser@127.0.0.1`
    *   **Port**: `2224`
    *   **Authentication Method**: Choose `Password`.
    *   **Password**: `testpass`
    *   **MFA Secret**: `35LAMSO4H77WFN25`

3.  **Connect to the host**:
    ```bash
    sshpky conn t_mfa
    ```
    `go-sshpky` will automatically generate the TOTP code from the secret and use it to authenticate.

## 3. Stop the Test Environment

Once you are finished with testing, you can stop and remove the containers.

Navigate to the `ssh_test_servers` directory and run:
```bash
cd ssh_test_servers
docker-compose down
```
