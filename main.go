package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AmadlaOrg/doorman-keepassxc/keepassxc"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	version        = "1.0.0"
	infoOutputFlag string
	infoHeryFlag   bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "doorman-keepassxc",
		Short: "KeePassXC password database plugin for Doorman",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	infoCmd := newInfoCmd()
	infoCmd.Flags().StringVarP(&infoOutputFlag, "output", "o", "table", "Output format: table, json, yaml")
	infoCmd.Flags().BoolVar(&infoHeryFlag, "hery", false, "Wrap output in HERY envelope (_type, _body)")

	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(newGetCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print plugin metadata as JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := map[string]any{
				"name":        "doorman-keepassxc",
				"version":     version,
				"engine":      "keepassxc",
				"supports":    []string{"amadla.org/entity/secret@^v1.0.0"},
				"description": "KeePassXC password database plugin for Doorman",
			}
			return writeInfoOutput(os.Stdout, infoOutputFlag, infoHeryFlag, info)
		},
	}
}

type heryEnvelope struct {
	Type string `json:"_type" yaml:"_type"`
	Body any    `json:"_body" yaml:"_body"`
}

func writeInfoOutput(w io.Writer, format string, hery bool, data map[string]any) error {
	if hery {
		return writeHeryOutput(w, format, data)
	}

	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "yaml":
		bytes, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, string(bytes))
		return err
	default:
		table := tablewriter.NewWriter(w)
		table.Header("Field", "Value")
		table.Append("Name", fmt.Sprint(data["name"]))
		table.Append("Version", fmt.Sprint(data["version"]))
		if eng, ok := data["engine"]; ok {
			table.Append("Engine", fmt.Sprint(eng))
		}
		table.Append("Description", fmt.Sprint(data["description"]))
		if supports, ok := data["supports"].([]string); ok {
			table.Append("Supports", strings.Join(supports, "\n"))
		}
		if exts, ok := data["file_extensions"].([]string); ok {
			table.Append("File Extensions", strings.Join(exts, ", "))
		}
		table.Render()
		return nil
	}
}

func writeHeryOutput(w io.Writer, format string, data map[string]any) error {
	envelope := heryEnvelope{
		Type: "amadla.org/entity/tools/info@v1.0.0",
		Body: data,
	}

	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(envelope)
	case "table":
		fmt.Fprintf(w, "_type: %s\n\n", envelope.Type)
		table := tablewriter.NewWriter(w)
		table.Header("Field", "Value")
		table.Append("Name", fmt.Sprint(data["name"]))
		table.Append("Version", fmt.Sprint(data["version"]))
		if eng, ok := data["engine"]; ok {
			table.Append("Engine", fmt.Sprint(eng))
		}
		table.Append("Description", fmt.Sprint(data["description"]))
		if supports, ok := data["supports"].([]string); ok {
			table.Append("Supports", strings.Join(supports, "\n"))
		}
		if exts, ok := data["file_extensions"].([]string); ok {
			table.Append("File Extensions", strings.Join(exts, ", "))
		}
		table.Render()
		return nil
	default:
		bytes, err := yaml.Marshal(envelope)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, string(bytes))
		return err
	}
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Retrieve a secret from KeePassXC",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := os.Getenv("KEEPASSXC_DB")
			if dbPath == "" {
				fmt.Fprintln(os.Stderr, "error: KEEPASSXC_DB environment variable is required")
				os.Exit(2)
			}

			client := keepassxc.New(keepassxc.Config{
				Database: dbPath,
				KeyFile:  os.Getenv("KEEPASSXC_KEYFILE"),
				Password: os.Getenv("KEEPASSXC_PASSWORD"),
			})

			secret, err := client.GetSecret(args[0])
			if err != nil {
				return fmt.Errorf("failed to retrieve secret: %w", err)
			}

			fmt.Print(secret)
			return nil
		},
	}
}
