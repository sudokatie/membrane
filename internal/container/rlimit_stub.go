//go:build !linux

package container

import (
	"fmt"

	"github.com/sudokatie/membrane/pkg/oci"
)

// applyRlimits is a stub for non-Linux systems.
func applyRlimits(rlimits []oci.POSIXRlimit) error {
	if len(rlimits) > 0 {
		return fmt.Errorf("rlimits not supported on this platform")
	}
	return nil
}
