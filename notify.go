package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	c "github.com/TwiN/go-color"
)

const addr = "http://localhost:3000/notify"

type Notification struct {
	Address int
	Type    string
	Test    string
}

func notify(Address int, Type string) {
	w := new(bytes.Buffer)
	enc := json.NewEncoder(w)

	enc.Encode(Notification{
		Address: Address - 10000,
		Type:    Type,
		Test:    "hello",
	})

	if _, err := http.Post(addr, "application/json", w); err != nil {
		fmt.Println(c.InRed("Failed to notify"), err)
	}
}
