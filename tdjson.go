package main

//#cgo LDFLAGS: -ltdjson
//#include <stdlib.h>
//#include <td/telegram/td_json_client.h>
//#include <td/telegram/td_log.h>
import "C"

import (
	"unsafe"
)

// Client is an instance of tdlib's JSON client.
type Client struct {
	client unsafe.Pointer
}

// NewClient creates a new tdlib JSON client.
func NewClient() *Client {
	return &Client{
		client: C.td_json_client_create(),
	}
}

func SetLogVerbosityLevel(level int) {
	C.td_set_log_verbosity_level(C.int(level))
}

func SetFilePath(path string) {
	query := C.CString(path)
	defer C.free(unsafe.Pointer(query))

	C.td_set_log_file_path(query)
}

func (c *Client) Send(jsonQuery string) {
	query := C.CString(jsonQuery)
	defer C.free(unsafe.Pointer(query))

	C.td_json_client_send(c.client, query)
}

func (c *Client) Receive(timeout float64) string {
	result := C.td_json_client_receive(c.client, C.double(timeout))
	return C.GoString(result)
}

func (c *Client) Execute(jsonQuery string) string {
	query := C.CString(jsonQuery)
	defer C.free(unsafe.Pointer(query))

	result := C.td_json_client_execute(c.client, query)
	return C.GoString(result)
}

func (c *Client) Destroy() {
	C.td_json_client_destroy(c.client)
}
