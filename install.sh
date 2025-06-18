#!/usr/bin/env bash

set -e

# Configuration
REPO="AlexsanderHamir/prof"
API_URL="https://api.github.com/repos/$REPO/releases/latest"
INSTALL_DIR="$HOME/.prof"
VENV_DIR="$INSTALL_DIR/venv"
WRAPPER_PATH="$HOME/bin/prof"

echo "Installing prof from the latest GitHub release..."

# Check and fix bin directory permissions if needed
if [ -d "$HOME/bin" ]; then
    if [ ! -O "$HOME/bin" ]; then
        echo "Fixing permissions for $HOME/bin directory..."
        if command -v sudo >/dev/null 2>&1; then
            sudo chown -R "$USER:$(id -gn)" "$HOME/bin" || {
                echo "Error: Could not fix permissions. Please run: sudo chown -R $USER:$(id -gn) $HOME/bin"
                exit 1
            }
        else
            echo "Error: sudo command not found. Please run: chown -R $USER:$(id -gn) $HOME/bin as root"
            exit 1
        fi
    fi
fi

# 1. Download and extract the latest release tarball
if [ -d "$INSTALL_DIR" ]; then
    echo "Removing existing installation at $INSTALL_DIR..."
    rm -rf "$INSTALL_DIR"
fi
mkdir -p "$INSTALL_DIR"

# Fetch tarball_url and tag_name from GitHub API
RELEASE_INFO=$(curl -s $API_URL)
TARBALL_URL=$(echo "$RELEASE_INFO" | grep tarball_url | cut -d '"' -f 4)
TAG_NAME=$(echo "$RELEASE_INFO" | grep tag_name | cut -d '"' -f 4)
if [ -z "$TARBALL_URL" ]; then
    echo "Error: Could not fetch latest release tarball URL."
    exit 1
fi

echo "Latest release tag: $TAG_NAME"
echo "Downloading and extracting latest release..."
curl -L "$TARBALL_URL" | tar -xz -C "$INSTALL_DIR" --strip-components=1

# 2. Create virtual environment if not exists
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv "$VENV_DIR"
fi

# 3. Activate venv and install dependencies
echo "Installing Python dependencies..."
source "$VENV_DIR/bin/activate"
pip install --upgrade pip
if [ -f "$INSTALL_DIR/requirements.txt" ]; then
    pip install -r "$INSTALL_DIR/requirements.txt"
else
    echo "Warning: requirements.txt not found!"
fi
deactivate

# 4. Create bin directory if needed and check permissions
echo "Setting up bin directory..."
if [ ! -d "$HOME/bin" ]; then
    mkdir -p "$HOME/bin" || { echo "Error: Could not create $HOME/bin directory"; exit 1; }
fi

# Check if we have write permissions to the bin directory
if [ ! -w "$HOME/bin" ]; then
    echo "Error: No write permission for $HOME/bin directory"
    echo "Please run: chmod u+w $HOME/bin"
    exit 1
fi

# 5. Create wrapper script
echo "Creating wrapper script at $WRAPPER_PATH"
if [ -f "$WRAPPER_PATH" ]; then
    rm "$WRAPPER_PATH" || { echo "Error: Could not remove existing wrapper script"; exit 1; }
fi

cat > "$WRAPPER_PATH" << EOF
#!/usr/bin/env bash
source "$VENV_DIR/bin/activate"
python "$INSTALL_DIR/prof" "\$@"
EOF

if [ ! -f "$WRAPPER_PATH" ]; then
    echo "Error: Failed to create wrapper script at $WRAPPER_PATH"
    exit 1
fi

chmod +x "$WRAPPER_PATH" || { echo "Error: Could not make wrapper script executable"; exit 1; }

# 6. Print instructions for adding to PATH
echo ""
echo "┌──────────────────────────────────────────────────────────────────────────┐"
echo "│                         IMPORTANT: PATH CONFIGURATION                    │"
echo "├──────────────────────────────────────────────────────────────────────────┤"
echo "│                                                                          │"
echo "│  To use the 'prof' command from anywhere, you need to add this line to   │"
echo "│  your shell configuration file (.zshrc, .bashrc, etc.):                  │"
echo "│                                                                          │"
echo "│    export PATH=\"\$HOME/bin:\$PATH\"                                     │"
echo "│                                                                          │"
echo "│  Then either:                                                            │"
echo "│    • Restart your terminal, or                                           │"
echo "│    • Run: source ~/.zshrc (or your shell's config file)                  │"
echo "│                                                                          │"
echo "└──────────────────────────────────────────────────────────────────────────┘"
echo ""

echo "Installation complete! You can now run 'prof' from your terminal after"
echo "completing the PATH configuration steps above."
