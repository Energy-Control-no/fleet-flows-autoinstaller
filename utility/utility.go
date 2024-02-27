package utility

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
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
	Red          = "\u001b[31m"
	BrightRed    = "\u001b[31;1m"
	Green        = "\u001b[32m"
	BrightGreen  = "\u001b[32;1m"
	Yellow       = "\u001b[33m"
	BrightYellow = "\u001b[33;1m"
	Reset        = "\u001b[0m"
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

// ensure Installed function to ensure packages are present, if not it installs them.
func EnsureInstalled(commandName string) {
	// defer Wg.Done() // Decrement the wait group when the goroutine completes
	// Check if the command is available
	_, err := exec.LookPath(commandName)
	if err != nil {
		fmt.Printf(Yellow+"Installing %s...\n %s", commandName, Reset)
		// Install the command using apt
		installCmd := exec.Command("sudo", "apt", "-y", "install", commandName)
		_, err := installCmd.CombinedOutput()
		if err != nil {
			Logger(err)
			log.Fatal(Red, "error installing : ", commandName, err, Reset)
		}
		fmt.Println(Green+commandName, "installed successfully.", Reset)

	} else {
		fmt.Println(Green+commandName, " is already installed.", Reset)
	}
}

// helper function to check if node is installed,if not it installs it.
func EnsureNodeInstalled() {
	// Check if npm 10.4.0 is installed
	cmdNPM := exec.Command("npm", "-v")
	_, err := cmdNPM.Output()
	if err != nil {
		// npm is not installed or version is not 10.4.0, uninstall and install npm 10.4.0
		fmt.Println(BrightYellow, "npm is not installed. Installing npm....", Reset)
		installNPM := exec.Command("sudo", "apt-get", "install", "-y", "npm="+os.Getenv("NPM_VERSION"))
		output, err := installNPM.Output()
		if err != nil {
			Logger(err)
			log.Fatal(Red, "error installing npm 10.4.0: ", string(output), Reset)
		}
		fmt.Println(Green, "npm installed successfully, version: ", os.Getenv("NPM_VERSION"), Reset)

	} else {
		// npm is installed and version is 10.4.0
		fmt.Println(Green, "npm is installed", Reset)
	}
	// // Check if n is installed
	// cmd := exec.Command("which", "n")
	// output, err := cmd.Output()
	// if err != nil {
	// 	// n is not installed, install it
	// 	fmt.Println(BrightYellow, "n is not installed. Installing n...", Reset)
	// 	installN := exec.Command("sudo", "npm", "install", "-g", "n")
	// 	_, err := installN.Output()
	// 	if err != nil {
	// 		Logger(err)
	// 		log.Fatal(Red, "error installing n: ", err, Reset)
	// 	}

	// } else {
	// 	// n is installed, output its path
	// 	nPath := strings.TrimSpace(string(output))
	// 	fmt.Println(Green, "n is installed at:", nPath, Reset)
	// }

	// Use n to install Node.js version 18
	installNode18 := exec.Command("sudo", "apt-get", "install", "-y", "nodejs="+os.Getenv("NODE_VERSION"))
	output, err := installNode18.Output()
	if err != nil {
		log.Println("output : ", string(output))
		log.Fatal(Red, "error installing Node.js version 18: ", err, Reset)
	}
	fmt.Println(Green, "Node is installed at version: ", os.Getenv("NODE_VERSION"), Reset)
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

		// Fetch the script using curl
		scriptURL := "https://raw.githubusercontent.com/node-red/linux-installers/master/deb/update-nodejs-and-nodered"
		scriptContent, err := FetchScript(scriptURL)
		if err != nil {
			Logger(err)
			log.Fatal(Red, "error fetching Node-RED installation script: ", err, Reset)
		}

		// Save the script content to a file
		scriptPath := "/node-red-script.sh"
		err = ioutil.WriteFile(scriptPath, []byte(scriptContent), 0755)
		if err != nil {
			Logger(err)
			log.Fatal(Red, "error saving installation script: ", err, Reset)
		}

		// Execute the script file
		installCmd := exec.Command("bash", scriptPath, "--confirm-install", "--node18", "--confirm-root")

		// Capture both stdout and stderr
		combinedOutput, err := installCmd.CombinedOutput()

		// Log or print the combined output for debugging purposes
		log.Println("combinet output: ", string(combinedOutput))
		// Check for error and log the actual error message
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				// The command exited with a non-zero status
				stderr := exitError.Stderr
				log.Println("Error: ", string(stderr))
			}

			log.Fatal(Red, "Error installing Node-RED: ", err, Reset)
		}

	} else {
		fmt.Println(BrightGreen, "Node-RED is installed at: ", nRedCmd, Reset)
	}
}

/*Go:Open file for Log Critical Error Message */
func OpenLogFile() *os.File {
	currentDir, _ := os.Getwd()
	loggedErrorPath := currentDir + "/"
	if loggedErrorPath == "" {
		loggedErrorPath = "auto-installer-with-go/"
		fmt.Println("Error Log File: env variable not found")
	}
	logFileURI := loggedErrorPath + "Errorlogged.txt"
	LogFile, err := os.OpenFile(logFileURI, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println(Red+"Error Log File: ", err, Reset)
	}
	ErrorLog = log.New(LogFile, "[DANGER] ", log.Llongfile|log.LstdFlags) // set new logger ErrorLog because of Logger is Reserved word

	// fmt.Println("Log File Open: ", logFileURI)
	return LogFile
}

/* Go: Log Critical Error Message on file if CI_ENVIRONMENT is production in env file then send email to EMAIL_FOR_CRITICAL_ERROR */
func Logger(errObject error) {
	if errObject != nil { // null checking because of stuck server when error is null
		//using 1 indicate actually error
		_, _, _, ok := runtime.Caller(1)
		if !ok {
			err := errors.New("failed to get filename")
			fmt.Println(Red+"Error Log File: ", err, Reset)
		}
		ErrorLog.Output(2, errObject.Error())
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

func IsAccessLogOlderThanMinutes(creationTimeString string, minutes int) bool {
	// Parse the creation time string into a time.Time object with the "Asia/Kolkata" time zone.
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Println("Error loading time zone:", err)
		return false
	}

	creationTime, err := time.ParseInLocation("2006-01-02 15:04:05", creationTimeString, location)
	if err != nil {
		log.Println("Error parsing creation time:", err)
		return false
	}

	// Calculate the time difference in minutes.
	timeDifference := time.Since(creationTime).Minutes()

	// Check if the time difference is non-negative and greater than the specified threshold.
	return timeDifference >= 0 && timeDifference > float64(minutes)
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
			case "Tried access":
				accessAttempts = fields[1]
			}
		}
	}

	// Check for errors during scanning.
	if err := scanner.Err(); err != nil {
		return err, false
	}

	if hostname == actualHostname && !IsAccessLogOlderThanMinutes(creationTime, 10) && accessAttempts == "yes" {
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
		Logger(err)
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
