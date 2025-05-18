package main

import (
	"encoding/json"
	"os"

	"github.com/kovidgoyal/dbus"
	"github.com/kovidgoyal/dbus/introspect"
)

func main() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	node, err := introspect.Call(conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus"))
	if err != nil {
		panic(err)
	}
	data, _ := json.MarshalIndent(node, "", "    ")
	os.Stdout.Write(data)
}
