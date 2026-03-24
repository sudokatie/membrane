package cgroup

import (
	"testing"

	"github.com/sudokatie/membrane/pkg/oci"
)

func TestFromSpec(t *testing.T) {
	memLimit := int64(104857600) // 100MB
	cpuQuota := int64(50000)     // 50ms
	cpuPeriod := uint64(100000)  // 100ms
	pidsLimit := int64(100)

	spec := &oci.LinuxResources{
		Memory: &oci.LinuxMemory{
			Limit: &memLimit,
		},
		CPU: &oci.LinuxCPU{
			Quota:  &cpuQuota,
			Period: &cpuPeriod,
		},
		Pids: &oci.LinuxPids{
			Limit: pidsLimit,
		},
	}

	r := FromSpec(spec)

	if r.MemoryLimit != memLimit {
		t.Errorf("MemoryLimit = %d, want %d", r.MemoryLimit, memLimit)
	}
	if r.CPUQuota != cpuQuota {
		t.Errorf("CPUQuota = %d, want %d", r.CPUQuota, cpuQuota)
	}
	if r.CPUPeriod != cpuPeriod {
		t.Errorf("CPUPeriod = %d, want %d", r.CPUPeriod, cpuPeriod)
	}
	if r.PidsLimit != pidsLimit {
		t.Errorf("PidsLimit = %d, want %d", r.PidsLimit, pidsLimit)
	}
}

func TestFromSpecNil(t *testing.T) {
	r := FromSpec(nil)

	if r.MemoryLimit != -1 {
		t.Errorf("MemoryLimit = %d, want -1", r.MemoryLimit)
	}
	if r.MemoryHigh != -1 {
		t.Errorf("MemoryHigh = %d, want -1", r.MemoryHigh)
	}
	if r.CPUQuota != -1 {
		t.Errorf("CPUQuota = %d, want -1", r.CPUQuota)
	}
	if r.PidsLimit != -1 {
		t.Errorf("PidsLimit = %d, want -1", r.PidsLimit)
	}
}

func TestFromSpecPartial(t *testing.T) {
	memLimit := int64(52428800) // 50MB

	spec := &oci.LinuxResources{
		Memory: &oci.LinuxMemory{
			Limit: &memLimit,
		},
		// No CPU or PIDs
	}

	r := FromSpec(spec)

	if r.MemoryLimit != memLimit {
		t.Errorf("MemoryLimit = %d, want %d", r.MemoryLimit, memLimit)
	}
	if r.CPUQuota != -1 {
		t.Errorf("CPUQuota = %d, want -1 (unset)", r.CPUQuota)
	}
	if r.PidsLimit != -1 {
		t.Errorf("PidsLimit = %d, want -1 (unset)", r.PidsLimit)
	}
}

func TestFromSpecCPUShares(t *testing.T) {
	shares := uint64(512)

	spec := &oci.LinuxResources{
		CPU: &oci.LinuxCPU{
			Shares: &shares,
		},
	}

	r := FromSpec(spec)

	if r.CPUShares != shares {
		t.Errorf("CPUShares = %d, want %d", r.CPUShares, shares)
	}
}

func TestFromSpecIOWeight(t *testing.T) {
	weight := uint16(500)

	spec := &oci.LinuxResources{
		BlockIO: &oci.LinuxBlockIO{
			Weight: &weight,
		},
	}

	r := FromSpec(spec)

	if r.IOWeight != weight {
		t.Errorf("IOWeight = %d, want %d", r.IOWeight, weight)
	}
}

func TestFromSpecIOThrottle(t *testing.T) {
	spec := &oci.LinuxResources{
		BlockIO: &oci.LinuxBlockIO{
			ThrottleReadBpsDevice: []oci.LinuxThrottleDevice{
				{Major: 8, Minor: 0, Rate: 1048576}, // 1 MB/s
			},
			ThrottleWriteBpsDevice: []oci.LinuxThrottleDevice{
				{Major: 8, Minor: 0, Rate: 524288}, // 512 KB/s
			},
			ThrottleReadIOPSDevice: []oci.LinuxThrottleDevice{
				{Major: 8, Minor: 0, Rate: 1000},
			},
			ThrottleWriteIOPSDevice: []oci.LinuxThrottleDevice{
				{Major: 8, Minor: 0, Rate: 500},
			},
		},
	}

	r := FromSpec(spec)

	if len(r.IOReadBPS) != 1 {
		t.Fatalf("IOReadBPS length = %d, want 1", len(r.IOReadBPS))
	}
	if r.IOReadBPS[0].Rate != 1048576 {
		t.Errorf("IOReadBPS[0].Rate = %d, want 1048576", r.IOReadBPS[0].Rate)
	}

	if len(r.IOWriteBPS) != 1 {
		t.Fatalf("IOWriteBPS length = %d, want 1", len(r.IOWriteBPS))
	}
	if r.IOWriteBPS[0].Rate != 524288 {
		t.Errorf("IOWriteBPS[0].Rate = %d, want 524288", r.IOWriteBPS[0].Rate)
	}

	if len(r.IOReadIOPS) != 1 {
		t.Fatalf("IOReadIOPS length = %d, want 1", len(r.IOReadIOPS))
	}
	if r.IOReadIOPS[0].Rate != 1000 {
		t.Errorf("IOReadIOPS[0].Rate = %d, want 1000", r.IOReadIOPS[0].Rate)
	}

	if len(r.IOWriteIOPS) != 1 {
		t.Fatalf("IOWriteIOPS length = %d, want 1", len(r.IOWriteIOPS))
	}
	if r.IOWriteIOPS[0].Rate != 500 {
		t.Errorf("IOWriteIOPS[0].Rate = %d, want 500", r.IOWriteIOPS[0].Rate)
	}
}

func TestFromSpecSwapAndReservation(t *testing.T) {
	memLimit := int64(104857600)
	swapLimit := int64(209715200)
	reservation := int64(52428800)

	spec := &oci.LinuxResources{
		Memory: &oci.LinuxMemory{
			Limit:       &memLimit,
			Swap:        &swapLimit,
			Reservation: &reservation,
		},
	}

	r := FromSpec(spec)

	if r.MemoryLimit != memLimit {
		t.Errorf("MemoryLimit = %d, want %d", r.MemoryLimit, memLimit)
	}
	if r.MemorySwapLimit != swapLimit {
		t.Errorf("MemorySwapLimit = %d, want %d", r.MemorySwapLimit, swapLimit)
	}
	if r.MemoryReservation != reservation {
		t.Errorf("MemoryReservation = %d, want %d", r.MemoryReservation, reservation)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test-container")

	if config.Name != "test-container" {
		t.Errorf("Name = %s, want test-container", config.Name)
	}
	if config.Parent != "/membrane" {
		t.Errorf("Parent = %s, want /membrane", config.Parent)
	}
	if config.Resources == nil {
		t.Fatal("Resources is nil")
	}
	if config.Resources.MemoryLimit != -1 {
		t.Errorf("MemoryLimit = %d, want -1", config.Resources.MemoryLimit)
	}
	if config.Resources.MemoryHigh != -1 {
		t.Errorf("MemoryHigh = %d, want -1", config.Resources.MemoryHigh)
	}
}

func TestNewV2Manager(t *testing.T) {
	config := DefaultConfig("test-container")
	manager := NewV2Manager(config)

	expectedPath := "/sys/fs/cgroup/membrane/test-container"
	if manager.Path() != expectedPath {
		t.Errorf("Path() = %s, want %s", manager.Path(), expectedPath)
	}
}

func TestV2ManagerCustomParent(t *testing.T) {
	config := &Config{
		Name:   "my-container",
		Parent: "/custom/path",
	}
	manager := NewV2Manager(config)

	expectedPath := "/sys/fs/cgroup/custom/path/my-container"
	if manager.Path() != expectedPath {
		t.Errorf("Path() = %s, want %s", manager.Path(), expectedPath)
	}
}

func TestResourcesStruct(t *testing.T) {
	r := &Resources{
		MemoryLimit:       104857600,
		MemoryHigh:        78643200,
		MemorySwapLimit:   209715200,
		MemoryReservation: 52428800,
		CPUQuota:          50000,
		CPUPeriod:         100000,
		CPUShares:         1024,
		PidsLimit:         100,
		IOWeight:          500,
		IOReadBPS: []ThrottleDevice{
			{Major: 8, Minor: 0, Rate: 1048576},
		},
		IOWriteBPS: []ThrottleDevice{
			{Major: 8, Minor: 0, Rate: 524288},
		},
	}

	if r.MemoryLimit != 104857600 {
		t.Errorf("MemoryLimit = %d, want 104857600", r.MemoryLimit)
	}
	if r.MemoryHigh != 78643200 {
		t.Errorf("MemoryHigh = %d, want 78643200", r.MemoryHigh)
	}
	if r.CPUPeriod != 100000 {
		t.Errorf("CPUPeriod = %d, want 100000", r.CPUPeriod)
	}
	if len(r.IOReadBPS) != 1 {
		t.Errorf("IOReadBPS length = %d, want 1", len(r.IOReadBPS))
	}
}

func TestThrottleDevice(t *testing.T) {
	d := ThrottleDevice{
		Major: 8,
		Minor: 16,
		Rate:  1000000,
	}

	if d.Major != 8 {
		t.Errorf("Major = %d, want 8", d.Major)
	}
	if d.Minor != 16 {
		t.Errorf("Minor = %d, want 16", d.Minor)
	}
	if d.Rate != 1000000 {
		t.Errorf("Rate = %d, want 1000000", d.Rate)
	}
}
