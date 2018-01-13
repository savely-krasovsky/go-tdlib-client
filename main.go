package main

import (
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

	// Create new instance of client
	client := NewClient()

	// Handle Ctrl+C
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		client.Destroy()
		os.Exit(1)
	}()

	// Seed rand with time
	rand.Seed(time.Now().UnixNano())

	for update := range client.Updates {
		// Show all updates in JSON
		fmt.Println(update)

		// Authorization block
		if update["@type"].(string) == "updateAuthorizationState" {
			if authorizationState, ok := update["authorization_state"].(Update)["@type"].(string); ok {
				res, err := client.Auth(authorizationState, apiId, apiHash)
				if err != nil {
					log.Println(err)
				}
				log.Println(res)
			}
		}
	}
}
