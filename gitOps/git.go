package gitops

import (
	"fmt"

	"installer/airtable"
	config "installer/configs"
	"installer/services"
	"installer/utility"

	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// all git operation functions are defined here

// Checks Access to git server and clones the repository if access is verfied
func CheckGitAccessAndCloneIfAccess() {

	defer utility.Wg.Done() // Decrement the wait group when the goroutine completes
	loopCount := 1
	maxTries, err := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	if err != nil {
		fmt.Println(utility.Yellow, "Please update value of MAX_RETRIES in environment...currently setting it to 5")
		maxTries = 5
	}
	sleepBetween, err := strconv.Atoi(os.Getenv("SLEEP_BETWEEN"))
	if err != nil {
		fmt.Println(utility.Yellow, "Please update value of SLEEP_BETWEEN in environment...currently setting it to 5 minutes")
		sleepBetween = 5
	}

	for {
		if maxTries == 0 {
			// we need to close program as we already tried for access for the max number of times
			log.Fatal(utility.BrightRed, "Max tries reached, Couldn't get access to git server. Try again later! exiting....", utility.Reset)
		}
		err, isOk := utility.CheckAccessLog()
		if !isOk && err == nil {
			// err == nil because if its the first time its going to throw an
			// error for : no such file or directory
			// we'll bypass it for the first time so that access log can get created
			// call to check if accesslog is older than 10 mins
			isOk, difference := utility.IsAccessLogOlderThanMinutes(utility.AccessLogCreatedOn, 10, "git")
			if isOk {
				// try again then
				continue
			}
			fmt.Println(utility.Yellow, "It hasn't been 10 mins since you last tried to access the server and it wasn't verified, please try again after some time or wait "+fmt.Sprintf("%d", difference)+"mins for an auto re-attempt.", utility.Reset)
			// sleep before retrying again for the time remaining
			time.Sleep(time.Minute * time.Duration(difference))

		}
		// routine function to check for server access and clone repoS
		utility.ErrorLog.Output(2, "checking git server access...")
		fmt.Println(utility.Yellow + "checking git server access..." + utility.Reset)
		if CheckGitAccess(*config.Repository) {
			// we have server access, break the loop and continue cloning repositories
			break
		} else {
			if *config.Key == "" {
				fmt.Println("Usage: -k <AIRTABLE_API_KEY>")
				log.Fatal(utility.Red + "No AIRTABLE_API_KEY provided." + utility.Reset)
			} else {
				// checking the key validity
				if !airtable.CheckAirtableAPIKey(*config.Key, *config.Base, *config.Table) {
					log.Fatal(utility.Red + "Invalid API key. Stopping the script." + utility.Reset)
				}
				// check if the public SSH key exists or not
				homeDir, err := os.UserHomeDir()
				if err != nil {
					utility.Logger(err, utility.Error)
					log.Fatal(utility.Red, "error getting user home dir: ", err, utility.Reset)
				}
				_, err1 := os.Stat(homeDir + "/" + config.SSHKeyPath)
				if err1 == nil {
					utility.ErrorLog.Output(2, "SSH key file already exists.")
					fmt.Println(utility.Yellow + "SSH key file already exists." + utility.Reset)
					fmt.Println(utility.Yellow + "Replicating the same keys in " + os.Getenv("HOME_DIR") + utility.Reset)
					err := utility.CopyDir(homeDir+"/"+".ssh", os.Getenv("HOME_DIR")+"/.ssh")
					if err != nil {
						log.Fatal(utility.Red+"Error while replicating ssh keys in", os.Getenv("HOME_DIR"), "error: ", err, utility.Reset)
					}
					// if already exists update it in Air Table
					utility.ErrorLog.Output(2, "calling updateSSHKeyInAirtable()...")
					fmt.Println(utility.Yellow + "calling updateSSHKeyInAirtable()..." + utility.Reset)
					airtable.UpdateSSHKeyInAirtable()
				} else if !os.IsNotExist(err1) {
					utility.ErrorLog.Output(2, "Error while searching for PUB SSH KEY: "+err1.Error())
					log.Fatal(utility.Red+"Error while searching for PUB SSH KEY: ", err1, utility.Reset)
				} else {
					// generate a new PUB SSH KEY
					utility.ErrorLog.Output(2, "calling generateSSHKey()...")
					fmt.Println(utility.Yellow+"calling generateSSHKey()... with SSH_KEY_PATH: ", config.SSHKeyPath, utility.Reset)
					err := utility.GenerateSSHKey(config.SSHKeyPath)
					if err != nil {
						utility.Logger(err, utility.Error)
						log.Fatal(utility.Red+"Error while generating PUB SSH KEY: ", err, utility.Reset)
					}
					err = utility.CopyDir(homeDir+"/"+".ssh", os.Getenv("HOME_DIR")+"/.ssh")
					if err != nil {
						log.Fatal(utility.Red+"Error while replicating ssh keys in", os.Getenv("HOME_DIR"), "error: ", err, utility.Reset)
					}
					utility.ErrorLog.Output(2, "calling updateSSHKeyInAirtable()...")
					fmt.Println(utility.Yellow + "calling updateSSHKeyInAirtable()..." + utility.Reset)
					// on successful creation update in airtable
					airtable.UpdateSSHKeyInAirtable()
				}
				utility.CreateAccessLog()
				// creating accesslog for current host
				utility.ErrorLog.Output(2, "Server access attempeted!! log created, Don't close the program we are retrying for access in "+os.Getenv("SLEEP_BETWEEN")+" mins "+"("+fmt.Sprintf("%d", loopCount)+"/"+os.Getenv("MAX_RETRIES")+")")
				fmt.Println(utility.BrightYellow + "Server access attempeted!! log created, Don't close the program we are retrying for access in " + os.Getenv("SLEEP_BETWEEN") + " mins " + "(" + fmt.Sprintf("%d", loopCount) + "/" + os.Getenv("MAX_RETRIES") + ")" + utility.Reset)
			}
		}
		// decrement maxTries after each iteration
		maxTries -= 1
		// increment loop count
		loopCount += 1
		// sleep before retrying again
		time.Sleep(time.Minute * time.Duration(sleepBetween))
	}
	utility.ErrorLog.Output(2, "calling ensureNodeInstalled().....")
	fmt.Println(utility.Yellow + "calling ensureNodeInstalled()....." + utility.Reset)
	// ensure node version manager installed & node -v 18
	utility.EnsureNodeInstalled()
	// after we get access to git server & node is installed
	SwitchDirectoriesAndCloneRepos()
	utility.ErrorLog.Output(2, "calling ensureNodeRedInstalled().....")
	log.Println(utility.Yellow, "calling ensureNodeRedInstalled().....", utility.Reset)
	// ensure node-red installed
	utility.EnsureNodeRedInstalled() // enable node-red installation for now
	// create service files
	services.CreateServices()

	utility.ErrorLog.Output(2, "Installation complete.")
	fmt.Println(utility.BrightGreen, "Installation complete.", utility.Reset)
}

func CheckGitAccess(repository string) bool {
	gitServer := repository

	log.Println(utility.BrightYellow, "Checking Git clone access to ", repository, utility.Reset)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git-clone-check")
	if err != nil {
		fmt.Println(utility.Red, "Error creating temporary directory:", err, utility.Reset)
		return false
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory when done

	// Attempt to clone into the temporary directory
	var repoTocheckAccessFrom = os.Getenv("REPO_TO_CHECK_SERVER_ACCESS")
	if repoTocheckAccessFrom == "" {
		repoTocheckAccessFrom = "fleet-files"
	}
	cmd := exec.Command("git", "clone", "-n", gitServer+"/"+repoTocheckAccessFrom, tempDir)
	output, err := cmd.CombinedOutput()

	if err == nil {
		// If the command succeeded, it means clone access is verified
		fmt.Println(utility.BrightGreen, "Git clone access to", repository, "verified.", utility.Reset)
		return true
	}

	// Check if the error is a timeout error
	if strings.Contains(string(output), "Connection timed out") {
		fmt.Println(utility.Red, "Git server access timed out.", utility.Reset)
		return false
	}

	// Check if the error indicates authentication failure
	if strings.Contains(string(output), "Permission denied") {
		fmt.Println(utility.Red, "Authentication to Git server failed.", utility.Reset)
		return false
	}

	// Check if the error contains a specific message indicating clone access issues
	if strings.Contains(string(output), "could not read Username") {
		fmt.Println(utility.Red, "Git clone access to", repository, "failed.", utility.Red)
		return false
	}

	fmt.Println(utility.Red, "Git clone access to", repository, "failed:", string(output), utility.Reset)
	return false
}

func FindPermissionDenied(output string) bool {
	// Check if the output contains "Permissions denied"
	if _, err := fmt.Sscanf(output, "%s", "Permissions denied"); err == nil {
		return true
	}
	return false
}

// Runs the `git clone` command and clones repos
func CloneRepository(repoName, branch string, gitServer string) error {

	//homeDir, err := os.UserHomeDir()
	//if err != nil {
	//	log.Fatal(utility.Red, "Unable to get user home directory..", utility.Reset)
	//}
	homeDir := os.Getenv("HOME_DIR")
	repoPath := filepath.Join(homeDir, repoName) // Change this to the actual path

	// Check if the repository directory already exists
	if _, err := os.Stat(repoPath); err == nil {
		// Repository already exists, perform git pull
		fmt.Printf("Repository %s already exists. Updating...\n", repoName)

		cmd := exec.Command("git", "pull", "--ff-only", "--recurse-submodules=yes")
		cmd.Dir = repoPath

		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf(utility.Red+"Error updating repository %s: %v\n %s", repoName, err, utility.Reset)
			return fmt.Errorf("Git pull of %s error. Exiting...", repoName)
		}
		err = utility.SetPermissions(repoPath)
		if err != nil {
			return err
		}

		fmt.Printf(utility.BrightGreen+"Repository %s updated successfully.\n %s", repoName, utility.Reset)
		return nil
	}
	for {
		cmd := exec.Command("git", "clone", "--single-branch", "--branch", branch, gitServer+"/"+repoName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf(utility.Red+"Error cloning repository %s, (public key) access denied %v\n %s", repoName, err, utility.Reset)
			permDenied := FindPermissionDenied(string(output))
			if permDenied {
				fmt.Printf("Git clone of %s failed. Retrying...\n", repoName)
				time.Sleep(60 * time.Second) // Wait for a minute before retrying
				continue
			} else {
				return fmt.Errorf("Git clone of %s error. Exiting...", repoName)
			}
		} else {
			fmt.Printf(utility.BrightGreen+"Repository %s cloned successfully.\n %s", repoName, utility.Reset)
			break
		}
	}
	err := utility.SetPermissions(repoPath)
	if err != nil {
		return err
	}
	return nil
}

func CreateEnvFile() {
	// get home dir of user
	//homeDir, err := os.UserHomeDir()
	//if err != nil {
	//	utility.Logger(err, utility.Error)
	//	log.Println(utility.Red, "error getting user home dir: ", err, utility.Reset)
	//}
	homeDir := os.Getenv("HOME_DIR")
	fleetFlowsJsDir := filepath.Join(homeDir, "fleet-flows-js")
	// changing dir to fleet-flows-js dir
	err := os.Chdir(fleetFlowsJsDir)
	if err != nil {
		utility.Logger(err, utility.Error)
		fmt.Println(utility.Red, "Error changing directory to fleet-flows-js directory:", err, utility.Reset)
		os.Exit(1)
	}
	// defining path to .env file
	envFilePath := filepath.Join(fleetFlowsJsDir, ".env")
	file, err := os.Create(envFilePath)
	if err != nil {
		// Handle error
		utility.Logger(err, utility.Error)
		log.Println(utility.Red, "Error creating .env file: ", err, utility.Reset)
		os.Exit(1)
	}
	defer file.Close()
	schemaFilePath := *config.SchemaFilePath

	if schemaFilePath == "" {
		schemaFilePath = "schema.yml"
	}
	if schemaFilePath != "" && !strings.Contains(schemaFilePath, ".yml") {
		fmt.Println(utility.Red, "Usage -sf='path/to/schema.yml'", utility.Reset)
		log.Fatal(utility.Red, "Please provide a valid schema file path", utility.Reset)
	}
	// content for env file
	// Define content for the environment file
	envContent := []byte(fmt.Sprintf(`# Environment variables
	LOCAL_REPO_PATH=%s/fleet-files
	MASTER_BRANCH=%s
	FLOW_FILE_PATH=%s/.node-red/flows.json
	FLOWS_DIR=%s/fleet-files/flows
	SUBFLOWS_DIR=%s/fleet-files/subflows
	RETRY_TIME=5
	SCHEMA_FILE_PATH=%s
	NODE_RED_DIRECTORY=%s/
	CONFIGS_DIR=%s/fleet-files/config
	RESTART_COMMAND='find %s/fleet-files -maxdepth 1 -type f -exec cp {} %s/.node-red/ \; && sudo killall node-red & node-red'
	`, homeDir, *config.FilesBranch, homeDir, homeDir, homeDir, schemaFilePath, homeDir, homeDir, homeDir, homeDir))
	err = ioutil.WriteFile(envFilePath, envContent, 0777)
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error creating environment file: ", err, utility.Reset)
	}
	fmt.Println(utility.BrightGreen, "Environment file created successfully.", utility.Reset)
}
func CreateSchemaFile() {
	// get home dir of user
	/*
		homeDir, err := os.UserHomeDir()
		if err != nil {
			utility.Logger(err, utility.Error)
			log.Println(utility.Red, "error getting user home dir: ", err, utility.Reset)
		}
	*/
	homeDir := os.Getenv("HOME_DIR")
	fleetFlowsJsDir := filepath.Join(homeDir, "fleet-flows-js")
	// changing dir to fleet-flows-js dir
	err := os.Chdir(fleetFlowsJsDir)
	if err != nil {
		utility.Logger(err, utility.Error)
		fmt.Println(utility.Red, "Error changing directory to fleet-flows-js directory:", err, utility.Reset)
		os.Exit(1)
	}
	// defining path to .env file
	schemaFilePath := filepath.Join(fleetFlowsJsDir, "schema.yml")
	file, err := os.Create(schemaFilePath)
	if err != nil {
		// Handle error
		utility.Logger(err, utility.Error)
		log.Println(utility.Red, "Error creating .schema file: ", err, utility.Reset)
		os.Exit(1)
	}
	defer file.Close()

	// content for schema file
	// Define content for the environment file
	schemaContent := []byte(fmt.Sprintf(`# Environment variables
	dependencies:
		- ''
	flows:
		'Welcome':
		  	- basedOn: "flow://welcome"
			- description: "Welcome to Fleet-Flows"
	`))
	err = ioutil.WriteFile(schemaFilePath, schemaContent, 0777)
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "error creating schema file: ", err, utility.Reset)
	}
	fmt.Println(utility.BrightGreen, "Schema file created successfully.", utility.Reset)
}

// switches directories to clone repositories and runs npm install to set them up.
func SwitchDirectoriesAndCloneRepos() {
	utility.ErrorLog.Output(2, "Switching directories.....")
	fmt.Println(utility.Yellow + "Switching directories....." + utility.Reset)
	// after all this is done now we switch directories
	// getting users home dir
	/*
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting users home directory:", err)
			os.Exit(1)
		}
	*/
	homeDir := os.Getenv("HOME_DIR")

	// temporarily switching directories to clone repos
	// Change directory to home directory
	fmt.Println(utility.Yellow + "Changing to home dir....." + utility.Reset)
	err := os.Chdir(homeDir)
	if err != nil {
		fmt.Println("Error changing directory to home directory:", err)
		os.Exit(1)
	}

	// Function to clone Git repositories
	fmt.Println(utility.Yellow + "Cloning repositories..." + utility.Reset)
	err = CloneRepository("fleet-files", *config.FilesBranch, *config.Repository)
	if err != nil {
		utility.Logger(err, utility.Error)
		os.Exit(1)
	}

	err = CloneRepository("fleet-flows-js", *config.FilesBranch, *config.Repository)
	if err != nil {
		utility.Logger(err, utility.Error)
		os.Exit(1)
	}

	// Run npm install in fleet-flows-js
	fleetFlowsJsDir := filepath.Join(homeDir, "fleet-flows-js")
	log.Println(utility.Yellow, "Changing to fleet-flows-js dir.....", utility.Reset)
	err = os.Chdir(fleetFlowsJsDir)
	if err != nil {
		fmt.Println(utility.Red, "Error changing directory to fleet-flows-js directory:", err, utility.Reset)
		os.Exit(1)
	}

	fmt.Println(utility.Yellow, "Running npm install in fleet-flows-js...", utility.Reset)
	cmd := exec.Command("npm", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running npm install in fleet-flows-js:", err)
		os.Exit(1)
	}

	fmt.Println(utility.BrightGreen, "npm install completed successfully.", utility.Reset)

	// create env file at fleet-flows-js dir
	utility.ErrorLog.Output(2, "calling createEnvFile()......")
	fmt.Println(utility.Yellow, "calling createEnvFile()......", utility.Reset)
	CreateEnvFile()

	// create schema file at fleet-flows-js dir
	utility.ErrorLog.Output(2, "calling createSchemaFile()......")
	fmt.Println(utility.Yellow, "calling createSchemaFile()......", utility.Reset)
	CreateSchemaFile()

}
