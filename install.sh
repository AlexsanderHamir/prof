#!/usr/bin/env bash

set -e

# Configuration
REPO_URL="https://github.com/AlexsanderHamir/prof.git"
INSTALL_DIR="$HOME/.prof"
VENV_DIR="$INSTALL_DIR/venv"
WRAPPER_PATH="$HOME/bin/prof"

echo "Installing prof from GitHub..."

# 1. Clone or pull latest repo
if [ ! -d "$INSTALL_DIR" ]; then
    echo "Cloning repository into $INSTALL_DIR"
    git clone "$REPO_URL" "$INSTALL_DIR"
else
    echo "Repository already exists, pulling latest changes"
    git -C "$INSTALL_DIR" pull
fi

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

# 4. Create bin directory if needed
mkdir -p "$HOME/bin"

# 5. Create wrapper script
echo "Creating wrapper script at $WRAPPER_PATH"
cat > "$WRAPPER_PATH" << EOF
#!/usr/bin/env bash
source "$VENV_DIR/bin/activate"
python "$INSTALL_DIR/prof" "\$@"
EOF

chmod +x "$WRAPPER_PATH"

# 6. Print instructions for adding to PATH
echo "To use the 'prof' command from anywhere, add this line to your shell config file (.zshrc, .bashrc, etc.):"
echo 'export PATH="$HOME/bin:$PATH"'
echo "After adding this line, restart your terminal or run: source ~/.zshrc (or your shell's config file)"

echo "Installation complete! You can now run 'prof' from your terminal."
