#!/bin/bash
# Constants
GIT_SERVER="ssh://git@fleet-flow-git.lizzardsolutions.com/home/git/git"
BRANCH="main"
SSH_KEY_PATH="$HOME/.ssh/id_rsa"

# Function to clone Git repositories
clone_repository() {
    local repo_name=$1
    local branch=$2
    until git clone --single-branch --branch ${branch} ${GIT_SERVER}/${repo_name}.git; do
        echo "Git clone of $repo_name failed. Retrying..."
        sleep 5
    done
}

# Clone repositories
clone_repository "fleet-files" "main"
clone_repository "fleet-flows-js" ${BRANCH}

# Run npm install in fleet-flows-js
cd fleet-flows-js
npm install
cd ..

# Create .env file
create_env_file() {
    local env_file_path="$1/.env"
    cat <<EOL > $env_file_path
LOCAL_REPO_PATH=$HOME/fleet-files
MASTER_BRANCH=main
FLOW_FILE_PATH=$HOME/.node-red/flows.json
FLOWS_DIR=$HOME/fleet-files/flows
SUBFLOWS_DIR=$HOME/fleet-files/subflows
RETRY_TIME=5
SCHEMA_FILE_PATH=$1/schema.yml
NODE_RED_DIRECTORY=$HOME/
CONFIGS_DIR=$HOME/fleet-files/config
RESTART_COMMAND='sudo systemctl restart node-red.service'
EOL
}

create_env_file "$(realpath ./fleet-flows-js)"

# Create a systemd service for fleet-flows-js
create_systemd_service() {
    local service_path="/etc/systemd/system/fleet-flows-js.service"
    sudo bash -c "cat > $service_path" <<EOL
[Unit]
Description=OTA flow updates and flow compiling
After=network.target

[Service]
ExecStart=/usr/bin/npm run start
WorkingDirectory=$(realpath ./fleet-flows-js)
Restart=always
User=$(whoami)

[Install]
WantedBy=multi-user.target
EOL

    sudo systemctl enable fleet-flows-js.service
# disabled , we do not have schema file upon install.    sudo systemctl start fleet-flows-js.service
}

create_systemd_service

# Create auto update job
create_auto_update_job() {
    AUTO_UPDATER_SCRIPT="/usr/local/bin/auto_updater_ffjs.sh"

    sudo tee $AUTO_UPDATER_SCRIPT > /dev/null <<EOL
#!/bin/bash

# Set the project directory and backup directory
PROJECT_DIR=$HOME/fleet-flows-js
BACKUP_DIR=$HOME/fleet-flows-js.backup
REMOTE_REPO=${GIT_SERVER}/fleet-flows-js.git
FILES_TO_BACKUP=("schema.yml" ".env") # Add other files as needed
BRANCH=${BRANCH}
if [ -d "$PROJECT_DIR" ] && [ -d "$PROJECT_DIR/.git" ]; then
    cd $PROJECT_DIR
    echo "Current Directory: $(pwd)"

    # Fetch the latest commits from the remote
    git fetch

    # Compare local HEAD and remote HEAD
    LOCAL_SHA=$(git rev-parse HEAD)
    REMOTE_SHA=$(git rev-parse origin/main)

    if [ "$LOCAL_SHA" != "$REMOTE_SHA" ]; then
        echo "Updating the project..."

        # Backup specified files
        echo "Backing up files..."
        mkdir -p $BACKUP_DIR
        for file in "${FILES_TO_BACKUP[@]}"; do
            if [ -f "$file" ]; then
                cp $file $BACKUP_DIR/
            else
                echo "Warning: File $file not found for backup."
            fi
        done

        # Delete the project directory
        echo "Deleting old project directory..."
        cd ..
        rm -rf $PROJECT_DIR

        # Clone the remote repository
        echo "Cloning the remote repository..."
        git clone --single-branch --branch $BRANCH $REMOTE_REPO $PROJECT_DIR

        # Restore the backed-up files
        echo "Restoring files..."
        cd $PROJECT_DIR
        npm i 
        for file in "${FILES_TO_BACKUP[@]}"; do
            if [ -f "$BACKUP_DIR/$file" ]; then
                cp $BACKUP_DIR/$file .
            else
                echo "Warning: Backup of $file not found for restoration."
            fi
        done

        echo "Update complete."
    else
        echo "Project is already up to date."
    fi
else
    echo "Error: The directory $PROJECT_DIR is not a valid Git repository."
fi
EOL
    sudo chmod +rx $AUTO_UPDATER_SCRIPT
    LOG_FILE="/var/log/auto_updater_ffjs.log"
    # Setup the cronjob for auto-update
(crontab -l 2>/dev/null; echo "0 * * * * $AUTO_UPDATER_SCRIPT >> $LOG_FILE 2>&1") | crontab -
    echo "auto-update is setup $AUTO_UPDATER_SCRIPT"

}
create_auto_update_job

# Restart on changes
restart_on_changes() {
    RESTART_SCRIPT="/usr/local/bin/restart_change_ffjs.sh"
    LOG_FILE="/usr/local/bin/frestart_change_ffjs.log"

    sudo tee $RESTART_SCRIPT > /dev/null <<EOL
#!/bin/bash

# Define the project directory
PROJECT_DIR="$HOME/fleet-flows-js"

# List of files to monitor within the project directory
FILES=("schema.yml" ".env")

# Function to monitor changes and restart the service
monitor_and_restart() {
    echo "Monitoring files for changes..." >> "$LOG_FILE"
    inotifywait -m -e modify -q "${FILES[@]/#/$PROJECT_DIR/}" | while read -r line; do
        if systemctl is-active --quiet fleet-flows-js.service; then
            echo "File change detected: $line" >> "$LOG_FILE"
            echo "Restarting the fleet-flows-js service..." >> "$LOG_FILE"
            sudo systemctl restart fleet-flows-js.service >> "$LOG_FILE" 2>&1
        else
            echo "File change detected, but fleet-flows-js.service is not running." >> "$LOG_FILE"
        fi
    done
}

# Run the monitor function
monitor_and_restart

EOL
    sudo chmod +rx $RESTART_SCRIPT

    echo "Restart Script is setup $RESTART_SCRIPT"
}
restart_on_changes

# Create Fleet Flow JS Listener Service
create_fleet_flow_js_listener_service() {
    SERVICE_FILE="/etc/systemd/system/fleet-flows-js-listener.service"

    sudo tee $SERVICE_FILE > /dev/null <<EOL
[Unit]
Description=Fleet Flow JS File Change Listener
After=network.target

[Service]
Type=simple
User=$(whoami)
ExecStart=/usr/local/bin/restart_change_ffjs.sh
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOL

    sudo systemctl daemon-reload
    sudo systemctl enable fleet-flows-js-listener
    sudo systemctl start fleet-flows-js-listener

    echo "Fleet Flows JS Listener Service is setup and started. "
}
create_fleet_flow_js_listener_service