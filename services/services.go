package services

import (
	"fmt"
	config "installer/configs"
	"installer/utility"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// All functions to create services

// A function to create all the services one by one
func CreateServices() {
	// enable create & start services
	utility.ErrorLog.Output(2, "calling createSystemdService()......")
	fmt.Println(utility.Yellow, "calling createSystemdService()......", utility.Reset)
	createSystemdService()
	utility.ErrorLog.Output(2, "calling restartOnChanges()......")
	fmt.Println(utility.Yellow, "calling restartOnChanges()......", utility.Reset)
	restartOnChanges()
	utility.ErrorLog.Output(2, "calling createFleetFlowJSListenerService()......")
	fmt.Println(utility.Yellow, "calling createFleetFlowJSListenerService()......", utility.Reset)
	createFleetFlowJSListenerService()
	utility.ErrorLog.Output(2, "calling createAutoUpdateJob()......")
	fmt.Println(utility.Yellow, "calling createAutoUpdateJob()......", utility.Reset)
	createAutoUpdateJob()
}

// creates a systemd service names fleet-flows-js.service for running node-red-helper
func createSystemdService() {
	//homeDir, err := os.UserHomeDir()
	homeDir := os.Getenv("HOME_DIR")
	// user, _ := user.Current()  // PREV
	user := os.Getenv("USER_NAME")
	/*
		if err != nil {
			utility.Logger(err, utility.Error)
			log.Fatal(utility.Red, "error getting user's home directory: ", err, utility.Reset)
		}
	*/
	npmLocation, err := findNpmLocation()
	if err != nil {
		log.Fatal("error finding npm location, please provide it yourself: ", err)
	}
	serviceFilePath := "/etc/systemd/system/fleet-flows-js.service"
	// serviceFilePath := "fleet-flows-js.service"
	file, err := os.OpenFile(serviceFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error creating systemd service file: ", err, utility.Reset)
	}
	err = utility.SetPermissions(serviceFilePath)
	if err != nil {
		fmt.Println(utility.Yellow, "error setting permission for systemd file: ", err, utility.Reset)
	}

	defer file.Close()

	serviceContent := fmt.Sprintf(`[Unit]
Description=OTA flow updates and flow compiling for node-red
After=network.target

[Service]
ExecStart=%s run start--force
WorkingDirectory=%s/fleet-flows-js
Restart=always
User=%s
KillSignal=SIGINT

[Install]
WantedBy=multi-user.target
`, npmLocation, homeDir, user)

	_, err = file.WriteString(serviceContent)
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error writing to systemd service file: ", err, utility.Reset)
	}
	// Enable and start the service
	cmd := exec.Command("sudo", "systemctl", "enable", "fleet-flows-js.service")
	err = cmd.Run()
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error enabling systemd service: ", err, utility.Reset)
	}

	cmd = exec.Command("sudo", "systemctl", "start", "fleet-flows-js.service")
	err = cmd.Run()
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error starting systemd service: ", err, utility.Reset)
	}

	fmt.Println("Systemd service created, enabled, and started successfully.")

}
func findNpmLocation() (string, error) {
	cmd := exec.Command("whereis", "npm")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Extracting the location from the output
	location := strings.Fields(string(output))[1]
	return location, nil
}

// creates a restart script that monitors any changes in fleet-files and restarts the node-red-helper
func restartOnChanges() {
	//homeDir, _ := os.UserHomeDir()
	homeDir := os.Getenv("HOME_DIR")
	// Define restart script path
	projectDir := homeDir + "/fleet-flows-js"
	logFile := homeDir + "/restart_change_ffjs.log"

	// Write script content to file
	scriptContent := fmt.Sprintf(`#!/bin/bash

echo "Listening..."
# Define the project directory
LOG_FILE="%s"

# List of files to monitor within the project directory
FILES=("%s/schema.yml" "%s/.env")

# Function to monitor changes and restart the service
monitor_and_restart() {
    echo "Monitoring files for changes..." >> "$LOG_FILE"
    inotifywait -m -e modify -q "${FILES[@]}" | while read -r line; do
        if systemctl is-active --quiet fleet-flows-js.service; then
            echo "File change detected: \$line" >> "$LOG_FILE"
            echo "Restarting the fleet-flows-js service..." >> "$LOG_FILE"
            sudo systemctl restart fleet-flows-js.service >> "$LOG_FILE" 2>&1
        else
            echo "File change detected, but fleet-flows-js.service is not running." >> "$LOG_FILE"
        fi
    done
}

# Run the monitor function
monitor_and_restart
`, logFile, projectDir, projectDir)

	err := ioutil.WriteFile(config.RestartScript, []byte(scriptContent), 0755)
	if err != nil {
		log.Fatal(utility.Red, "error writing restart script: ", err, utility.Reset)
	}
	// give permission to the user
	err = utility.SetPermissions(config.RestartScript)
	if err != nil {
		fmt.Println(utility.Yellow, "error setting permission for restart script: ", err, utility.Reset)
	}
	// Change permissions of the restart script
	cmd := exec.Command("sudo", "chmod", "+rx", config.RestartScript)
	err = cmd.Run()
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error changing permissions of restart script: ", err, utility.Reset)
	}

	fmt.Printf("Restart script is set up: %s\n", config.RestartScript)
}

// this function triggers the restart script
func createFleetFlowJSListenerService() {
	// get host name
	// user, _ := user.Current() // prev
	user := os.Getenv("USER_NAME")
	// Define service file path
	serviceFilePath := "/etc/systemd/system/fleet-flows-js-listener.service"

	// Create the service file
	file, err := os.Create(serviceFilePath)
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error creating systemd service file: ", err, utility.Reset)
	}
	defer file.Close()

	// Write the content to the service file
	serviceContent := fmt.Sprintf(`[Unit]
Description=Fleet Flow JS File Change Listener
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=/usr/local/bin/
ExecStart=/usr/local/bin/restart_change_ffjs.sh
Restart=on-failure
RestartSec=5
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
`, user)

	_, err = file.WriteString(serviceContent)
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error writing to systemd service file: ", err, utility.Reset)
	}
	// give permission to the user
	err = utility.SetPermissions(serviceFilePath)
	if err != nil {
		fmt.Println(utility.Yellow, "error setting permission for FleetFlowJSListenerService: ", err, utility.Reset)
	}
	// Reload systemd daemon to read the new service file
	cmd := exec.Command("sudo", "systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error reloading systemd daemon: ", err, utility.Reset)
	}

	// Enable and start the service
	cmd = exec.Command("sudo", "systemctl", "enable", "fleet-flows-js-listener")
	err = cmd.Run()
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error enabling systemd service: ", err, utility.Reset)
	}

	cmd = exec.Command("sudo", "systemctl", "start", "fleet-flows-js-listener")
	err = cmd.Run()
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error starting systemd service: ", err, utility.Reset)
	}

	fmt.Println("Fleet Flows JS Listener Service is setup and started.")
}

func createAutoUpdateJob() {
	// Define auto-updater script path
	autoUpdaterScript := config.AutoUpdaterScript
	// logFile := LOG_FILE
	//homeDir, _ := os.UserHomeDir()
	homeDir := os.Getenv("HOME_DIR")
    user := os.Getenv("USER_NAME")
	// Write script content to file
	scriptContent := fmt.Sprintf(`#!/bin/bash

# Set the project directory and backup directory
PROJECT_DIR=%s/fleet-flows-js
BACKUP_DIR=%s/fleet-flows-js.backup
REMOTE_REPO=%s/fleet-flows-js.git
FILES_TO_BACKUP=(".env","schema.yml") # Add other files as needed
BRANCH=%s

if [ -d "$PROJECT_DIR" ] && [ -d "$PROJECT_DIR/.git" ]; then
    cd "$PROJECT_DIR"
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

		# trying a pull
		# Attempt to pull changes
    	git pull origin $BRANCH

		# Check if the pull was successful
    	if [ $? -ne 0 ]; then
        	echo "Git pull failed. Trying to reset the branch and pull again..."
        
			# Reset the branch to the remote state
			git reset --hard origin/$BRANCH
			
			# Try pulling again
			git pull origin $BRANCH
		fi
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
`, homeDir, homeDir, *config.Repository, *config.FilesBranch)

	err := ioutil.WriteFile(autoUpdaterScript, []byte(scriptContent), 0755)
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error writing auto-updater script: ", err, utility.Reset)
	}
	// give permission to the user
	err = utility.SetPermissions(autoUpdaterScript)
	if err != nil {
		fmt.Println(utility.Yellow, "error setting permission for autoUpdaterScript: ", err, utility.Reset)
	}
	// Setup the cronjob for auto-update
	cronjob := fmt.Sprintf("0 * * * * su %s -c \"%s\"", user, autoUpdaterScript)


	// Write the updated cronjobs to the crontab
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo \"%s\" | crontab -", cronjob))
	err = cmd.Run()
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error setting up cronjob: ", err, utility.Reset)
	}

	fmt.Println(utility.Green, "Cron job set up successfully: ", cronjob, utility.Reset)
}
