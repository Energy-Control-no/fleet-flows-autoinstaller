
#!/bin/bash
# Check if Airtable API key is provided as an argument
if [ -z "$1" ]; then
    echo "Error: No Airtable API key provided."
    echo "Usage: $0 <AIRTABLE_API_KEY>"
    exit 1
fi

# Constants
GIT_SERVER="ssh://git@fleet-flow-git.lizzardsolutions.com/home/git/git"
AIRTABLE_API_KEY="$1"
AIRTABLE_BASE_ID="appYWVOaoPhQB0nmA"
AIRTABLE_TABLE_NAME="Unipi"
HOSTNAME=$(hostname)
SSH_KEY_PATH="$HOME/.ssh/id_rsa"
BRANCH="main"
# Update package lists
sudo apt update

# Check and install required packages
ensure_installed() {
    if ! command -v $1 &> /dev/null; then
        echo "Installing $1..."
        sudo apt install -y $1
    else
        echo "$1 is already installed."
    fi
}
# Check if 'n' is installed
if ! command -v n >/dev/null 2>&1; then
    echo "n is not installed. Installing n..."
    # Install n (Node.js version manager)
    curl -L https://raw.githubusercontent.com/tj/n/master/bin/n -o n
    sudo bash n lts
    # Ensure the n command is available
    PATH="$PATH"
fi

# Update Node.js to the latest version using 'n'
sudo n latest

# Update npm to the latest version
sudo npm install -g npm@latest

echo "Node.js and npm are updated to the latest versions."
# Install Node.js, npm, and Node-RED
ensure_installed inotify-tools
ensure_installed node
ensure_installed npm
ensure_installed git
ensure_installed jq
ensure_installed nano

if ! command -v node-red &> /dev/null; then
    bash <(curl -sL https://raw.githubusercontent.com/node-red/linux-installers/master/deb/update-nodejs-and-nodered)
fi

# Function to update SSH key in Airtable
update_ssh_key_in_airtable() {
    local pubkey=$(cat $SSH_KEY_PATH.pub)
    local record_id=$(fetch_airtable_record_id_by_hostname "$HOSTNAME")
    echo "Fetched record id ?${record_id}?"
    if [ -n "$record_id" ]; then
        # Update existing record
        echo "Updating Existing Record"
        update_airtable_record "$record_id" "$pubkey"
    else
        echo "Create new Record"
        # Create new record
        create_airtable_record "$HOSTNAME" "$pubkey"
    fi
}

fetch_airtable_record_id_by_hostname() {
    local hostname=$1
    # Encoding the formula: {Device id} = 'hostname'
    local encodedFormula=$(printf "{Device id} = '%s'" "$hostname" | jq -sRr @uri)

    local response=$(curl -X GET \
        "https://api.airtable.com/v0/${AIRTABLE_BASE_ID}/${AIRTABLE_TABLE_NAME}?filterByFormula=${encodedFormula}" \
        -H "Authorization: Bearer ${AIRTABLE_API_KEY}" \
        -H "Content-Type: application/json")

  

    echo $(echo $response | jq -r '.records[0].id // empty')
}
# Function to update Airtable record with SSH Public Key
update_airtable_record() {
    local record_id=$1
    local ssh_key=$2
    local data=$(jq -n \
                    --arg sshKey "$ssh_key" \
                    '{fields: {"SSH Public Key": $sshKey}}')
    echo " record id ${record_id}"
    echo \n   
    local response=$(curl -X PATCH \
        "https://api.airtable.com/v0/${AIRTABLE_BASE_ID}/${AIRTABLE_TABLE_NAME}/${record_id}" \
        -H "Authorization: Bearer ${AIRTABLE_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "$data")

    echo "Response from Airtable:"
    echo "$response"
}

# Function to create new Airtable record with Device id and SSH Public Key
create_airtable_record() {
    local hostname=$1
    local ssh_key=$2

    # Split the hostname into type and SN
    local type="${hostname%-sn*}"      # Extracts the part before "-sn"
    local sn="${hostname##*-sn}"       # Extracts the part after "-sn"

    # Construct the JSON data
    local data=$(jq -n \
                    --arg type "$type" \
                    --arg sn "$sn" \
                    --arg sshKey "$ssh_key" \
                    '{records: [{fields: {"type": $type, "Unipi SN": ($sn | tonumber), "SSH Public Key": $sshKey}}]}')

    local response=$(curl -X POST \
        "https://api.airtable.com/v0/${AIRTABLE_BASE_ID}/${AIRTABLE_TABLE_NAME}" \
        -H "Authorization: Bearer ${AIRTABLE_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "$data")

    echo "Response from Airtable:"
    echo "$response"
}

# Function to check Git server access
check_git_access() {
    ssh -o BatchMode=yes -T $GIT_SERVER 2>&1 | grep -q "successfully authenticated"
}


# Check and generate SSH key
if [ ! -f "$SSH_KEY_PATH" ]; then
    echo "Generating new SSH key..."
    ssh-keygen -t rsa -b 4096 -f $SSH_KEY_PATH -N ""
fi

# Check Git access and update Airtable if necessary
if ! check_git_access; then
    echo "Git server access failed. Updating SSH key in Airtable..."
    update_ssh_key_in_airtable
fi

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

}

create_systemd_service

create_auto_update_job() {
    AUTO_UPDATER_SCRIPT="/usr/local/bin/auto_updater_ffjs.sh"

    sudo tee $AUTO_UPDATER_SCRIPT > /dev/null <<'EOL'
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

restart_on_changes() {
    RESTART_SCRIPT="/usr/local/bin/restart_change_ffjs.sh"
    LOG_FILE="/var/log/restart_change_ffjs.log"

    sudo tee $RESTART_SCRIPT > /dev/null <<'EOL'
#!/bin/bash

# Define the project directory
PROJECT_DIR="$HOME/fleet-flows-js"

# Log file location
LOG_FILE="$LOG_FILE"

# List of files to monitor within the project directory
FILES=("schema.yml" ".env")

# Function to monitor changes and restart the service
monitor_and_restart() {
    echo "Monitoring files for changes..." >> "$LOG_FILE"
    inotifywait -m -e modify -q "$PROJECT_DIR/schema.yml" "$PROJECT_DIR/.env" | while read -r line; do
        if systemctl is-active --quiet fleet-flow-js.service; then
            echo "File change detected: $line" >> "$LOG_FILE"
            echo "Restarting the fleet-flow-js service..." >> "$LOG_FILE"
            sudo systemctl restart fleet-flow-js.service >> "$LOG_FILE" 2>&1
        else
            echo "File change detected, but fleet-flow-js.service is not running." >> "$LOG_FILE"
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

create_fleet_flow_js_listener_service() {
    SERVICE_FILE="/etc/systemd/system/fleet-flow-js-listener.service"

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
    sudo systemctl enable fleet-flow-js-listener
    sudo systemctl start fleet-flow-js-listener

    echo "Fleet Flow JS Listener Service is setup and started."
}
create_fleet_flow_js_listener_service
