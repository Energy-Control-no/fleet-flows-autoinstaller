#!/bin/bash
# Check if Airtable API key is provided as an argument
if [ -z "$1" ]; then
    echo "Error: No Airtable API key provided."
    echo "Usage: $0 <AIRTABLE_API_KEY>"
    exit 1
fi

# Constants
AIRTABLE_API_KEY="$1"
HOSTNAME=$(hostname)
SSH_KEY_PATH="$HOME/.ssh/id_rsa"
# Update package lists
sudo apt update
sudo apt upgrade

# Check and install required packages
ensure_installed() {
    if ! command -v $1 &> /dev/null; then
        echo "Installing $1..."
        sudo apt install -y $1
    else
        echo "$1 is already installed."
    fi
}

# Update npm to the latest version
sudo npm install -g npm@latest
# Check if 'n' is installed
if ! command -v n >/dev/null 2>&1; then
    echo "n is not installed. Installing n..."
    # Install n (Node.js version manager)
    sudo npm install -g n
fi

# Update Node.js to the latest version using 'n'
sudo n lts
sudo n latest

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

update_airtable_record() {
    local record_id=$1
    local ssh_key=$2
    local data=$(jq -n \
                    --arg sshKey "$ssh_key" \
                    '{fields: {"SSH Public Key": $sshKey}}')

    local response=$(curl -X PATCH \
        "https://api.airtable.com/v0/${AIRTABLE_BASE_ID}/${AIRTABLE_TABLE_NAME}/${record_id}" \
        -H "Authorization: Bearer ${AIRTABLE_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "$data")

    echo "Response from Airtable:"
    echo "$response"
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

    echo "Response from Airtable:"
    echo "$response"
}

# Check and generate SSH key
if [ ! -f "$SSH_KEY_PATH" ]; then
    echo "Generating new SSH key..."
    ssh-keygen -t rsa -b 4096 -f $SSH_KEY_PATH -N ""
fi
check_git_access() {
    ssh -o BatchMode=yes -T $GIT_SERVER 2>&1 | grep -q "successfully authenticated"
}
# Check Git access and update Airtable if necessary
if ! check_git_access; then
    echo "Git server access failed. Updating SSH key in Airtable..."
    update_ssh_key_in_airtable
fi
