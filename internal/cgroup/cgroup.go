// Package cgroup manages Linux cgroups v2.
package cgroup

import (
	"github.com/sudokatie/membrane/pkg/oci"
)

// Manager is the interface for cgroup operations.
type Manager interface {
	// Create creates the cgroup.
	Create() error
	// AddProcess adds a process to the cgroup.
	AddProcess(pid int) error
	// SetResources sets resource limits.
	SetResources(resources *Resources) error
	// GetResources gets current resource limits.
	GetResources() (*Resources, error)
	// Delete removes the cgroup.
	Delete() error
	// Path returns the cgroup path.
	Path() string
}

// Resources represents cgroup resource limits.
type Resources struct {
	// Memory limits
	MemoryLimit      int64 // bytes, -1 for unlimited
	MemorySwapLimit  int64 // bytes, -1 for unlimited
	MemoryReservation int64 // bytes

	// CPU limits
	CPUQuota  int64  // microseconds per period, -1 for unlimited
	CPUPeriod uint64 // microseconds (default 100000)
	CPUShares uint64 // relative weight

	// PIDs limit
	PidsLimit int64 // max pids, -1 for unlimited

	// IO limits (simplified)
	IOWeight uint16 // 1-10000
}

// FromSpec creates Resources from an OCI LinuxResources spec.
func FromSpec(spec *oci.LinuxResources) *Resources {
	if spec == nil {
		return &Resources{
			MemoryLimit: -1,
			CPUQuota:    -1,
			PidsLimit:   -1,
		}
	}

	r := &Resources{
		MemoryLimit: -1,
		CPUQuota:    -1,
		PidsLimit:   -1,
	}

	// Memory
	if spec.Memory != nil {
		if spec.Memory.Limit != nil {
			r.MemoryLimit = *spec.Memory.Limit
		}
		if spec.Memory.Swap != nil {
			r.MemorySwapLimit = *spec.Memory.Swap
		}
		if spec.Memory.Reservation != nil {
			r.MemoryReservation = *spec.Memory.Reservation
		}
	}

	// CPU
	if spec.CPU != nil {
		if spec.CPU.Quota != nil {
			r.CPUQuota = *spec.CPU.Quota
		}
		if spec.CPU.Period != nil {
			r.CPUPeriod = *spec.CPU.Period
		}
		if spec.CPU.Shares != nil {
			r.CPUShares = *spec.CPU.Shares
		}
	}

	// PIDs
	if spec.Pids != nil {
		r.PidsLimit = spec.Pids.Limit
	}

	// IO
	if spec.BlockIO != nil && spec.BlockIO.Weight != nil {
		r.IOWeight = *spec.BlockIO.Weight
	}

	return r
}

// Config holds cgroup configuration.
type Config struct {
	// Name is the cgroup name (container ID).
	Name string
	// Parent is the parent cgroup path (default: /membrane).
	Parent string
	// Resources are the resource limits.
	Resources *Resources
}

// DefaultConfig returns a default cgroup config.
func DefaultConfig(name string) *Config {
	return &Config{
		Name:   name,
		Parent: "/membrane",
		Resources: &Resources{
			MemoryLimit: -1,
			CPUQuota:    -1,
			PidsLimit:   -1,
		},
	}
}
