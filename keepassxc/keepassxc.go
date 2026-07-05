package keepassxc

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ExecCommand is a package-level variable for mocking exec.Command in tests.
var ExecCommand = exec.Command

// Config holds the configuration for connecting to a KeePassXC database.
type Config struct {
	Database string // Path to .kdbx database file (required)
	KeyFile  string // Path to key file (optional)
	Password string // Database password (optional; if empty, keepassxc-cli reads from stdin)
}

// Client provides methods to interact with a KeePassXC database.
type Client struct {
	config Config
}

// New creates a new KeePassXC client with the given configuration.
func New(cfg Config) *Client {
	return &Client{config: cfg}
}

// GetSecret retrieves the password attribute of the given entry from the KeePassXC database.
// The key is the entry path/title (e.g., "Root/myapp/db-password").
func (c *Client) GetSecret(key string) (string, error) {
	args := []string{"show", "-s", "-a", "Password"}

	if c.config.KeyFile != "" {
		args = append(args, "--key-file", c.config.KeyFile)
	}

	args = append(args, c.config.Database, key)

	cmd := ExecCommand("keepassxc-cli", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if c.config.Password != "" {
		cmd.Stdin = strings.NewReader(c.config.Password + "\n")
	}

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("keepassxc-cli: %s", errMsg)
	}

	return strings.TrimRight(stdout.String(), "\n"), nil
}
