/*package keepassxc

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

func Test_integration_Key(t *testing.T) {
	service := &SKeepassxc{}

	createDbCmd := exec.Command("keepassxc-cli", "db-create", "/tmp/test-db.kdbx")

	var out bytes.Buffer
	createDbCmd.Stdout = &out

	err := createDbCmd.Run()
	if err != nil {
		t.Fatalf("error creating db: %v", err)
		return
	}

	addSecretCmd := exec.Command(
		"keepassxc-cli",
		"add",
		"/tmp/test-db.kdbx",
		`"MySecret"`,
		"--password",
		`"supersecure"`,
		"--username",
		`"admin"`)

	addSecretCmd.Stdout = &out

	err = addSecretCmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	service.SecondOpt()
}*/

package keepassxc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func Test_integration_Key(t *testing.T) {
	service := &SKeepassxc{}

	// Define paths
	dbPath := "/tmp/test-db.kdbx"
	keyFilePath := "/tmp/test-key.key"

	// Step 1: Generate a key file
	keyData := []byte("random_test_key_content")
	err := os.WriteFile(keyFilePath, keyData, 0600)
	if err != nil {
		t.Fatalf("failed to create key file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("failed to remove key file: %v", err)
		}

		err = os.RemoveAll(dbPath)
		if err != nil {
			t.Fatalf("failed to remove key file: %v", err)
		}
	}(keyFilePath) // Clean up after test

	// Step 2: Create a new KeePassXC database with the key file
	createDbCmd := exec.Command("keepassxc-cli", "db-create", "--set-key-file", keyFilePath, dbPath)

	var out bytes.Buffer
	createDbCmd.Stdout = &out

	err = createDbCmd.Run()
	if err != nil {
		t.Fatalf("error creating db: %v", err)
	}

	// Step 3: Add an entry to the database with a password
	addSecretCmd := exec.Command(
		"keepassxc-cli",
		"add",
		"--no-password",
		"--key-file", keyFilePath,
		dbPath,
		"MySecret",
		"--username", "admin",
		"--password-prompt", // Use prompt mode for password entry
	)

	// Pipe the password input
	addSecretCmd.Stdin = bytes.NewBufferString("supersecure\n")

	addSecretCmd.Stdout = &out
	err = addSecretCmd.Run()
	if err != nil {
		t.Fatalf("error adding secret: %v", err)
	}

	fmt.Println("Integration test completed successfully.")

	// Call your service function
	//service.SecondOpt()
	secret, err := service.RetrieveSecret("MySecret")
	if err != nil {
		t.Fatalf("error retrieving secret: %v", err)
	}
	println(secret)
	//assert.NotNil(t, secret)
}

/*func Test_integration_RetrieveSecret(t *testing.T) {
	service := &SKeepassxc{}

	// Define paths
	dbPath := "/tmp/test-db.kdbx"
	keyFilePath := "/tmp/test-key.key"

	// Step 1: Generate a key file
	keyData := []byte("random_test_key_content")
	err := os.WriteFile(keyFilePath, keyData, 0600)
	if err != nil {
		t.Fatalf("failed to create key file: %v", err)
	}
	defer func() {
		_ = os.Remove(keyFilePath)
		_ = os.RemoveAll(dbPath)
	}()

	// Step 2: Create a new KeePassXC database with the key file
	createDbCmd := exec.Command("keepassxc-cli", "db-create", "--set-key-file", keyFilePath, "--no-password", dbPath)

	var out bytes.Buffer
	createDbCmd.Stdout = &out

	err = createDbCmd.Run()
	if err != nil {
		t.Fatalf("error creating db: %v", err)
	}

	// Step 3: Add an entry to the database with a password
	addSecretCmd := exec.Command(
		"keepassxc-cli",
		"add",
		"--no-password",
		"--key-file", keyFilePath,
		dbPath,
		"MySecret",
		"--username", "admin",
		"--password-prompt",
	)

	// Pipe the password input
	addSecretCmd.Stdin = bytes.NewBufferString("supersecure\n")

	addSecretCmd.Stdout = &out
	err = addSecretCmd.Run()
	if err != nil {
		t.Fatalf("error adding secret: %v", err)
	}

	fmt.Println("Secret added successfully!")

	// Step 4: Retrieve the secret
	password, err := service.RetrieveSecret("MySecret")
	if err != nil {
		t.Fatalf("failed to retrieve secret: %v", err)
	}

	fmt.Println("Retrieved password:", password)
}*/
