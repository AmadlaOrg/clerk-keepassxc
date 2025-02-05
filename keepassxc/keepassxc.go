package keepassxc

import (
	"bytes"
	"fmt"
	"github.com/godbus/dbus/v5"
	"log"
	"os/exec"
	"strings"
)

type IKeepassxc interface{}
type SKeepassxc struct{}

func (s *SKeepassxc) Key() string {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Println("Failed to connect to D-Bus:", err)
		return ""
	}
	defer func(conn *dbus.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Failed to close D-Bus:", err)
		}
	}(conn)

	obj := conn.Object("org.freedesktop.secrets", "/org/freedesktop/secrets")

	// Example: Get collection or item (Replace with actual method)
	var result string
	err = obj.Call("org.freedesktop.Secret.Service.OpenSession", 0, "plain", "").Store(&result)
	if err != nil {
		fmt.Println("Error calling D-Bus:", err)
		return ""
	}

	fmt.Println("D-Bus Response:", result)

	return ""
}

func (s *SKeepassxc) SecondOpt() {
	// Connect to the session D-Bus
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Fatalf("Failed to connect to D-Bus: %v", err)
	}
	defer func(conn *dbus.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Failed to close D-Bus:", err)
		}
	}(conn)

	// Define the KeePassXC D-Bus service and object
	service := "org.keepassxc.KeePassXC.MainWindow"
	objectPath := "/org/freedesktop/secrets" //"/org/keepassxc/KeePassXC"

	// Get the main object for KeePassXC
	obj := conn.Object(service, dbus.ObjectPath(objectPath))

	// Check if KeePassXC is running
	var isOpen bool
	err = obj.Call("org.keepassxc.KeePassXC.DatabaseOpened", 0).Store(&isOpen)
	if err != nil {
		log.Println("KeePassXC is not open or database is not unlocked.")
		return
	}

	fmt.Println("KeePassXC is running and database is open.")

	// List available entries in the database
	var entries []string
	err = obj.Call("org.keepassxc.KeePassXC.GetDatabaseEntries", 0).Store(&entries)
	if err != nil {
		log.Fatalf("Failed to retrieve database entries: %v", err)
	}

	fmt.Println("Entries in database:", entries)

	// Retrieve the secret for "MySecret"
	var password string
	err = obj.Call("org.keepassxc.KeePassXC.GetDatabaseEntryPassword", 0, "MySecret").Store(&password)
	if err != nil {
		log.Fatalf("Failed to retrieve password: %v", err)
	}

	fmt.Println("Retrieved password:", password)
}

func (s *SKeepassxc) RetrieveSecret(entryName string) (string, error) {
	dbPath := "/tmp/test-db.kdbx"
	keyFilePath := "/tmp/test-key.key"

	cmd := exec.Command(
		"keepassxc-cli",
		"show",
		"--no-password",
		"--key-file",
		keyFilePath,
		"--attributes",
		"Password",
		dbPath,
		entryName)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error retrieving secret: %v", err)
	}

	password := strings.TrimSpace(out.String())
	return password, nil
}
