package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Парочка структур, чтобы поймать ожидание кода
type Response struct {
	Type               string             `json:"@type"`
	AuthorizationState authorizationState `json:"authorization_state"`
}

type authorizationState struct {
	Type string `json:"@type"`
}

func main() {
	client := NewClient()

	// Обрабатываем Ctrl+C
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		client.Destroy()
		os.Exit(1)
	}()

	SetLogVerbosityLevel(1)

	for {
		// Получаем тут асинхронные ивенты после выполнения TdSend()
		event := client.Receive(10)

		// Декодируем полученный JSON для авторизации
		var res Response
		_ = json.Unmarshal([]byte(event), &res)

		// Если либа ждет номер или код -- ловим
		switch res.AuthorizationState.Type {
		case "authorizationStateWaitTdlibParameters":
			// Отправляем api_id и api_hash
			client.Send(`{
  "@type": "setTdlibParameters",
  "parameters": {
    "@type": "tdlibParameters",
    "use_message_database": true,
    "api_id": 15504,
    "api_hash": "1c7e91c33b0a9c1f4379a811f0826509",
    "system_language_code": "en",
    "device_model": "Server",
    "system_version": "Unknown",
    "application_version": "1.0",
    "enable_storage_optimizer": true
  }
}`)
		case "authorizationStateWaitEncryptionKey":
			// Проверяем ключик
			client.Send(`{
  "@type": "checkDatabaseEncryptionKey"
}`)
		case "authorizationStateWaitPhoneNumber":
			// Устанавливаем номер
			fmt.Print("Enter phone: ")
			var number string
			fmt.Scanln(&number)

			client.Send(fmt.Sprintf(`{
  "@type": "setAuthenticationPhoneNumber",
  "phone_number": "%s"
}`, number))
		case "authorizationStateWaitCode":
			// Вводим код
			fmt.Print("Enter code: ")
			var code string
			fmt.Scanln(&code)

			client.Send(fmt.Sprintf(`{
  "@type": "checkAuthenticationCode",
  "code": "%s"
}`, code))
		case "authorizationStateWaitPassword":
			// Вводим пароль, если есть
			fmt.Print("Enter password: ")
			var passwd string
			fmt.Scanln(&passwd)

			client.Send(fmt.Sprintf(`{
  "@type": "checkAuthenticationCode",
  "code": "%s"
}`, passwd))
		}

		// Выводим полученные JSON
		fmt.Println(event)
	}
}
