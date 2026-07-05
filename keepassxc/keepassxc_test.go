package keepassxc

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelperProcess is the subprocess helper used by fakeExecCommand.
// It is invoked as a test binary by the mocked exec.Command.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}

	// Behaviour is controlled via GO_TEST_HELPER_* env vars.
	switch os.Getenv("GO_TEST_HELPER_MODE") {
	case "success":
		fmt.Fprint(os.Stdout, os.Getenv("GO_TEST_HELPER_STDOUT"))
	case "fail":
		fmt.Fprint(os.Stderr, os.Getenv("GO_TEST_HELPER_STDERR"))
		os.Exit(1)
	}

	os.Exit(0)
}

// fakeExecCommand returns an exec.Cmd that re-invokes the test binary
// as a subprocess with the helper env vars set.
func fakeExecCommand(mode, stdout, stderr string) func(string, ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		csArgs := []string{"-test.run=TestHelperProcess", "--"}
		csArgs = append(csArgs, name)
		csArgs = append(csArgs, args...)

		cmd := exec.Command(os.Args[0], csArgs...)
		cmd.Env = append(os.Environ(),
			"GO_TEST_HELPER_PROCESS=1",
			"GO_TEST_HELPER_MODE="+mode,
			"GO_TEST_HELPER_STDOUT="+stdout,
			"GO_TEST_HELPER_STDERR="+stderr,
		)
		return cmd
	}
}

func TestGetSecret_Success(t *testing.T) {
	original := ExecCommand
	defer func() { ExecCommand = original }()

	ExecCommand = fakeExecCommand("success", "s3cret", "")

	client := New(Config{
		Database: "/tmp/test.kdbx",
		Password: "masterpass",
	})

	secret, err := client.GetSecret("Root/myapp/db-password")
	require.NoError(t, err)
	assert.Equal(t, "s3cret", secret)
}

func TestGetSecret_WithKeyFile(t *testing.T) {
	original := ExecCommand
	defer func() { ExecCommand = original }()

	ExecCommand = fakeExecCommand("success", "keyfile-secret", "")

	client := New(Config{
		Database: "/tmp/test.kdbx",
		KeyFile:  "/tmp/key.keyx",
		Password: "masterpass",
	})

	secret, err := client.GetSecret("Root/service/api-key")
	require.NoError(t, err)
	assert.Equal(t, "keyfile-secret", secret)
}

func TestGetSecret_NoPassword(t *testing.T) {
	original := ExecCommand
	defer func() { ExecCommand = original }()

	ExecCommand = fakeExecCommand("success", "interactive-secret", "")

	client := New(Config{
		Database: "/tmp/test.kdbx",
	})

	secret, err := client.GetSecret("Root/entry")
	require.NoError(t, err)
	assert.Equal(t, "interactive-secret", secret)
}

func TestGetSecret_Failure(t *testing.T) {
	original := ExecCommand
	defer func() { ExecCommand = original }()

	ExecCommand = fakeExecCommand("fail", "", "Could not find entry with path Root/missing.")

	client := New(Config{
		Database: "/tmp/test.kdbx",
		Password: "masterpass",
	})

	secret, err := client.GetSecret("Root/missing")
	assert.Error(t, err)
	assert.Empty(t, secret)
	assert.Contains(t, err.Error(), "Could not find entry with path Root/missing.")
}

func TestGetSecret_TrimsTrailingNewline(t *testing.T) {
	original := ExecCommand
	defer func() { ExecCommand = original }()

	ExecCommand = fakeExecCommand("success", "secret-value\n", "")

	client := New(Config{
		Database: "/tmp/test.kdbx",
		Password: "pass",
	})

	secret, err := client.GetSecret("Root/entry")
	require.NoError(t, err)
	assert.Equal(t, "secret-value", secret)
}
