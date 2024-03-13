package utility

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	config "installer/configs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func SendRequestToAirTable(requestType string, base string, table string, key string) {
	switch requestType {
	case "GET":

	case "PATCH":

	}
}

var (
	Red          = "\u001b[31m [ERROR] "
	BrightRed    = "\u001b[31;1m [ERROR] "
	Green        = "\u001b[32m [SUCCESS] "
	BrightGreen  = "\u001b[32;1m [SUCCESS] "
	Yellow       = "\u001b[33m [PROCESSING] "
	BrightYellow = "\u001b[33;1m [PROCESSING] "
	Reset        = "\u001b[0m"
	Error        = "[ERROR] "
	Debug        = "[DEBUG] "
)

// type Color struct {
// 	Red          string
// 	BrightRed    string
// 	Yellow       string
// 	BrightYellow string
// 	Green        string
// 	BrightGreen  string
// 	Reset        string
// }

var Wg sync.WaitGroup
var LogFile *os.File
var ErrorLog *log.Logger
var AccessLogCreatedOn string

// ensure Installed function to ensure packages are present, if not it installs them.
func EnsureInstalled(commandName string) {
	// defer Wg.Done() // Decrement the wait group when the goroutine completes
	// Check if the command is available
	_, err := exec.LookPath(commandName)
	if err != nil {
		fmt.Printf(Yellow+"Installing %s...\n %s", commandName, Reset)
		// Install the command using apt
		installCmd := exec.Command("apt", "-y", "install", commandName)
		_, err := installCmd.CombinedOutput()
		if err != nil {
			Logger(err, Error)
			log.Fatal(Red, "error installing : ", commandName, err, Reset)
		}
		fmt.Println(Green+commandName, "installed successfully.", Reset)

	} else {
		fmt.Println(Green+commandName, " is already installed.", Reset)
	}
}

func checkAndInstallNodejs() {
	// check if curl exists on host
	cmdCurl := exec.Command("curl", "-V")
	_, err := cmdCurl.Output()
	if err != nil {
		fmt.Println(Yellow, "Installing Curl...", Reset)
		cmdInstallCurl := exec.Command("sudo", "apt", "install", "curl", "-y")
		output, err := cmdInstallCurl.Output()
		if err == nil {
			fmt.Println(Green, "Curl installed successfully.", Reset)
		} else {
			log.Fatal(Red, "Unable to install curl: ", string(output), Reset)
		}
	}
	// install node version repo
	cmdNodeRepo := exec.Command("bash", "-c", "curl -fsSL "+os.Getenv("NODE_SETUP_URL")+" | sudo -E bash -")
	outputNodeRepo, err := cmdNodeRepo.CombinedOutput()
	if err != nil {
		log.Fatal(Red, "Unable to run curl to fetch node setup: ", string(outputNodeRepo), err, Reset)
	}
	fmt.Println(Yellow, "Installing nodejs version ", *config.NodeVersion, Reset)
	// now install node-js
	cmdNodejs := exec.Command("sudo", "apt", "install", "nodejs="+*config.NodeVersion, "-y")
	outputNodejs, err := cmdNodejs.Output()
	if err != nil {
		log.Fatal(Red, "Unable to install nodejs version :", *config.NodeVersion, "output: ", string(outputNodejs), Reset)
	}
	fmt.Println(Green, "Installed nodejs version ", *config.NodeVersion, Reset)
}

// helper function to check if node is installed,if not it installs it.
func EnsureNodeInstalled() {

	cmdNPM := exec.Command("node", "-v")
	outputNPM, err := cmdNPM.Output()
	if err != nil {
		checkAndInstallNodejs()
		return
	}
	output := string(outputNPM)
	log.Println("output : ", output)

	parts := strings.Split(output, "\n")
	latestAvailableVersion := strings.Split(parts[0], "|")
	log.Println("output: ", latestAvailableVersion[0])

	versionNumber := strings.Split(latestAvailableVersion[0], ".")
	// version number that is currently present on host
	log.Println("version no: ", versionNumber[0])

	versionTryingToInstallParts := strings.Split(*config.NodeVersion, ".")
	// // matching it with the requirement
	if !strings.Contains(versionNumber[0], versionTryingToInstallParts[0]) {
		// uninstall nodejs on system
		fmt.Println(Yellow, "Removing present node version and installing node version: ", *config.NodeVersion, Reset)
		cmdNPM := exec.Command("sudo", "apt", "remove", "nodejs")
		outputNPM, err := cmdNPM.Output()
		if err == nil {
			// install our nodejs version
			checkAndInstallNodejs()
		} else {
			log.Fatal(Red, "Unable to remove previous node version and re-install: ", string(outputNPM), Reset)
		}

	} else {
		fmt.Println(Green, "Node is already installed at version: ", *config.NodeVersion, Reset)
	}

}

// Function to fetch script & return as string
func FetchScript(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	scriptContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(scriptContent), nil
}

// helper function to check if Node-red is installed, if not it installs it
func EnsureNodeRedInstalled() {
	nRedCmd, err := exec.LookPath("node-red")
	if err != nil {
		fmt.Println(Yellow, "Node-RED is not installed. Installing...", Reset)

		// Execute the node-red install command
		installCmd := exec.Command("sudo", "npm", "install", "-g", "--unsafe-perm", "node-red")

		// Capture both stdout and stderr
		combinedOutput, err := installCmd.CombinedOutput()

		// Check for error and log the actual error message
		if err != nil {
			// Log or print the combined output for debugging purposes
			log.Println(Debug, "combinet output: ", string(combinedOutput))
			if exitError, ok := err.(*exec.ExitError); ok {
				// The command exited with a non-zero status
				stderr := exitError.Stderr
				log.Println("Error: ", string(stderr))
			}

			log.Fatal(Red, "Error installing Node-RED: ", err, Reset)
		}

		fmt.Println(BrightGreen, "Node-RED installed successfully: ", nRedCmd, Reset)
	} else {
		fmt.Println(BrightGreen, "Node-RED is installed at: ", nRedCmd, Reset)
		// create backup of flows.json as flows-backup.json and save it in home dir
		// function call
		makeFlowsBackupJson()
	}
}

func makeFlowsBackupJson() {
	// Define the directory where Node-RED is installed
	nodeRedDir := "/path/to/node-red-directory"

	// Check if flows.json exists in the Node-RED directory
	flowsFilePath := filepath.Join(nodeRedDir, "flows.json")
	_, err := os.Stat(flowsFilePath)
	if os.IsNotExist(err) {
		fmt.Println(Yellow, "flows.json not found in Node-RED directory! Not making any backup", Reset)
		return
	}

	// Read the content of flows.json
	flowsContent, err := ioutil.ReadFile(flowsFilePath)
	if err != nil {
		log.Fatal(Red, "Failed to read flows.json: \n", err, Reset)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(Red, "Unable to get home directory: ", err, Reset)
	}
	// Define the backup file path (e.g., home directory)
	backupFilePath := filepath.Join(home, "flows-backup.json")

	// Write flows.json content to flows-backup.json
	err = ioutil.WriteFile(backupFilePath, flowsContent, 0644)
	if err != nil {
		log.Fatal(Red, "Failed to write to flows-backup.json: \n", err, Reset)
	}

	fmt.Println(Green, "Backup of flows.json created successfully at ->", flowsFilePath, Reset)
}

/*Go:Open file for Log Critical Error Message */
func OpenLogFile() *os.File {
	currentDir, _ := os.Getwd()
	loggedErrorPath := currentDir + "/"
	logFileURI := loggedErrorPath + "Errorlogged.txt"
	LogFile, err := os.OpenFile(logFileURI, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println(Red+"Error Log File: ", err, Reset)
	}
	ErrorLog = log.New(LogFile, "", log.Llongfile|log.LstdFlags) // set new logger ErrorLog because of Logger is Reserved word

	// fmt.Println("Log File Open: ", logFileURI)
	return LogFile
}

/* Go: Log Critical Error Message on file if CI_ENVIRONMENT is production in env file then send email to EMAIL_FOR_CRITICAL_ERROR */
func Logger(errObject error, errorType string) {
	if errObject != nil { // null checking because of stuck server when error is null
		//using 1 indicate actually error
		_, _, _, ok := runtime.Caller(1)
		if !ok {
			err := errors.New("failed to get filename")
			fmt.Println(Red+"Error Log File: ", err, Reset)
		}
		ErrorLog.Output(2, errorType+" "+errObject.Error())
	}
}

func CreateAccessLog() error {
	// Get the hostname
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return err
	}

	// Get the current time
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Open a file for writing. Create it if it doesn't exist, truncate it otherwise.
	file, err := os.Create("access_log.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close() // Make sure to close the file when done.

	// Create a buffered writer to efficiently write to the file.
	writer := bufio.NewWriter(file)

	// Write data to the file.
	content := fmt.Sprintf("Hostname: %s\nTime of creation: %s\nTried access: yes\n", hostname, currentTime)
	_, err = writer.WriteString(content)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	// Flush the buffer to ensure all data is written to the file.
	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer:", err)
		return err
	}

	fmt.Println(Green + "access_log.txt created successfully." + Reset)
	return nil
}

func IsAccessLogOlderThanMinutes(creationTimeString string, minutes int, from string) (bool, int) {
	// Parse the creation time string into a time.Time object with the "Asia/Kolkata" time zone.
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Println("Error loading time zone:", err)
		log.Fatal(Red, "Check for 'tzdata' package availability, if not available please install it and configure to your local timezone.", Reset, Debug, "apt install tzdata -y")
		return false, 0
	}

	creationTime, err := time.ParseInLocation("2006-01-02 15:04:05", creationTimeString, location)
	if err != nil {
		log.Println("Error parsing creation time:", err)
		return false, 0
	}
	// Calculate the time difference in minutes.
	timeDifference := time.Since(creationTime).Minutes()
	if from == "git" {
		return timeDifference >= 0 && timeDifference > float64(minutes), int(minutes - int(timeDifference))
	}

	// Check if the time difference is non-negative and greater than the specified threshold.
	return timeDifference >= 0 && timeDifference > float64(minutes), 0
}

func CheckAccessLog() (error, bool) {
	actualHostname, _ := os.Hostname()
	filePath := "access_log.txt"
	// Open the file for reading.
	file, err := os.Open(filePath)
	if err != nil {
		return err, false
	}
	defer file.Close() // Make sure to close the file when done.

	// Create a scanner to read the file line by line.
	scanner := bufio.NewScanner(file)

	hostname := ""
	creationTime := ""

	accessAttempts := ""
	// Read each line from the file.
	for scanner.Scan() {
		line := scanner.Text()

		// Split the line into fields.
		fields := strings.Split(line, ": ")

		// Check for the fields you are interested in.
		if len(fields) == 2 {
			switch fields[0] {
			case "Hostname":
				hostname = fields[1]
			case "Time of creation":
				creationTime = fields[1]
				AccessLogCreatedOn = creationTime
			case "Tried access":
				accessAttempts = fields[1]
			}
		}
	}

	// Check for errors during scanning.
	if err := scanner.Err(); err != nil {
		return err, false
	}
	isOk, _ := IsAccessLogOlderThanMinutes(creationTime, 10, "")
	if hostname == actualHostname && !isOk && accessAttempts == "yes" {
		return nil, false
	}

	return nil, true
}

func CheckForElevatedPriveleges() bool {
	u, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	fmt.Println("Current User:", u.Username)

	if u.Uid != "0" {
		log.Fatal(BrightRed + "This program requires root (sudo) privileges." + Reset)
	}

	// Your program logic here...
	fmt.Println(Green + "Running with sudo privileges." + Reset)
	return true
}

// Generates a new SSH Key
func GenerateSSHKey(SSHKeyPath string) error {
	// Generate a new RSA private key
	homeDir, err := os.UserHomeDir()
	if err != nil {
		Logger(err, Error)
		log.Fatal(Red, "error getting user home dir: ", err, Reset)
	}
	SSHKeyPath = filepath.Join(homeDir, SSHKeyPath)
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// Write private key to file
	privateKeyFile, err := os.Create(SSHKeyPath)
	if err != nil {
		log.Fatal("here while encodee", err, "SSH_KEY_PATH: ", SSHKeyPath)
	}
	defer privateKeyFile.Close()
	err = pem.Encode(privateKeyFile, privateKeyPEM)
	if err != nil {
		return err
	}
	// fix permissions
	err = fixKeyFilePermissions(SSHKeyPath)
	if err != nil {
		return err
	}
	// Generate public key from private key
	publicKey := privateKey.PublicKey
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}
	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// Write public key to file
	publicKeyPath := SSHKeyPath + ".pub"
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()
	err = pem.Encode(publicKeyFile, publicKeyPEM)
	if err != nil {
		return err
	}

	fmt.Println("SSH key pair generated successfully.")
	return nil
}

func fixKeyFilePermissions(path string) error {
	// Get the current user's UID and GID
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	// Parse the UID and GID as integers
	uid, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return err
	}

	gid, err := strconv.Atoi(currentUser.Gid)
	if err != nil {
		return err
	}

	// Change the file ownership and permissions
	err = os.Chown(path, uid, gid)
	if err != nil {
		return err
	}

	err = os.Chmod(path, 0600)
	if err != nil {
		return err
	}

	return nil
}
func StringInArray(target string, arr []string) bool {
	// Can change to slices.Contain if we're targetting 1.21+
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}

// this function checks for if the the flags passed are correct or not
func CheckFlags(args []string) bool {
	allowedFlags := []string{"-b", "-fb", "-k", "-r", "-sb", "-t"}
	if len(args) > 1 {

		for i, arg := range args[1:] {
			if arg[0] != '-' {
				log.Println(Red, "Inavlid arguement passed at", i, ": ", arg, Reset)
				printAvailableFlags()
				return false
			}
			parts := strings.Split(arg, "=")
			if !StringInArray(parts[0], allowedFlags) {
				log.Println(Red, "Inavlid flag passed at", i, ": ", arg, Reset)
				printAvailableFlags()
				return false
			}
			// in case of empty flag
			if parts[1] == "" {
				fmt.Print("Flag without value: ", parts[0])
				return false
			}
		}
		return true
	}
	fmt.Println("Initiate the executable with flags")
	printAvailableFlags()
	return false

}

func printAvailableFlags() {
	fmt.Println("Available flags: ")
	fmt.Println("-b string \nAirtable base id, default: " + os.Getenv("AIRTABLE_BASE_ID"))
	fmt.Println("-fb string \nBranch for fleet-files, default: " + os.Getenv("FILES_BRANCH"))
	fmt.Println("-k string \nAirtable API key, default: " + "YOUR_API_KEY_HERE")
	fmt.Println("-r string \nRepository URL, default: " + os.Getenv("GIT_SERVER"))
	fmt.Println("-sb string \nBranch for fleet-flows-js, default: " + os.Getenv("FLOW_JS_BRANCH"))
	fmt.Println("-t string \nAirtable table name, default: " + os.Getenv("AIRTABLE_TABLE"))
	fmt.Println("-sf string \nAbsolute path for schema file, default: " + os.Getenv("SCHEMA_FILE_PATH"))
}

func ExtractNodeJsRepoVersion() {
	// cmdNodeRepoVersion := exec.Command("sudo", "apt-cache", "madison", "nodejs")
	// outputNodeRepos, err := cmdNodeRepoVersion.Output()

	// if err != nil {
	// 	log.Fatal(Red, "Unable to get nodejs versions from cache", Reset)
	// }
	// output := string(outputNodeRepos)
	// cacheVersionStringArray := strings.Split(output, "|")
	// log.Println("cache version: ", cacheVersionStringArray[1])

	// cmdNodeRepo := exec.Command("bash", "-c", "curl", "-fsSL", os.Getenv("NODE_SETUP_URL"), "|", "sudo", "-E", "bash", "-")
	// cmdNodeRepo := exec.Command("bash", "-c", "curl -fsSL "+os.Getenv("NODE_SETUP_URL")+" | sudo -E bash -")
	// outputNodeRepo, err := cmdNodeRepo.CombinedOutput()
	// if err != nil {
	// 	log.Fatal(Red, "Unable to run curl to fetch node setup: ", string(outputNodeRepo), err, Reset)
	// }
	// fmt.Println(Yellow, "Installing nodejs version ", *config.NodeVersion, Reset)
	// // now install node-js
	// cmdNodejs := exec.Command("sudo", "apt", "install", "nodejs="+*config.NodeVersion, "-y")
	// outputNodejs, err := cmdNodejs.Output()
	// if err != nil {
	// 	log.Fatal(Red, "Unable to install nodejs version :", *config.NodeVersion, "output: ", string(outputNodejs), Reset)
	// }
	// fmt.Println(Green, "Installed nodejs version ", *config.NodeVersion, Reset)

}

func EnvVariablesCheck() bool {
	if *config.Repository == "" && os.Getenv("GIT_SERVER") == "" {
		fmt.Println(Red, "Niether Repository flag nor GIT_SERVER env variable set, please set one", Reset)
		return false
	}
	if *config.SoftwareBranch == "" && os.Getenv("FLOW_JS_BRANCH") == "" {
		fmt.Println(Red, "Niether SoftwareBranch flag nor FLOW_JS_BRANCH env variable set, please set one", Reset)
		return false
	}
	if *config.FilesBranch == "" && os.Getenv("FILES_BRANCH") == "" {
		fmt.Println(Red, "Niether FilesBranch flag nor FILES_BRANCH env variable set, please set one", Reset)
		return false
	}
	if *config.Base == "" && os.Getenv("AIRTABLE_BASE_ID") == "" {
		fmt.Println(Red, "Niether Base flag nor AIRTABLE_BASE_ID env variable set, please set one", Reset)
		return false
	}
	if *config.Table == "" && os.Getenv("AIRTABLE_TABLE") == "" {
		fmt.Println(Red, "Niether Table flag nor AIRTABLE_TABLE env variable set, please set one", Reset)
		return false
	}
	if *config.Key == "" && os.Getenv("AIRTABLE_API_KEY") == "" {
		fmt.Println(Red, "Niether Key flag nor AIRTABLE_API_KEY env variable set, please set one", Reset)
		return false
	}
	if *config.SchemaFilePath == "" && os.Getenv("SCHEMA_FILE_PATH") == "" {
		fmt.Println(Red, "Niether SchemaFilePath flag nor SCHEMA_FILE_PATH env variable set, please set one", Reset)
		return false
	}
	if *config.NodeVersion == "" && os.Getenv("NODE_VERSION") == "" {
		fmt.Println(Red, "Niether NodeVersion flag nor NODE_VERSION env variable set, please set one", Reset)
		return false
	}
	if os.Getenv("NODE_SETUP_URL") == "" {
		fmt.Println(Red, "NODE_SETUP_URL env var not set...please set it to the desired version.", Reset)
		return false
	}
	return true
}
