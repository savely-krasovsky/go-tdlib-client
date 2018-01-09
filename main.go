package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Structs for auth
type Response struct {
	Type               string             `json:"@type"`
	AuthorizationState authorizationState `json:"authorization_state"`
}

type authorizationState struct {
	Type string `json:"@type"`
}

func main() {
	SetLogVerbosityLevel(1)
	SetFilePath("./errors.txt")

	// Get API_ID and API_HASH from env vars
	apiId := os.Getenv("API_ID")
	if apiId == "" {
		log.Fatal("API_ID env variable not specified")
	}
	apiHash := os.Getenv("API_HASH")
	if apiHash == "" {
		log.Fatal("API_HASH env variable not specified")
	}

	client := NewClient()

	// Handle Ctrl+C
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		client.Destroy()
		os.Exit(1)
	}()

	// Func for quick marshaling map into string
	marshal := func(jsonQuery map[string]interface{}) string {
		jsonBytes, _ := json.Marshal(jsonQuery)
		return string(jsonBytes)
	}

	for {
		// Get all updates here
		event := client.Receive(10)

		// Decode event to Response struct for auth
		var res Response
		json.Unmarshal([]byte(event), &res)

		// Main auth switch mechanism
		switch res.AuthorizationState.Type {
		case "authorizationStateWaitTdlibParameters":
			client.Send(marshal(map[string]interface{}{
				"@type": "setTdlibParameters",
				"parameters": map[string]interface{}{
					"@type":                    "tdlibParameters",
					"use_message_database":     true,
					"api_id":                   apiId,
					"api_hash":                 apiHash,
					"system_language_code":     "en",
					"device_model":             "Server",
					"system_version":           "Unknown",
					"application_version":      "1.0",
					"enable_storage_optimizer": true,
				},
			}))
		case "authorizationStateWaitEncryptionKey":
			client.Send(marshal(map[string]interface{}{
				"@type": "checkDatabaseEncryptionKey",
			}))
		case "authorizationStateWaitPhoneNumber":
			fmt.Print("Enter phone: ")
			var number string
			fmt.Scanln(&number)

			client.Send(marshal(map[string]interface{}{
				"@type":        "setAuthenticationPhoneNumber",
				"phone_number": number,
			}))
		case "authorizationStateWaitCode":
			fmt.Print("Enter code: ")
			var code string
			fmt.Scanln(&code)

			client.Send(marshal(map[string]interface{}{
				"@type": "checkAuthenticationCode",
				"code":  code,
			}))
		case "authorizationStateWaitPassword":
			fmt.Print("Enter password: ")
			var passwd string
			fmt.Scanln(&passwd)

			client.Send(marshal(map[string]interface{}{
				"@type":    "checkAuthenticationPassword",
				"password": passwd,
			}))
		}

		// Show all events in json
		fmt.Println(event)
	}
}
