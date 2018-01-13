package main

//#cgo LDFLAGS: -ltdjson
//#include <stdlib.h>
//#include <td/telegram/td_json_client.h>
//#include <td/telegram/td_log.h>
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

type Client struct {
	Client    unsafe.Pointer
	Updates   chan map[string]interface{}
	Callbacks chan map[string]interface{}
}

func NewClient() *Client {
	client := Client{Client: C.td_json_client_create()}
	updates := make(chan map[string]interface{})
	callbacks := make(chan map[string]interface{})

	go func() {
		for {
			event := client.Receive(10)

			var update map[string]interface{}
			json.Unmarshal([]byte(event), &update)

			if updateExtra, ok := update["@extra"].(string); ok {
				fmt.Println("sending into channel:", updateExtra)
				callbacks <- update
			}

			if _, ok := update["@type"].(string); ok {
				updates <- update
			} else {
				fmt.Println("update without @type field")
			}
		}
	}()

	client.Updates = updates
	client.Callbacks = callbacks
	return &client
}

func (c *Client) Destroy() {
	C.td_json_client_destroy(c.Client)
}

func (c *Client) Send(jsonQuery string) {
	query := C.CString(jsonQuery)
	defer C.free(unsafe.Pointer(query))

	C.td_json_client_send(c.Client, query)
}

func (c *Client) Receive(timeout float64) string {
	result := C.td_json_client_receive(c.Client, C.double(timeout))
	return C.GoString(result)
}

func (c *Client) Execute(jsonQuery string) string {
	query := C.CString(jsonQuery)
	defer C.free(unsafe.Pointer(query))

	result := C.td_json_client_execute(c.Client, query)
	return C.GoString(result)
}

func SetFilePath(path string) {
	query := C.CString(path)
	defer C.free(unsafe.Pointer(query))

	C.td_set_log_file_path(query)
}

func SetLogVerbosityLevel(level int) {
	C.td_set_log_verbosity_level(C.int(level))
}

func (c *Client) SendAndCatch(jsonQuery string) (map[string]interface{}, error) {
	var jsonWithoutExtra map[string]interface{}
	json.Unmarshal([]byte(jsonQuery), &jsonWithoutExtra)

	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	//rand.Seed(time.Now().UnixNano())
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	randomString := string(b)

	jsonWithoutExtra["@extra"] = randomString
	jsonWithExtra, _ := json.Marshal(&jsonWithoutExtra)

	c.Send(string(jsonWithExtra))

	fmt.Printf("callbacks: %+v\n", c.Callbacks)

	select {
	case callback := <-c.Callbacks:
		if updateExtra, ok := callback["@extra"].(string); ok {
			if updateExtra == randomString {
				fmt.Println("catched")
				return callback, nil
			} else {
				fmt.Println("catched, but not equal")
				return map[string]interface{}{}, errors.New("catched, but not equal")
			}
		} else {
			return map[string]interface{}{}, errors.New("there is no @extra field")
		}
	case <-time.After(10 * time.Second):
		fmt.Println("timeout")
		return map[string]interface{}{}, errors.New("timeout")
	}
}
