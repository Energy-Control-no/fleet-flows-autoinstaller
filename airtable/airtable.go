package airtable

import (
	"bytes"
	"encoding/json"
	"fmt"
	config "installer/configs"
	"installer/utility"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// contains all the airtable functions

// Checks if provided AIRTABLE_API_KEY is valid or not
func CheckAirtableAPIKey(apiKey string, AIRTABLE_BASE_ID string, AIRTABLE_TABLE_NAME string) bool {
	maxTries, _ := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", AIRTABLE_BASE_ID, AIRTABLE_TABLE_NAME)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utility.Logger(err, utility.Error)
		fmt.Println("Error creating request:", err)
		return false
	}
	log.Println("url: ", url)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Added Timeout property to 10 seconds if server takes longer than this our program will exit
	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	for {
		resp, err = client.Do(req)
		if err != nil {
			if maxTries == 0 {
				utility.ErrorLog.Output(2, utility.Error+"Max Retries for request exceeded for verifying airtable key")
				fmt.Println(utility.Red, "Max retries exceeded...", err, utility.Reset)
				break
			}
			fmt.Println(utility.BrightYellow, "Error sending request: retrying..", utility.Reset)
			maxTries -= 1
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
	}
	defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("Error reading response body:", err)

	// }

	// // Print the response body as a string
	// fmt.Println("Response Body:", string(body))
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Invalid API key. Stopping the script.")
		return false
	}

	fmt.Println("API key is valid.")
	return true
}

// Updates SSHKey in airtable
func UpdateAirtableRecord(recordID, sshKey string) {
	maxTries, _ := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	// Construct the JSON payload
	payload := map[string]interface{}{
		"fields": map[string]string{
			"SSH Public Key": sshKey,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(utility.Red, "Cannot marshal payload to update SSH Public Key in Airtable", utility.Reset)
	}

	// Create HTTP PATCH request
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", *config.Base, *config.Table, recordID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatal(utility.Red, "Cannot form new http request to update SSH Public Key in Airtable", utility.Reset)
	}

	// Set request headers
	req.Header.Set("Authorization", "Bearer "+*config.Key)
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	// Added Timeout property to 10 seconds if server takes longer than this our program will exit
	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	for {
		resp, err = client.Do(req)
		if err != nil {
			if maxTries == 0 {
				utility.ErrorLog.Output(2, utility.Error+"Max Retries for request exceeded for updating airtable Record")
				fmt.Println(utility.Red, "Max retries exceeded...", err, utility.Reset)
				break
			}
			fmt.Println(utility.BrightYellow, "Error sending request: retrying..", utility.Reset)
			maxTries -= 1
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Fatal(utility.Red, "update failed with status code ", resp.StatusCode, utility.Reset)
	}
}

// creates a new airtable record
func CreateAirtableRecord(hostname, sshKey string) error {
	maxTries, _ := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	payload := map[string]interface{}{
		"records": []map[string]interface{}{
			{
				"fields": map[string]string{
					"SSH Public Key": sshKey,
					"type":           strings.Split(hostname, "-")[0],
					"Unipi SN":       strings.Split(hostname, "-")[1],
					// "type":           "TEST",  // test
					// "Unipi SN":       "SN10",  // test
				},
			},
		},
		"typecast": true, // set true for auto options addition
	}
	// log.Println("payload: ", payload)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create HTTP POST request
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", *config.Base, *config.Table)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	// Set request headers
	req.Header.Set("Authorization", "Bearer "+*config.Key)
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	// Added Timeout property to 10 seconds if server takes longer than this our program will exit
	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	for {
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if maxTries == 0 {
				utility.Logger(err, utility.Error)
				fmt.Println(utility.Red, "Max retries exceeded...", utility.Reset)
				break
			}
			fmt.Println(utility.BrightYellow, "Error sending request: retrying..", utility.Reset)
			maxTries -= 1
		}

		if resp.StatusCode == http.StatusOK {
			break
		}
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Print the response body as a string
	// fmt.Println("Response Body:", string(body))
	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Fatal(utility.Red, "create record failed with status code ", resp.StatusCode, utility.Reset)
	}

	return nil
}

// Fetches Record Id by SSHKey that is passed
func FetchAirtableRecordIDBySSHKey(SSHKey string) (string, error) {
	maxTries, _ := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	// Encode the hostname to use in the URL
	encodedSSHKey := url.QueryEscape(SSHKey)

	// Construct the URL with the encoded hostname and filterByFormula parameter
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s?filterByFormula={SSH%%20Public%%20Key}%%20=%%20'%s'", *config.Base, *config.Table, encodedSSHKey)

	// Create a new HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Set headers, including the Authorization header with the API key
	req.Header.Set("Authorization", "Bearer "+*config.Key)
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	// Added Timeout property to 10 seconds if server takes longer than this our program will exit
	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	for {
		resp, err = client.Do(req)
		if err != nil {
			if maxTries == 0 {
				utility.Logger(err, utility.Error)
				fmt.Println(utility.Red, "Max retries exceeded...", err, utility.Reset)
				break
			}
			fmt.Println(utility.BrightYellow, "Error sending request: retrying..", utility.Reset)
			maxTries -= 1
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse the JSON response
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	// Extract the record ID from the response
	records := data["records"].([]interface{})
	if len(records) > 0 {
		record := records[0].(map[string]interface{})
		return record["id"].(string), nil
	}

	// If no record found, return an empty string
	return "", nil
}
func FetchAirtableRecordIDByDeviceId(deviceId string) (string, error) {
	maxTries, _ := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	// for test
	// deviceId = "TEST-SN10"
	// Encode the hostname to use in the URL
	part2 := strings.Split(deviceId, "-")[1]
	if len(part2) <= 2 {
		part2 = "sn" + part2
	}
	if len(strings.Split(deviceId, "-")) > 2 {
		deviceId = strings.Split(deviceId, "-")[0] + "-" + part2
	}
	encodedSSHKey := url.QueryEscape(deviceId)

	// Construct the URL with the encoded hostname and filterByFormula parameter
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s?filterByFormula={Device%%20id}%%20=%%20'%s'", *config.Base, *config.Table, encodedSSHKey)

	// Create a new HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Set headers, including the Authorization header with the API key
	req.Header.Set("Authorization", "Bearer "+*config.Key)
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	// Added Timeout property to 10 seconds if server takes longer than this our program will exit
	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	for {
		resp, err = client.Do(req)
		if err != nil {
			if maxTries == 0 {
				utility.Logger(err, utility.Error)
				fmt.Println(utility.Red, "Max retries exceeded...", err, utility.Reset)
				break
			}
			fmt.Println(utility.BrightYellow, "Error sending request: retrying..", utility.Reset)
			maxTries -= 1
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Parse the JSON response
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	// Extract the record ID from the response
	records := data["records"].([]interface{})
	if len(records) > 0 {
		record := records[0].(map[string]interface{})
		return record["id"].(string), nil
	}

	// If no record found, return an empty string
	return "", nil
}

// Updates SSHKey in AIRTABLE
func UpdateSSHKeyInAirtable() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "Error getting user home dir: ", err, utility.Reset)
	}
	publicKeyPath := config.SSHKeyPath + ".pub"
	publicKeyPath = filepath.Join(homeDir, publicKeyPath)
	// Read the contents of the SSH public key file
	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		utility.Logger(err, utility.Error)
		fmt.Println("Error reading public key file:", err)
		return
	}
	// Convert bytes to string
	publicKey := string(publicKeyBytes)
	hostname, _ := os.Hostname()
	// recordId, err := FetchAirtableRecordIDBySSHKey(publicKey)
	recordId, err := FetchAirtableRecordIDByDeviceId(hostname)
	if err != nil {
		utility.Logger(err, utility.Error)
		log.Fatal(utility.Red, "Error fetching record id: ", err, utility.Reset)
	}
	if recordId != "" {
		// updating existing record
		log.Println("calling updateAirtableRecord().............")
		UpdateAirtableRecord(recordId, publicKey)
	} else {
		// create new record
		log.Println("calling createAirtableRecord().............")
		err := CreateAirtableRecord(hostname, publicKey)
		if err != nil {
			utility.Logger(err, utility.Error)
			log.Fatal(utility.Red, "Error creating airtable record: ", err, utility.Reset)
		}
	}

}
