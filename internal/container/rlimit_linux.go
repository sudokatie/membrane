//go:build linux

package container

import (
	"fmt"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/sudokatie/membrane/pkg/oci"
)

// rlimitMap maps OCI rlimit type strings to unix constants.
var rlimitMap = map[string]int{
	"RLIMIT_AS":         unix.RLIMIT_AS,
	"RLIMIT_CORE":       unix.RLIMIT_CORE,
	"RLIMIT_CPU":        unix.RLIMIT_CPU,
	"RLIMIT_DATA":       unix.RLIMIT_DATA,
	"RLIMIT_FSIZE":      unix.RLIMIT_FSIZE,
	"RLIMIT_LOCKS":      unix.RLIMIT_LOCKS,
	"RLIMIT_MEMLOCK":    unix.RLIMIT_MEMLOCK,
	"RLIMIT_MSGQUEUE":   unix.RLIMIT_MSGQUEUE,
	"RLIMIT_NICE":       unix.RLIMIT_NICE,
	"RLIMIT_NOFILE":     unix.RLIMIT_NOFILE,
	"RLIMIT_NPROC":      unix.RLIMIT_NPROC,
	"RLIMIT_RSS":        unix.RLIMIT_RSS,
	"RLIMIT_RTPRIO":     unix.RLIMIT_RTPRIO,
	"RLIMIT_RTTIME":     unix.RLIMIT_RTTIME,
	"RLIMIT_SIGPENDING": unix.RLIMIT_SIGPENDING,
	"RLIMIT_STACK":      unix.RLIMIT_STACK,
}

// applyRlimits applies resource limits to the current process.
func applyRlimits(rlimits []oci.POSIXRlimit) error {
	for _, rl := range rlimits {
		// Normalize type name
		typeName := strings.ToUpper(rl.Type)
		if !strings.HasPrefix(typeName, "RLIMIT_") {
			typeName = "RLIMIT_" + typeName
		}

		resource, ok := rlimitMap[typeName]
		if !ok {
			return fmt.Errorf("unknown rlimit type: %s", rl.Type)
		}

		limit := unix.Rlimit{
			Cur: rl.Soft,
			Max: rl.Hard,
		}

		if err := unix.Setrlimit(resource, &limit); err != nil {
			return fmt.Errorf("setrlimit %s: %w", rl.Type, err)
		}
	}
	return nil
}
