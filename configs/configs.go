// config/config.go

package config

import (
	"flag"
	"os"
)

var (
	Repository        = flag.String("r", "", "sets the location of the repository")
	SoftwareBranch    = flag.String("sb", "", "branch for fleet-flows-js")
	FilesBranch       = flag.String("fb", "", "branch for fleet-files")
	Base              = flag.String("b", "", "airtable base id")
	Table             = flag.String("t", "", "airtable table name")
	Key               = flag.String("k", "", "airtable API key")
	SchemaFilePath    = flag.String("sf", "", "Schema File Path")
	RestartScript     = "/usr/local/bin/restart_change_ffjs.sh"
	AutoUpdaterScript = "/usr/local/bin/auto_updater_ffjs.sh"
	LogFile           = "/var/log/auto_updater_ffjs.log"
	SSHKeyPath        = ".ssh/id_rsa"
)

func Init() {
	flag.Parse()
	if *Repository == "" {
		*Repository = os.Getenv("GIT_SERVER")
	}
	if *SoftwareBranch == "" {
		*SoftwareBranch = os.Getenv("FLOW_JS_BRANCH")
	}
	if *FilesBranch == "" {
		*FilesBranch = os.Getenv("FILES_BRANCH")
	}
	if *Base == "" {
		*Base = os.Getenv("AIRTABLE_BASE_ID")
	}
	if *Table == "" {
		*Table = os.Getenv("AIRTABLE_TABLE")
	}
	if *Key == "" {
		*Key = os.Getenv("AIRTABLE_API_KEY")
	}
	if *SchemaFilePath == "" {
		*SchemaFilePath = os.Getenv("SCHEMA_FILE_PATH")
	}

}
