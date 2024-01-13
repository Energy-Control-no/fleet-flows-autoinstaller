Fleet Flow JS Auto Installer

This repository contains scripts for setting up and configuring the environment for the fleet-flow-js project. It automates system updates, dependency installations, repository cloning, and service configurations.
Scripts

    system_update_and_dependencies.sh: This script updates the system, installs necessary dependencies, and configures the environment for Node.js and Node-RED.
    repository_cloning_and_service_setup.sh: Handles the cloning of the fleet-flow-js repository and sets up various services and listeners.

Getting Started
Prerequisites

    A Linux-based operating system (preferably Ubuntu or Debian)
    Git and Curl installed
    Sudo privileges for the executing user

Installation

    Clone the fleet-flow-js auto installer repository:

    bash

git clone [repository-url]

Navigate to the cloned directory:

bash

cd [repository-name]

Make the scripts executable:

bash

    chmod +x system_update_and_dependencies.sh
    chmod +x repository_cloning_and_service_setup.sh

Usage

    System Update and Dependencies Installation:

    Run system_update_and_dependencies.sh with the Airtable API key as an argument:

    bash

./system_update_and_dependencies.sh <AIRTABLE_API_KEY>

This script will update the system, install required packages, and configure the environment.

Repository Cloning and Service Setup:

Execute repository_cloning_and_service_setup.sh:

bash

    ./repository_cloning_and_service_setup.sh

    This script clones the fleet-flow-js repository and sets up systemd services and file change listeners.

Configuration

    Adjust the scripts according to your fleet-flow-js project requirements.
    Update constants, paths, and other configurations as necessary.
    Ensure the correct Airtable API key, Git server URL, and branch names are set for your project.

Contributions

Contributions to the fleet-flow-js auto installer are welcome. Please adhere to standard Git practices for submitting pull requests or opening issues.