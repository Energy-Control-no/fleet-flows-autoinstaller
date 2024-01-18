#!/bin/bash
# Check if Airtable API key is provided as an argument
debug=false
noupdate=true
AIRTABLE_API_KEY=""
cecho(){
    RED="\033[0;31m"
    GREEN="\033[0;32m"  # <-- [0 means not bold
    YELLOW="\033[1;33m" # <-- [1 means bold
    CYAN="\033[1;36m"
    Error="\033[0;31m"
    DEBUG="\033[0;31;47m"
    # ... Add more colors if you like

    NC="\033[0m" # No Color

    # printf "${(P)1}${2} ${NC}\n" # <-- zsh
    printf "${!1}${2} ${NC}\n" # <-- bash
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    -d|--debug)
      debug=true
      shift
      ;;
    -nou|--no-update)
      noupdate=false
      shift
      ;;
    -k|--key)
      AIRTABLE_API_KEY="$2"
      shift 2
      ;;
    *)
      echo "Error: Unsupported flag $1" >&2
      exit 1
      ;;
  esac
done


debug_echo() {
    if [ "$debug" = true ]; then
        cecho "$1" "$2"
    fi
}
debug_echo "DEBUG" "Debugging "


if [ -z "$AIRTABLE_API_KEY" ]; then
    cecho "RED" "Error: No Airtable API key provided."
    cecho "CYAN" "Usage: $0 -k <AIRTABLE_API_KEY>"
    exit 1
fi


# Constants
# Constants
GIT_SERVER="ssh://git@fleet-flows-git.lizzardsolutions.com"
AIRTABLE_BASE_ID="appYWVOaoPhQB0nmA"
AIRTABLE_TABLE_NAME="Unipi"
HOSTNAME=$(hostname)
SSH_KEY_PATH="$HOME/.ssh/id_rsa"
BRANCH="main"
# Update package lists

check_airtable_api_key() {
    debug_echo "DEBUG" "check_airtable_api_key called"
    # Making a test request to Airtable
    local response=$(curl -s -o /dev/null -w "%{http_code}" -X GET \
        "https://api.airtable.com/v0/${AIRTABLE_BASE_ID}/${AIRTABLE_TABLE_NAME}" \
        -H "Authorization: Bearer ${AIRTABLE_API_KEY}")

    # Check if the HTTP status code is 200 (OK)
    if [ "$response" -ne 200 ]; then

        cecho "RED" "Invalid API key. Stopping the script."
        exit 1
    else
        cecho "GREEN" "API key is valid."
        # Continue with the rest of the script
    fi
}

debug_echo "DEBUG" "noupdate flasg is: $noupdate"
if [ "$noupdate" = true ]; then
    debug_echo "DEBUG" "running update"
    sudo apt update
    debug_echo "DEBUG" "running upgrade"
    sudo apt upgrade
    fi

debug_echo "DEBUG" "calling check_airtable_api_key"
check_airtable_api_key 

# Check and install required packages
ensure_installed() {
    
    if ! command -v $1 &> /dev/null; then
        cecho "YELLOW" "Installing $1..."
        sudo apt install -y $1
    else
        cecho "YELLOW" "$1 is already installed."
    fi
}
debug_echo "DEBUG" "updating npm to the latest version"
# Update npm to the latest version

# Check if 'n' is installed


# Update Node.js to the latest version using 'n'
debug_echo "DEBUG" " installing specific version of node"



cecho "GREEN" "Node.js and npm are updated to the latest versions."

# Install Node.js, npm, and Node-RED
ensure_installed inotify-tools
ensure_installed git
ensure_installed jq
ensure_installed nano


debug_echo "DEBUG" "checking if n is installed"
if [ -f "/usr/local/bin/n" ]; then
debug_echo "DEBUG" "installed /usr/local/bin/n"
elif [ -f "/usr/bin/n" ]; then
debug_echo "DEBUG" "installed /usr/bin/n"
else
    cecho "YELLOW" "n is not installed. Installing n..."
    # Install n (Node.js version manager)
    sudo npm install -g n
fi

sudo n install 18

debug_echo "DEBUG" "checking if node-red is installed"
if [ -f "/usr/local/bin/node-red" ]; then
debug_echo "DEBUG" "installed /usr/local/bin/node-red"
elif [ -f "/usr/bin/node-red" ]; then
debug_echo "DEBUG" "installed /usr/bin/node-red"
else
    bash <(curl -sL https://raw.githubusercontent.com/node-red/linux-installers/master/deb/update-nodejs-and-nodered)  --confirm-install  --confirm-pi --node18  
fi


# Check and generate SSH key
if [ ! -f "$SSH_KEY_PATH" ]; then
    cecho "GREEN" "Generating new SSH key..."
    ssh-keygen -t rsa -b 4096 -f $SSH_KEY_PATH -N ""
fi

check_git_access() {
        local server_address=$1
    local username=$2

    # Attempt to SSH into the server with a timeout of 10 seconds
    if ssh -o ConnectTimeout=10 -q $GIT_SERVER ; then
      cecho "BLUE" "SSH access to Git server verified."
    else
        cecho "RED" "Git server access failed. Updating SSH key in Airtable..."
        update_ssh_key_in_airtable
    fi
}
create_airtable_record() {
    local hostname=$1
    local ssh_key=$2

    local data=$(jq -n \
                    --arg hostname "$hostname" \
                    --arg sshKey "$ssh_key" \
                    '{records: [{fields: {"Device id": $hostname, "SSH Public Key": $sshKey}}]}')

    local response=$(curl -X POST \
        "https://api.airtable.com/v0/${AIRTABLE_BASE_ID}/${AIRTABLE_TABLE_NAME}" \
        -H "Authorization: Bearer ${AIRTABLE_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "$data")

    #echo "Response from Airtable:"
    #echo "$response"
}
update_airtable_record() {
    debug_echo "DEBUG" "update_airtable_record called"

    local record_id=$1
    local data=$(jq -n \
                    --arg sshKey "$$2" \
                    '{fields: {"SSH Public Key": $sshKey}}')

    local response=$(curl -X PATCH \
        "https://api.airtable.com/v0/${AIRTABLE_BASE_ID}/${AIRTABLE_TABLE_NAME}/${record_id}" \
        -H "Authorization: Bearer ${AIRTABLE_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "$data")

    #echo "Response from Airtable:"
    #echo "$response"
}


fetch_airtable_record_id_by_hostname() {
     debug_echo "DEBUG" "fetch_airtable_record_id_by_hostname called"
    local hostname=$1
    # Encoding the formula: {Device id} = 'hostname'
    local encodedFormula=$(printf "{Device id} = '%s'" "$hostname" | jq -sRr @uri)

    local response=$(curl -X GET \
        "https://api.airtable.com/v0/${AIRTABLE_BASE_ID}/${AIRTABLE_TABLE_NAME}?filterByFormula=${encodedFormula}" \
        -H "Authorization: Bearer ${AIRTABLE_API_KEY}" \
        -H "Content-Type: application/json")

    echo $(echo $response | jq -r '.records[0].id // empty')
}

# Check Git access and update Airtable if necessary
update_ssh_key_in_airtable() {
    debug_echo "DEBUG" "update_ssh_key_in_airtable called"
    local pubkey=$(cat $SSH_KEY_PATH.pub)
    local record_id=$(fetch_airtable_record_id_by_hostname "$HOSTNAME")
    debug_echo "DEBUG" "$pubkey "
    debug_echo "DEBUG" "$record_id "
    echo "Fetched record id ?${record_id}?"
    if [ -n "$record_id" ]; then
        # Update existing record
        cecho "GREEN" "Updating Existing Record"
        update_airtable_record "$record_id" "$pubkey"
    else
        cecho "GREEN" "Create new Record"
        # Create new record
        create_airtable_record "$HOSTNAME" "$pubkey"
    fi
}


 debug_echo "DEBUG" "checking git access"
 check_git_access
 debug_echo "DEBUG" "Git access successfull"

# Constants
GIT_SERVER="ssh://git@fleet-flow-git.lizzardsolutions.com/home/git/git"
BRANCH="main"
SSH_KEY_PATH="$HOME/.ssh/id_rsa"
cd $HOME
# Function to clone Git repositories
 debug_echo "DEBUG" "cloning repository"
clone_repository() {
    local repo_name=$1
    local branch=$2
    until git clone --single-branch --branch ${branch} ${GIT_SERVER}/${repo_name}.git; do
        cecho "YELLOW" "Git clone of $repo_name failed. Retrying..."
        sleep 5
    done
}

# Clone repositories
clone_repository "fleet-files" ${BRANCH}
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
RESTART_COMMAND='find  $HOME/fleet-files -maxdepth 1 -type f -exec cp {}  $HOME//.node-red/ \ && sudo systemctl restart nodered'
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
ExecStart=/usr/bin/npm run start--force
WorkingDirectory=$(realpath ./fleet-flows-js)
Restart=always
User=$(whoami)

KillSignal=SIGINT
[Install]
WantedBy=multi-user.target
EOL

    sudo systemctl enable fleet-flows-js.service
    sudo systemctl start fleet-flows-js.service
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
if [ -d "\$PROJECT_DIR" ] && [ -d "\$PROJECT_DIR/.git" ]; then
    cd \$PROJECT_DIR
    echo "Current Directory: $(pwd)"

    # Fetch the latest commits from the remote
    git fetch

    # Compare local HEAD and remote HEAD
    LOCAL_SHA=\$(git rev-parse HEAD)
    REMOTE_SHA=\$(git rev-parse origin/main)

    if [ "\$LOCAL_SHA" != "\$REMOTE_SHA" ]; then
        echo "Updating the project..."

        # Backup specified files
        echo "Backing up files..."
        mkdir -p \$BACKUP_DIR
        for file in "\${FILES_TO_BACKUP[@]}"; do
            if [ -f "\$file" ]; then
                cp \$file \$BACKUP_DIR/
            else
                echo "Warning: File \$file not found for backup."
            fi
        done

        # Delete the project directory
        echo "Deleting old project directory..."
        cd ..
        rm -rf \$PROJECT_DIR

        # Clone the remote repository
        echo "Cloning the remote repository..."
        git clone --single-branch --branch \$BRANCH \$REMOTE_REPO \$PROJECT_DIR

        # Restore the backed-up files
        echo "Restoring files..."
        cd \$PROJECT_DIR
        npm i 
        for file in "\${FILES_TO_BACKUP[@]}"; do
            if [ -f "\$BACKUP_DIR/\$file" ]; then
                cp \$BACKUP_DIR/\$file .
            else
                echo "Warning: Backup of \$file not found for restoration."
            fi
        done

        echo "Update complete."
    else
        echo "Project is already up to date."
    fi
else
    echo "Error: The directory \$PROJECT_DIR is not a valid Git repository."
fi
EOL
    sudo chmod +rx $AUTO_UPDATER_SCRIPT
    LOG_FILE="/var/log/auto_updater_ffjs.log"
    # Setup the cronjob for auto-update
(crontab -l 2>/dev/null; echo "0 * * * * $AUTO_UPDATER_SCRIPT >> $LOG_FILE 2>&1") | crontab -
    cecho "GREEN" "auto-update is setup $AUTO_UPDATER_SCRIPT"

}
create_auto_update_job

# Restart on changes
restart_on_changes() {
    RESTART_SCRIPT="/usr/local/bin/restart_change_ffjs.sh"
    sudo tee $RESTART_SCRIPT > /dev/null <<EOL
#!/bin/bash

echo "Listening..."
# Define the project directory
PROJECT_DIR=\$HOME/fleet-flows-js
LOG_FILE="$HOME/restart_change_ffjs.log"

# List of files to monitor within the project directory
FILES=("schema.yml" ".env")

# Function to monitor changes and restart the service
monitor_and_restart() {
    echo "Monitoring files for changes..." >> "\$LOG_FILE"
    inotifywait -m -e modify -q \${FILES[@]/#/\$PROJECT_DIR/} | while read -r line; do
        if systemctl is-active --quiet fleet-flows-js.service; then
            echo "File change detected: \$line" >> "\$LOG_FILE"
            echo "Restarting the fleet-flows-js service..." >> "\$LOG_FILE"
            sudo systemctl restart fleet-flows-js.service >> "\$LOG_FILE" 2>&1
        else
            echo "File change detected, but fleet-flows-js.service is not running." >> "\$LOG_FILE"
        fi
    done
}

# Run the monitor function
monitor_and_restart

EOL

    sudo chmod +rx $RESTART_SCRIPT
    cecho "GREEN" "Restart Script is setup $RESTART_SCRIPT"
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
WorkingDirectory=/usr/local/bin/
ExecStart=/usr/local/bin/restart_change_ffjs.sh
Restart=on-failure
RestartSec=5
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
EOL

    sudo systemctl daemon-reload
    sudo systemctl enable fleet-flows-js-listener
    sudo systemctl start fleet-flows-js-listener

    cecho "GREEN" "Fleet Flows JS Listener Service is setup and started. "
}
create_fleet_flow_js_listener_service
