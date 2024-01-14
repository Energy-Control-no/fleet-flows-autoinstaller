#!/bin/bash

# Constants
GIT_CLONED_DIRS=("$HOME/fleet-files" "$HOME/fleet-flows-js")
SERVICE_FILES=("/etc/systemd/system/fleet-flows-js.service" "/etc/systemd/system/fleet-flows-js-listener.service")
SCRIPT_FILES=("/usr/local/bin/auto_updater_ffjs.sh" "/usr/local/bin/restart_change_ffjs.sh")
cd /home/unipi/
# Function to remove cloned repositories
remove_cloned_repos() {
    for dir in "${GIT_CLONED_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            echo "Removing cloned repository: $dir"
            rm -rf "$dir"
        else
            echo "Directory not found: $dir"
        fi
    done
}

# Function to remove systemd service files and disable services
remove_service_files() {
    for file in "${SERVICE_FILES[@]}"; do
        if [ -f "$file" ]; then
            echo "Disabling and removing service file: $file"
            local service_name=$(basename "$file" .service)
            sudo systemctl stop "$service_name"
            sudo systemctl disable "$service_name"
            sudo rm -f "$file"
        else
            echo "Service file not found: $file"
        fi
    done
    # Reload systemd daemon to apply changes
    sudo systemctl daemon-reload
}

# Function to remove script files
remove_script_files() {
    for file in "${SCRIPT_FILES[@]}"; do
        if [ -f "$file" ]; then
            echo "Removing script file: $file"
            sudo rm -f "$file"
        else
            echo "Script file not found: $file"
        fi
    done
}

# Remove cloned repositories
remove_cloned_repos

# Remove service files
remove_service_files

# Remove script files
remove_script_files

echo "fleet-flow-js removal process completed."
