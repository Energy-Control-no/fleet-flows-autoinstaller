package main

import (
	"fmt"
	config "installer/configs"
	gitops "installer/gitOps"
	"installer/utility"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		if strings.Contains(err.Error(), ".env: no such file or directory") {
			fmt.Println(utility.Yellow, ".env file not found, checking to run with system variables...", utility.Reset)
		} else {
			fmt.Println(utility.Red, ".env file wasn't found, looking at env variables: ", err.Error(), utility.Reset)
			return
		}
	}

	// parsing flags to access them later
	config.Init()

	// env vars check
	if !utility.EnvVariablesCheck() {
		os.Exit(1)
	}

	// check if the user is root user or not
	if !utility.CheckForElevatedPriveleges() {
		return
	}
	// Check if any flags are passed and are they valid
	if !utility.CheckFlags(os.Args) {
		return
	}

	// opening log file
	utility.LogFile = utility.OpenLogFile()

	// ensuring neccesary tools installed
	utility.ErrorLog.Output(2, "calling ensureInstalled() for [inotify-tools,git,jq,nano].....")
	fmt.Println(utility.Yellow, "calling ensureInstalled() for [inotify-tools,git,jq,nano].....", utility.Reset)

	utility.EnsureInstalled("inotify-tools")
	utility.EnsureInstalled("git")
	utility.EnsureInstalled("jq")
	utility.EnsureInstalled("nano")
	utility.EnsureInstalled("curl")

	// adding a waitgroup here
	// Increment the wait group for each goroutine
	utility.Wg.Add(1)
	go gitops.CheckGitAccessAndCloneIfAccess()

	// Wait for all goroutines to finish
	utility.Wg.Wait()

}
