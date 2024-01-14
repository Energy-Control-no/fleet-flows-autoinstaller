# Fleet Flows Autoinstaller

## Overview

The Fleet Flows Autoinstaller is a powerful and easy-to-use script designed to streamline the installation of the Fleet Flow application. It automates the setup process, ensuring a quick and efficient deployment.

## Prerequisites

Before running the autoinstaller, ensure that you have:
- An active internet connection
- Your Airtable API Key with data.records:read & write scopes. For the METADATA base

## Installation

To install and run the Fleet Flows Autoinstaller, execute the following command:

```bash
curl -sSL https://raw.githubusercontent.com/Energy-Control-no/fleet-flows-autoinstaller/main/full-install.sh | bash -s -- [Your Airtable API Key] 
``` 
## Features

- **Automated Installation**: Quickly sets up Fleet Flows without manual intervention.
- **Secure**: Implements best practices to ensure a secure installation.
- **Customizable**: Easily adaptable to accommodate different environments and configurations.

## How It Works

### The script performs the following actions:

- Validates the provided Airtable API Key.
- Downloads and installs necessary dependencies and packages.
- Configures the Fleet Flows environment based on the provided API Key.
- Performs system checks to ensure successful installation.

## Troubleshooting

### If you encounter any issues during the installation process, please check the following:

- Ensure that your Airtable API Key is correct.
- Verify that your server meets all the prerequisites.
- Check your internet connection.

For further assistance, [create an issue](https://github.com/Energy-Control-no/fleet-flows-autoinstaller/issues) on the GitHub repository.

## License

This project is licensed under the [MIT License](LICENSE).
