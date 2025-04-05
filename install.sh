#!/bin/bash

set -e

INSTALL_DIR="/usr/local/bin"
TARGET="$INSTALL_DIR/gs"

if [ ! -f "./gs" ]; then
    echo "Error: 'gs' binary not found in the current directory."
    exit 1
fi

if [ -f "$TARGET" ]; then
    echo "An existing installation of gs was found at $TARGET."
    read -p "Do you want to overwrite it? (y/n): " response
    case "$response" in
      [yY][eE][sS]|[yY])
        echo "Overwriting existing installation..."
        ;;
      *)
        echo "Installation aborted."
        exit 0
        ;;
    esac
fi

if [ "$(id -u)" -ne 0 ]; then
    echo "Installing as root using sudo..."
    sudo cp ./gs "$TARGET"
    sudo chmod +x "$TARGET"
else
    cp ./gs "$TARGET"
    chmod +x "$TARGET"
fi

echo "Installation complete. You can now run the tool by typing 'gs' in your terminal."
