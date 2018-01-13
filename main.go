package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	SetLogVerbosityLevel(1)
	//SetFilePath("./errors.txt")

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
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		client.Destroy()
		os.Exit(1)
	}()

	// Func for quick marshaling map into string
	marshal := func(jsonQuery map[string]interface{}) string {
		jsonBytes, _ := json.Marshal(jsonQuery)
		return string(jsonBytes)
	}

	rand.Seed(time.Now().UnixNano())

	for update := range client.Updates {
		// Show all updates in JSON
		fmt.Println(update)

		// Authorization block
		if update["@type"].(string) == "updateAuthorizationState" {
			if authorizationState, ok := update["authorization_state"].(map[string]interface{})["@type"].(string); ok {
				switch authorizationState {
				case "authorizationStateWaitTdlibParameters":
					client.SendAndCatch(marshal(map[string]interface{}{
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
					client.SendAndCatch(marshal(map[string]interface{}{
						"@type": "checkDatabaseEncryptionKey",
					}))
				case "authorizationStateWaitPhoneNumber":
					fmt.Print("Enter phone: ")
					var number string
					fmt.Scanln(&number)

					client.SendAndCatch(marshal(map[string]interface{}{
						"@type":        "setAuthenticationPhoneNumber",
						"phone_number": number,
					}))
				case "authorizationStateWaitCode":
					fmt.Print("Enter code: ")
					var code string
					fmt.Scanln(&code)

					client.SendAndCatch(marshal(map[string]interface{}{
						"@type": "checkAuthenticationCode",
						"code":  code,
					}))
				case "authorizationStateWaitPassword":
					fmt.Print("Enter password: ")
					var passwd string
					fmt.Scanln(&passwd)

					client.SendAndCatch(marshal(map[string]interface{}{
						"@type":    "checkAuthenticationPassword",
						"password": passwd,
					}))
				}
			}
		}
	}
}
