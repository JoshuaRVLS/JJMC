<div align="center">
  <img src="website/logo.png" alt="JJMC Logo" width="200"/>
  <h1>JJMC</h1>
</div>

A simple and modern Minecraft Server Manager.

## Features

### 🖥️ Dashboard & Instance Management
- **Create Instantly:** Deploy servers in seconds with support for Vanilla, Fabric, Forge, NeoForge, Quilt, and Paper.
- **Import Existing:** Easily import existing server folders.
- **Server Versions:** Automatic fetching of latest Minecraft and Loader versions.
- **Java Manager:** Built-in manager to download, install, and switch between multiple Java runtimes (8, 11, 17, 21) effortlessly.

### ⚙️ advanced Control
- **Live Console:** Real-time server console with color support and command input.
- **Process Management:** Start, stop, restart, and kill server processes.
- **Resource Monitoring:** Real-time CPU and Memory usage tracking.
- **Crash Detection:** Automatic restart on server crash.

### 📂 File System
- **File Manager:** Built-in web-based file explorer.
- **Editor:** Edit configuration files directly in the browser.
- **Archives:** Compress and decompress files/folders.
- **Upload/Download:** Easy file transfer.

### 🧩 Mods & Plugins
- **Mod Manager:** Search and install mods directly from Modrinth and CurseForge (planned).
- **Plugin Manager:** Manage plugins for Paper/Spigot servers.
- **Modpacks:** One-click install for popular modpacks.

### 📅 Scheduling & Automation
- **Task Scheduler:** Schedule recurring commands (e.g., "say Hello", "stop", "backup").
- **CRON Support:** Flexible scheduling options.

### 🛡️ Security & Backups
- **Backups:** Create and restore server backups.
- **Authentication:** Secure login system.

## How to run
Simply run the script for your OS:
- **Linux**: `./run.sh`
- **Windows**: `run.bat` or `run.ps1`

The script will automatically set up the portable Go and Node.js toolchains if needed.

## Advanced Usage & Docker
For instructions on running with **Docker**, or using **RCON**, **Telnet**, and **SFTP**, please see the [Tutorial & Usage Guide](TUTORIAL.md).

## License
MIT
