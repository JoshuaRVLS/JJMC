#!/bin/bash

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root (sudo ./install_service.sh)"
  exit 1
fi

# Get the directory where the script is located (assuming it's the project root)
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
USER_NAME=${SUDO_USER:-$USER}
GROUP_NAME=$(id -gn $USER_NAME)

SERVICE_FILE="/etc/systemd/system/jjmc.service"

echo "Installing JJMC Service..."
echo "  Directory: $SCRIPT_DIR"
echo "  User:      $USER_NAME"

# Build the binary first to ensure it exists
echo "Building JJMC..."
cd "$SCRIPT_DIR"
# Ensure go is in path if running as sudo? 
# Sudo might reset path. Try absolute path or just hope env is preserved or go is in /usr/local/bin
if command -v go &> /dev/null; then
    go build -o jjmc
else
    echo "Warning: 'go' not found in PATH. Assuming 'jjmc' binary already exists."
fi

if [ ! -f "$SCRIPT_DIR/jjmc" ]; then
    echo "Error: jjmc binary not found. Please build it first."
    exit 1
fi

# Create Service File
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=JJMC Minecraft Panel
After=network.target

[Service]
Type=simple
User=$USER_NAME
Group=$GROUP_NAME
WorkingDirectory=$SCRIPT_DIR
ExecStart=$SCRIPT_DIR/jjmc
Restart=on-failure
Environment=PORT=3000

[Install]
WantedBy=multi-user.target
EOF

# Reload Systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable and Start
echo "Enabling and starting jjmc.service..."
systemctl enable jjmc
systemctl restart jjmc

echo "------------------------------------------------"
echo "Service Installed Successfully!"
echo "Check status: sudo systemctl status jjmc"
echo "View logs:    sudo journalctl -u jjmc -f"
echo "------------------------------------------------"
