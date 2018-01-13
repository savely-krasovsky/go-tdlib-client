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
	marshal := func(jsonQuery Update) string {
		jsonBytes, _ := json.Marshal(jsonQuery)
		return string(jsonBytes)
	}

	rand.Seed(time.Now().UnixNano())

	for update := range client.Updates {
		// Show all updates in JSON
		fmt.Println(update)

		// Authorization block
		if update["@type"].(string) == "updateAuthorizationState" {
			if authorizationState, ok := update["authorization_state"].(Update)["@type"].(string); ok {
				switch authorizationState {
				case "authorizationStateWaitTdlibParameters":
					res, err := client.SendAndCatch(marshal(Update{
						"@type": "setTdlibParameters",
						"parameters": Update{
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
					if err != nil {
						log.Panic(err)
					}
					log.Println(res)
				case "authorizationStateWaitEncryptionKey":
					res, err := client.SendAndCatch(marshal(Update{
						"@type": "checkDatabaseEncryptionKey",
					}))
					if err != nil {
						log.Panic(err)
					}
					log.Println(res)
				case "authorizationStateWaitPhoneNumber":
					fmt.Print("Enter phone: ")
					var number string
					fmt.Scanln(&number)

					res, err := client.SendAndCatch(marshal(Update{
						"@type":        "setAuthenticationPhoneNumber",
						"phone_number": number,
					}))
					if err != nil {
						log.Panic(err)
					}
					log.Println(res)
				case "authorizationStateWaitCode":
					fmt.Print("Enter code: ")
					var code string
					fmt.Scanln(&code)

					res, err := client.SendAndCatch(marshal(Update{
						"@type": "checkAuthenticationCode",
						"code":  code,
					}))
					if err != nil {
						log.Panic(err)
					}
					log.Println(res)
				case "authorizationStateWaitPassword":
					fmt.Print("Enter password: ")
					var passwd string
					fmt.Scanln(&passwd)

					res, err := client.SendAndCatch(marshal(Update{
						"@type":    "checkAuthenticationPassword",
						"password": passwd,
					}))
					if err != nil {
						log.Panic(err)
					}
					log.Println(res)
				}
			}
		}
	}
}
