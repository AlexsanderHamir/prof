#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_message() {
    echo -e "${GREEN}[prof]${NC} $1"
}

print_error() {
    echo -e "${RED}[prof] Error:${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[prof] Warning:${NC} $1"
}

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    print_error "Docker daemon is not running. Please start Docker first."
    exit 1
fi

# Create necessary directories
print_message "Setting up directories..."
mkdir -p ~/.prof/prompts

# Create default config if it doesn't exist
if [ ! -f ~/.prof/config.json ]; then
    print_message "Creating default config file..."
    cat > ~/.prof/config.json << EOL
{
    "api_key": "",
    "base_url": "https://api.openai.com/v1",
    "model_config": {
        "model": "gpt-4-turbo-preview",
        "max_tokens": 4096,
        "temperature": 0.7,
        "top_p": 1.0
    },
    "benchmark_configs": {}
}
EOL
    print_warning "Please update ~/.prof/config.json with your API key and other settings"
fi

# Pull the Docker image
print_message "Pulling Docker image..."
docker pull alexsanderhamir/prof:latest

# Create the alias command
ALIAS_CMD='alias prof="docker run --rm -v \"$HOME/.prof:/home/profuser/.prof\" -v \"$(pwd):/workspace\" -w /workspace -e PROF_CONFIG=\"$HOME/.prof/config.json\" -e PROF_PROMPT=\"$HOME/.prof/prompts\" alexsanderhamir/prof:latest"'

# Detect shell and add alias
SHELL_RC=""
if [[ "$SHELL" == *"zsh"* ]]; then
    SHELL_RC="$HOME/.zshrc"
elif [[ "$SHELL" == *"bash"* ]]; then
    SHELL_RC="$HOME/.bashrc"
    # Also check for .bash_profile on macOS
    if [[ -f "$HOME/.bash_profile" ]]; then
        SHELL_RC="$HOME/.bash_profile"
    fi
fi

if [ -n "$SHELL_RC" ]; then
    # Check if alias already exists
    if ! grep -q "alias prof=" "$SHELL_RC"; then
        print_message "Adding prof alias to $SHELL_RC..."
        echo -e "\n# prof alias for Docker installation" >> "$SHELL_RC"
        echo "$ALIAS_CMD" >> "$SHELL_RC"
    else
        print_warning "prof alias already exists in $SHELL_RC"
    fi
else
    print_warning "Could not detect shell configuration file. Please add this alias manually:"
    echo "$ALIAS_CMD"
fi

# Test the installation
print_message "Testing installation..."
if docker run --rm alexsanderhamir/prof:latest &> /dev/null; then
    print_message "Installation successful! 🎉"
    print_message "To start using prof, either:"
    print_message "1. Restart your terminal, or"
    print_message "2. Run: source $SHELL_RC"
else
    print_error "Installation test failed. Please check the Docker image pull was successful."
    exit 1
fi 