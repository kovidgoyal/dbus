package main

import (
	"fmt"
	"os"

	"github.com/kovidgoyal/dbus"
)

func main() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to SystemBus bus:", err)
		os.Exit(1)
	}
	defer conn.Close()

	var s string
	err = conn.Object("org.bluez", "/").Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to introspect bluez", err)
		os.Exit(1)
	}

	fmt.Println(s)
}
