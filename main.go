package main

import (
	"fmt"
	config "installer/configs"
	gitops "installer/gitOps"
	"installer/utility"
	"os"

	"github.com/joho/godotenv"
)

// var (
// 	repository          = flag.String("r", "ssh://git@fleet-flows-git.lizzardsolutions.com", "sets the location of the repository")
// 	softwareBranch      = flag.String("sb", "main", "branch for fleet-flows-js")
// 	filesBranch         = flag.String("fb", "main", "branch for fleet-files")
// 	base                = flag.String("b", "appYWVOaoPhQB0nmA", "airtable base id")
// 	table               = flag.String("t", "Unipi", "airtable table name")
// 	key                 = flag.String("k", "YOUR_API_KEY_HERE", "airtable API key")
// 	restartScript       = "/usr/local/bin/restart_change_ffjs.sh"
// 	AUTO_UPDATER_SCRIPT = "/usr/local/bin/auto_updater_ffjs.sh"
// 	LOG_FILE            = "/var/log/auto_updater_ffjs.log"
// 	SSH_KEY_PATH        = "/home/user/.ssh/id_rsa"
// )

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(utility.Red, ".env file wasn't found, looking at env variables", utility.Reset)
		return
	}

	// parsing flags to access them later
	config.Init()
	// check if the user is root user or not
	if !utility.CheckForElevatedPriveleges() {
		return
	}
	// Check if any flags are passed and are they valid
	if !utility.CheckFlags(os.Args) {
		return
	}

	// fmt.Println(utility.BrightGreen + "Starting install... " + utility.Reset)

	// cmdNPM := exec.Command("apt-cache", "madison", "nodejs")
	// outputNPM, err := cmdNPM.Output()
	// output := string(outputNPM)
	// parts := strings.Split(output, "\n")
	// latestAvailableVersion := strings.Split(parts[0], "|")
	// log.Println("output: ", latestAvailableVersion[1])

	// opening log file
	utility.LogFile = utility.OpenLogFile()

	// ensuring neccesary tools installed
	fmt.Println(utility.Yellow, "calling ensureInstalled() for [inotify-tools,git,jq,nano].....", utility.Reset)

	utility.EnsureInstalled("inotify-tools")
	utility.EnsureInstalled("git")
	utility.EnsureInstalled("jq")
	utility.EnsureInstalled("nano")

	// adding a waitgroup here
	// Increment the wait group for each goroutine
	utility.Wg.Add(1)
	go gitops.CheckGitAccessAndCloneIfAccess()

	// Wait for all goroutines to finish
	utility.Wg.Wait()

}
