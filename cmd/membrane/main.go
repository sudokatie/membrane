// Membrane - OCI Container Runtime
//
// A minimal container runtime implementing the OCI runtime specification.
// Uses Linux namespaces, cgroups v2, and seccomp for process isolation.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/sudokatie/membrane/internal/container"
)

const version = "0.1.0"

var (
	stateRoot string
	force     bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getManager() *container.Manager {
	config := container.DefaultConfig()
	if stateRoot != "" {
		config.StateRoot = stateRoot
	}
	return container.NewManager(config)
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
		info := map[string]string{
			"version":    version,
			"ociVersion": "1.0.2",
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(info)
	},
}

var createCmd = &cobra.Command{
	Use:   "create <container-id> <bundle>",
	Short: "Create a container",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		_, err := mgr.Create(&container.CreateOptions{
			ID:     args[0],
			Bundle: args[1],
		})
		return err
	},
}

var startCmd = &cobra.Command{
	Use:   "start <container-id>",
	Short: "Start a created container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		return mgr.Start(&container.StartOptions{
			ID: args[0],
		})
	},
}

var runCmd = &cobra.Command{
	Use:   "run <container-id> <bundle>",
	Short: "Create and start a container",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		return mgr.Run(&container.CreateOptions{
			ID:     args[0],
			Bundle: args[1],
		})
	},
}

var stateCmd = &cobra.Command{
	Use:   "state <container-id>",
	Short: "Query container state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		st, err := mgr.State(args[0])
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(st)
	},
}

var killCmd = &cobra.Command{
	Use:   "kill <container-id> [signal]",
	Short: "Send signal to container",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		signal := "SIGTERM"
		if len(args) > 1 {
			signal = args[1]
		}
		mgr := getManager()
		return mgr.Kill(&container.KillOptions{
			ID:     args[0],
			Signal: signal,
		})
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <container-id>",
	Short: "Delete a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		return mgr.Delete(args[0], force)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List containers",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getManager()
		containers, err := mgr.List()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tSTATUS\tPID\tBUNDLE")
		for _, c := range containers {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
				c.ID, c.State.Status, c.State.Pid, c.Bundle)
		}
		return w.Flush()
	},
}

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Generate a default OCI spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Output default spec
		spec := map[string]interface{}{
			"ociVersion": "1.0.2",
			"root": map[string]interface{}{
				"path":     "rootfs",
				"readonly": false,
			},
			"process": map[string]interface{}{
				"terminal": false,
				"user":     map[string]int{"uid": 0, "gid": 0},
				"args":     []string{"/bin/sh"},
				"env":      []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "TERM=xterm"},
				"cwd":      "/",
			},
			"hostname": "container",
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(spec)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&stateRoot, "root", "", "state directory (default: /run/membrane)")

	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "force delete running container")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stateCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(specCmd)
}
