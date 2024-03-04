## Usecase:
> This executable is an auto-installer for multiple services on a single IOT device, which monitors remote repoitories and keeps the local repositories in sync with it by pulling updates as and when found.

## How to use it:
> It is an executable Go file which runs through CLI and performs actions based on `flags` given to it during run time.

> Yup something like `-b,-c` etc...

## To start the autoinstaller:
- Open your Terminal or Command Prompt.
- `cd` to the directroy where the executable resides.
- Enter the name of the executable for example:
    > *`sudo`* ./installer `-b="this is the value for the flag for this run" && apt install ca-certificates && apt install tzdata`
- Hit enter!

**use `sudo` as this program requires elevated privileges.*

## Flags that are supported
> `-r`  :  Repository url --- type: *`string`*  
`default: `ssh://git@fleet-flows-git.lizzardsolutions.com 

> `-sb` :  Software branch --- type: *`string`*    
`default: ` main

> `-fb` :  Files Branch  --- type: *`string`*    
`default: ` main 

>  `-b` : Airtable base Id/name  --- type: *`string`*  
`default: ` appYWVOaoPhQB0nmA

>  `-t` : Airtable name   --- type: *`string`*  
`default: ` Unipi

>  `-k` : Airtable API key/token  --- type: *`string`*  
`default: ` YOUR_API_KEY_HERE

>  `-sf` : Schema file path  --- type: *`string`*  
`default: ` schema.yml

>  `-n` : Node version  --- type: *`string`*  
`default: ` 18.19.1-1nodesource1

## Dir for service & log files
- After successful installation the following services will be started in the users system and there paths are defined below:
#### `fleet-flows-js.service `
> /etc/systemd/system/fleet-flows-js.service
#### `restart_change_ffjs.sh`
> /usr/local/bin/restart_change_ffjs.sh
#### `fleet-flows-js-listener.service `
> /etc/systemd/system/fleet-flows-js-listener.service
#### `auto_updater_ffjs.sh`
> /usr/local/bin/auto_updater_ffjs.sh
#### `auto_updater_ffjs.log`
> /var/log/auto_updater_ffjs.log

## Logging

> All the errors during installation will be logged under `Errorlogged.txt` which will reside exactly where the executable is. 

## Pre-requisites
> apt install ca-certificates -y
> apt install tzdata