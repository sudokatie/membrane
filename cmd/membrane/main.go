// Membrane - OCI Container Runtime
//
// A minimal container runtime implementing the OCI runtime specification.
// Uses Linux namespaces, cgroups v2, and seccomp for process isolation.

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "membrane",
	Short: "OCI container runtime",
	Long:  "Membrane is a minimal OCI-compliant container runtime using Linux namespaces and cgroups v2.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("membrane version %s\n", version)
	},
}

var createCmd = &cobra.Command{
	Use:   "create <container-id> <bundle>",
	Short: "Create a container",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		bundle := args[1]
		fmt.Printf("Creating container %s from %s\n", id, bundle)
		// TODO: implement
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start <container-id>",
	Short: "Start a created container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		fmt.Printf("Starting container %s\n", id)
		// TODO: implement
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run <container-id> <bundle>",
	Short: "Create and start a container",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		bundle := args[1]
		fmt.Printf("Running container %s from %s\n", id, bundle)
		// TODO: implement
		return nil
	},
}

var stateCmd = &cobra.Command{
	Use:   "state <container-id>",
	Short: "Query container state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		// TODO: get actual state
		state := map[string]interface{}{
			"ociVersion": "1.0.2",
			"id":         id,
			"status":     "created",
			"pid":        0,
			"bundle":     "",
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(state)
	},
}

var killCmd = &cobra.Command{
	Use:   "kill <container-id> [signal]",
	Short: "Send signal to container",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		signal := "SIGTERM"
		if len(args) > 1 {
			signal = args[1]
		}
		fmt.Printf("Sending %s to container %s\n", signal, id)
		// TODO: implement
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <container-id>",
	Short: "Delete a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		fmt.Printf("Deleting container %s\n", id)
		// TODO: implement
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("ID\tSTATUS\tPID\tBUNDLE")
		// TODO: implement
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stateCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
}
