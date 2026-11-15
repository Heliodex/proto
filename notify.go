package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	c "github.com/TwiN/go-color"
)

type (
	Notification struct {
		Address int
		Type, Test string
	}

	SendNotification struct {
		Address int
		Type string
		To int
	}
)

const addr = "http://localhost:3000/notify"

func sendNotification(data any) {
	w := &bytes.Buffer{}

	json.NewEncoder(w).Encode(data)
	if _, err := http.Post(addr, "application/json", w); err != nil {
		fmt.Println(c.InRed("Failed to notify"), err)
	}
}

func notify(Type string) {
	sendNotification(Notification{
		Address: lport - 10000,
		Type: Type,
		Test: "hello",
	})
}

func notifySend(To int) {
	sendNotification(SendNotification{
		Address: lport - 10000,
		Type: "Send",
		To: To - 10000,
	})
}
