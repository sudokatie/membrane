//go:build linux

package cgroup

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// cgroupRoot is the default cgroup v2 mount point.
	cgroupRoot = "/sys/fs/cgroup"
)

// V2Manager implements Manager for cgroups v2.
type V2Manager struct {
	config *Config
	path   string
}

// NewV2Manager creates a new cgroups v2 manager.
func NewV2Manager(config *Config) *V2Manager {
	path := filepath.Join(cgroupRoot, config.Parent, config.Name)
	return &V2Manager{
		config: config,
		path:   path,
	}
}

// Create creates the cgroup directory.
func (m *V2Manager) Create() error {
	// Ensure parent exists
	parent := filepath.Dir(m.path)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("create parent cgroup: %w", err)
	}

	// Enable controllers in parent
	if err := m.enableControllers(parent); err != nil {
		// Non-fatal: controllers may already be enabled
	}

	// Create cgroup directory
	if err := os.MkdirAll(m.path, 0755); err != nil {
		return fmt.Errorf("create cgroup: %w", err)
	}

	return nil
}

// enableControllers enables controllers in the parent cgroup.
func (m *V2Manager) enableControllers(parent string) error {
	// Read available controllers
	controllersPath := filepath.Join(parent, "cgroup.controllers")
	data, err := os.ReadFile(controllersPath)
	if err != nil {
		return err
	}

	controllers := strings.Fields(string(data))
	if len(controllers) == 0 {
		return nil
	}

	// Build subtree_control string
	var enable []string
	for _, c := range controllers {
		enable = append(enable, "+"+c)
	}

	// Write to subtree_control
	subtreePath := filepath.Join(parent, "cgroup.subtree_control")
	return os.WriteFile(subtreePath, []byte(strings.Join(enable, " ")), 0644)
}

// AddProcess adds a process to the cgroup.
func (m *V2Manager) AddProcess(pid int) error {
	procsPath := filepath.Join(m.path, "cgroup.procs")
	return os.WriteFile(procsPath, []byte(strconv.Itoa(pid)), 0644)
}

// SetResources sets resource limits.
func (m *V2Manager) SetResources(resources *Resources) error {
	if resources == nil {
		return nil
	}

	// Memory limits
	if resources.MemoryLimit > 0 {
		if err := m.writeFile("memory.max", strconv.FormatInt(resources.MemoryLimit, 10)); err != nil {
			return fmt.Errorf("set memory.max: %w", err)
		}
	} else if resources.MemoryLimit == -1 {
		if err := m.writeFile("memory.max", "max"); err != nil {
			// Ignore error - file may not exist
		}
	}

	// memory.high (soft limit - triggers reclaim pressure)
	if resources.MemoryHigh > 0 {
		if err := m.writeFile("memory.high", strconv.FormatInt(resources.MemoryHigh, 10)); err != nil {
			// Ignore error - memory controller may not be enabled
		}
	} else if resources.MemoryHigh == -1 {
		if err := m.writeFile("memory.high", "max"); err != nil {
			// Ignore error
		}
	}

	if resources.MemorySwapLimit > 0 {
		if err := m.writeFile("memory.swap.max", strconv.FormatInt(resources.MemorySwapLimit, 10)); err != nil {
			// Ignore error - swap controller may not be enabled
		}
	}

	if resources.MemoryReservation > 0 {
		if err := m.writeFile("memory.low", strconv.FormatInt(resources.MemoryReservation, 10)); err != nil {
			// Ignore error
		}
	}

	// CPU limits
	if resources.CPUQuota > 0 || resources.CPUPeriod > 0 {
		period := resources.CPUPeriod
		if period == 0 {
			period = 100000 // default 100ms
		}
		quota := resources.CPUQuota
		if quota <= 0 {
			quota = -1 // max
		}

		var cpuMax string
		if quota == -1 {
			cpuMax = fmt.Sprintf("max %d", period)
		} else {
			cpuMax = fmt.Sprintf("%d %d", quota, period)
		}

		if err := m.writeFile("cpu.max", cpuMax); err != nil {
			return fmt.Errorf("set cpu.max: %w", err)
		}
	}

	if resources.CPUShares > 0 {
		// Convert shares to weight (1-10000)
		// shares 1024 = weight 100 (default)
		weight := (resources.CPUShares * 100) / 1024
		if weight < 1 {
			weight = 1
		}
		if weight > 10000 {
			weight = 10000
		}
		if err := m.writeFile("cpu.weight", strconv.FormatUint(weight, 10)); err != nil {
			// Ignore error
		}
	}

	// PIDs limit
	if resources.PidsLimit > 0 {
		if err := m.writeFile("pids.max", strconv.FormatInt(resources.PidsLimit, 10)); err != nil {
			return fmt.Errorf("set pids.max: %w", err)
		}
	} else if resources.PidsLimit == -1 {
		if err := m.writeFile("pids.max", "max"); err != nil {
			// Ignore error
		}
	}

	// IO weight
	if resources.IOWeight > 0 {
		if err := m.writeFile("io.weight", strconv.FormatUint(uint64(resources.IOWeight), 10)); err != nil {
			// Ignore error - io controller may not be enabled
		}
	}

	// IO throttle limits (io.max)
	// Format: "MAJ:MIN rbps=BYTES wbps=BYTES riops=IOPS wiops=IOPS"
	if err := m.setIOMax(resources); err != nil {
		// Ignore error - io controller may not be enabled
	}

	return nil
}

// setIOMax sets IO throttle limits via io.max.
func (m *V2Manager) setIOMax(resources *Resources) error {
	// Build io.max entries per device
	deviceLimits := make(map[string][]string)

	// Add read BPS limits
	for _, d := range resources.IOReadBPS {
		key := fmt.Sprintf("%d:%d", d.Major, d.Minor)
		deviceLimits[key] = append(deviceLimits[key], fmt.Sprintf("rbps=%d", d.Rate))
	}

	// Add write BPS limits
	for _, d := range resources.IOWriteBPS {
		key := fmt.Sprintf("%d:%d", d.Major, d.Minor)
		deviceLimits[key] = append(deviceLimits[key], fmt.Sprintf("wbps=%d", d.Rate))
	}

	// Add read IOPS limits
	for _, d := range resources.IOReadIOPS {
		key := fmt.Sprintf("%d:%d", d.Major, d.Minor)
		deviceLimits[key] = append(deviceLimits[key], fmt.Sprintf("riops=%d", d.Rate))
	}

	// Add write IOPS limits
	for _, d := range resources.IOWriteIOPS {
		key := fmt.Sprintf("%d:%d", d.Major, d.Minor)
		deviceLimits[key] = append(deviceLimits[key], fmt.Sprintf("wiops=%d", d.Rate))
	}

	// Write each device's limits
	for device, limits := range deviceLimits {
		line := device + " " + strings.Join(limits, " ")
		if err := m.writeFile("io.max", line); err != nil {
			return fmt.Errorf("set io.max for %s: %w", device, err)
		}
	}

	return nil
}

// GetResources gets current resource limits.
func (m *V2Manager) GetResources() (*Resources, error) {
	r := &Resources{
		MemoryLimit: -1,
		MemoryHigh:  -1,
		CPUQuota:    -1,
		PidsLimit:   -1,
	}

	// Memory max
	if data, err := m.readFile("memory.max"); err == nil {
		if data != "max" {
			if val, err := strconv.ParseInt(data, 10, 64); err == nil {
				r.MemoryLimit = val
			}
		}
	}

	// Memory high
	if data, err := m.readFile("memory.high"); err == nil {
		if data != "max" {
			if val, err := strconv.ParseInt(data, 10, 64); err == nil {
				r.MemoryHigh = val
			}
		}
	}

	// CPU
	if data, err := m.readFile("cpu.max"); err == nil {
		parts := strings.Fields(data)
		if len(parts) >= 2 {
			if parts[0] != "max" {
				if val, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					r.CPUQuota = val
				}
			}
			if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
				r.CPUPeriod = val
			}
		}
	}

	// PIDs
	if data, err := m.readFile("pids.max"); err == nil {
		if data != "max" {
			if val, err := strconv.ParseInt(data, 10, 64); err == nil {
				r.PidsLimit = val
			}
		}
	}

	return r, nil
}

// Delete removes the cgroup.
func (m *V2Manager) Delete() error {
	// First, kill all processes in the cgroup
	if err := m.killAll(); err != nil {
		// Continue anyway
	}

	// Remove the directory
	if err := os.Remove(m.path); err != nil {
		return fmt.Errorf("remove cgroup: %w", err)
	}
	return nil
}

// killAll kills all processes in the cgroup.
func (m *V2Manager) killAll() error {
	procsPath := filepath.Join(m.path, "cgroup.procs")
	f, err := os.Open(procsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pid, err := strconv.Atoi(scanner.Text())
		if err != nil {
			continue
		}
		// Send SIGKILL
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		proc.Kill()
	}
	return scanner.Err()
}

// Path returns the cgroup path.
func (m *V2Manager) Path() string {
	return m.path
}

// writeFile writes a value to a cgroup file.
func (m *V2Manager) writeFile(name, value string) error {
	path := filepath.Join(m.path, name)
	return os.WriteFile(path, []byte(value), 0644)
}

// readFile reads a value from a cgroup file.
func (m *V2Manager) readFile(name string) (string, error) {
	path := filepath.Join(m.path, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// Exists returns true if the cgroup exists.
func (m *V2Manager) Exists() bool {
	_, err := os.Stat(m.path)
	return err == nil
}

// GetPids returns all process IDs in the cgroup.
func (m *V2Manager) GetPids() ([]int, error) {
	procsPath := filepath.Join(m.path, "cgroup.procs")
	f, err := os.Open(procsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pids []int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pid, err := strconv.Atoi(scanner.Text())
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, scanner.Err()
}

// IsCgroupV2 returns true if cgroups v2 is available.
func IsCgroupV2() bool {
	// Check for cgroup.controllers file at root
	_, err := os.Stat(filepath.Join(cgroupRoot, "cgroup.controllers"))
	return err == nil
}
