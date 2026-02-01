# JJMC Tutorial & Usage Guide

This guide covers how to run JJMC using Docker and how to use its remote management features (RCON, Telnet, SFTP).

## Docker

JJMC can be easily run using Docker. This ensures all dependencies (like Java versions) are isolated.

### Prerequisites

- Docker
- Docker Compose

### Quick Start

We provide a helper script to get you started quickly:

```bash
./run_docker.sh
```

This script will build the Docker image and start the container in the background.

### Manual Usage

You can also use standard Docker Compose commands:

```bash
# Start in background
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Configuration

The `docker-compose.yml` file defines:
- **Ports**: Maps container ports to your host.
- **Volumes**: Persists data for instances, servers, backups, and the database.

> **Note**: By default, the web interface listens on port used in configuration (default 3000 or 3001). Check `docker-compose.yml` for the current mapping.

---

## Remote Management Protocols

JJMC exposes several protocols to manage your instances remotely.

### 1. RCON (Remote RCON Proxy)

JJMC provides a unified RCON server that proxies commands to specific instances.

- **Port**: `2024` (Default)
- **Authentication**: `password#instanceID`
  - `password`: Your global JJMC login password.
  - `instanceID`: The ID of the instance you want to control.

#### Usage Example (using `mcrcon` or similar tools)

Configure your RCON client to connect to `localhost:2024`.
For the password, use the format: `MySecretPass#instance-123`

This allows you to send commands like `/op player` or `/stop` directly to `instance-123` through the JJMC proxy.

### 2. Telnet Console

For a more interactive experience, you can use Telnet to attach to an instance's console.

- **Port**: `2023` (Default)

#### Usage

1. Open a terminal.
2. Run `telnet localhost 2023`.
3. You will be prompted for your **Password**. Enter your global JJMC password.
4. You will be prompted for the **Instance ID**. Enter the ID of the running instance.
5. You are now attached to the console! Type commands directly.
6. Type `/exit` to disconnect.

```text
$ telnet localhost 2023
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
JJMC Telnet Console
====================
Password: ****
Authenticated.
Enter Instance ID: my-server
Attached to My Server (my-server). Type /exit to disconnect.
> list
[Server] There are 0 of 20 players online: 
```

### 3. SFTP (File Management)

You can manage your instance files using any SFTP client (like FileZilla, WinSCP, or the `sftp` command).

- **Port**: `2022` (Default)
- **User**: (Any username is accepted, e.g., `admin`)
- **Password**: Your global JJMC login password.
- **Directory**: The root directory is the `instances` folder.

#### Usage

Connect using your preferred SFTP client:
- **Host**: `your-server-ip`
- **Port**: `2022`
- **Protocol**: SFTP (SSH File Transfer Protocol)

You will see folders for each of your instances. You can upload mods, download logs, or edit configs directly.
