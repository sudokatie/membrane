// Package network provides container networking functionality.
package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NetworkMode specifies the networking mode for a container.
type NetworkMode string

const (
	// ModeNone disables networking.
	ModeNone NetworkMode = "none"
	// ModeHost shares host networking.
	ModeHost NetworkMode = "host"
	// ModeBridge uses bridge networking with NAT.
	ModeBridge NetworkMode = "bridge"
)

// NetworkConfig contains network configuration for a container.
type NetworkConfig struct {
	// Mode specifies the networking mode.
	Mode NetworkMode
	// BridgeName is the name of the bridge to use (for bridge mode).
	BridgeName string
	// ContainerIP is the IP address for the container (for bridge mode).
	ContainerIP net.IP
	// Gateway is the gateway IP (for bridge mode).
	Gateway net.IP
	// Subnet is the network subnet (for bridge mode).
	Subnet *net.IPNet
	// DNS servers to use.
	DNS []string
	// PortMappings maps host ports to container ports.
	PortMappings []PortMapping
	// Hostname for the container.
	Hostname string
}

// PortMapping represents a port forwarding rule.
type PortMapping struct {
	// HostPort is the port on the host.
	HostPort int
	// ContainerPort is the port in the container.
	ContainerPort int
	// Protocol is "tcp" or "udp".
	Protocol string
}

// DefaultConfig returns a default network configuration.
func DefaultConfig() *NetworkConfig {
	_, subnet, _ := net.ParseCIDR("172.17.0.0/16")
	return &NetworkConfig{
		Mode:       ModeBridge,
		BridgeName: "membrane0",
		Gateway:    net.ParseIP("172.17.0.1"),
		Subnet:     subnet,
		DNS:        []string{"8.8.8.8", "8.8.4.4"},
	}
}

// Network manages networking for a container.
type Network struct {
	config      *NetworkConfig
	containerID string
	vethHost    string
	vethContainer string
	setup       bool
}

// New creates a new Network instance.
func New(containerID string, config *NetworkConfig) *Network {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Generate veth pair names
	shortID := containerID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	
	return &Network{
		config:        config,
		containerID:   containerID,
		vethHost:      fmt.Sprintf("veth%s", shortID),
		vethContainer: "eth0",
	}
}

// Setup configures networking for the container.
func (n *Network) Setup(containerPID int) error {
	switch n.config.Mode {
	case ModeNone:
		return nil
	case ModeHost:
		return nil
	case ModeBridge:
		return n.setupBridge(containerPID)
	default:
		return fmt.Errorf("unknown network mode: %s", n.config.Mode)
	}
}

// Teardown removes networking configuration.
func (n *Network) Teardown() error {
	if !n.setup {
		return nil
	}
	
	// Remove veth pair (removing host side removes both)
	if err := runIP("link", "del", n.vethHost); err != nil {
		// Ignore errors - interface may already be gone
	}
	
	// Remove port forwarding rules
	for _, pm := range n.config.PortMappings {
		n.removePortForward(pm)
	}
	
	n.setup = false
	return nil
}

// setupBridge sets up bridge networking.
func (n *Network) setupBridge(containerPID int) error {
	// Ensure bridge exists
	if err := n.ensureBridge(); err != nil {
		return fmt.Errorf("failed to create bridge: %w", err)
	}
	
	// Create veth pair
	if err := n.createVethPair(containerPID); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	}
	
	// Configure container interface
	if err := n.configureContainerInterface(containerPID); err != nil {
		return fmt.Errorf("failed to configure container interface: %w", err)
	}
	
	// Set up port forwarding
	for _, pm := range n.config.PortMappings {
		if err := n.setupPortForward(pm); err != nil {
			return fmt.Errorf("failed to set up port forward: %w", err)
		}
	}
	
	// Set up DNS
	if err := n.setupDNS(containerPID); err != nil {
		return fmt.Errorf("failed to set up DNS: %w", err)
	}
	
	n.setup = true
	return nil
}

// ensureBridge creates the bridge if it doesn't exist.
func (n *Network) ensureBridge() error {
	// Check if bridge exists
	if _, err := net.InterfaceByName(n.config.BridgeName); err == nil {
		return nil
	}
	
	// Create bridge
	if err := runIP("link", "add", n.config.BridgeName, "type", "bridge"); err != nil {
		return err
	}
	
	// Set bridge IP
	addr := fmt.Sprintf("%s/%d", n.config.Gateway.String(), maskSize(n.config.Subnet.Mask))
	if err := runIP("addr", "add", addr, "dev", n.config.BridgeName); err != nil {
		return err
	}
	
	// Bring bridge up
	if err := runIP("link", "set", n.config.BridgeName, "up"); err != nil {
		return err
	}
	
	// Enable IP forwarding
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}
	
	// Set up NAT for outbound traffic
	return runIPTables("-t", "nat", "-A", "POSTROUTING",
		"-s", n.config.Subnet.String(),
		"-j", "MASQUERADE")
}

// createVethPair creates a veth pair and moves one end into the container.
func (n *Network) createVethPair(containerPID int) error {
	// Create veth pair
	if err := runIP("link", "add", n.vethHost, "type", "veth", "peer", "name", n.vethContainer); err != nil {
		return err
	}
	
	// Attach host end to bridge
	if err := runIP("link", "set", n.vethHost, "master", n.config.BridgeName); err != nil {
		return err
	}
	
	// Bring host end up
	if err := runIP("link", "set", n.vethHost, "up"); err != nil {
		return err
	}
	
	// Move container end into container's network namespace
	nsPath := fmt.Sprintf("/proc/%d/ns/net", containerPID)
	return runIP("link", "set", n.vethContainer, "netns", nsPath)
}

// configureContainerInterface configures networking inside the container.
func (n *Network) configureContainerInterface(containerPID int) error {
	nsPath := fmt.Sprintf("/proc/%d/ns/net", containerPID)
	
	// Set container IP
	addr := fmt.Sprintf("%s/%d", n.config.ContainerIP.String(), maskSize(n.config.Subnet.Mask))
	if err := runNsenter(nsPath, "ip", "addr", "add", addr, "dev", n.vethContainer); err != nil {
		return err
	}
	
	// Bring interface up
	if err := runNsenter(nsPath, "ip", "link", "set", n.vethContainer, "up"); err != nil {
		return err
	}
	
	// Bring loopback up
	if err := runNsenter(nsPath, "ip", "link", "set", "lo", "up"); err != nil {
		return err
	}
	
	// Add default route
	return runNsenter(nsPath, "ip", "route", "add", "default", "via", n.config.Gateway.String())
}

// setupPortForward sets up port forwarding with iptables.
func (n *Network) setupPortForward(pm PortMapping) error {
	proto := pm.Protocol
	if proto == "" {
		proto = "tcp"
	}
	
	// DNAT rule for incoming traffic
	return runIPTables("-t", "nat", "-A", "PREROUTING",
		"-p", proto,
		"--dport", fmt.Sprintf("%d", pm.HostPort),
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", n.config.ContainerIP.String(), pm.ContainerPort))
}

// removePortForward removes port forwarding rules.
func (n *Network) removePortForward(pm PortMapping) error {
	proto := pm.Protocol
	if proto == "" {
		proto = "tcp"
	}
	
	return runIPTables("-t", "nat", "-D", "PREROUTING",
		"-p", proto,
		"--dport", fmt.Sprintf("%d", pm.HostPort),
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", n.config.ContainerIP.String(), pm.ContainerPort))
}

// setupDNS configures DNS resolution for the container.
func (n *Network) setupDNS(containerPID int) error {
	// Get container root
	rootPath := fmt.Sprintf("/proc/%d/root", containerPID)
	resolvPath := filepath.Join(rootPath, "etc", "resolv.conf")
	
	// Ensure /etc directory exists
	etcPath := filepath.Join(rootPath, "etc")
	if err := os.MkdirAll(etcPath, 0755); err != nil {
		return err
	}
	
	// Write resolv.conf
	var lines []string
	for _, dns := range n.config.DNS {
		lines = append(lines, fmt.Sprintf("nameserver %s", dns))
	}
	
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(resolvPath, []byte(content), 0644)
}

// GetContainerIP returns the IP address assigned to the container.
func (n *Network) GetContainerIP() net.IP {
	return n.config.ContainerIP
}

// GetGateway returns the gateway IP.
func (n *Network) GetGateway() net.IP {
	return n.config.Gateway
}

// Helper functions

func runIP(args ...string) error {
	cmd := exec.Command("ip", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runIPTables(args ...string) error {
	cmd := exec.Command("iptables", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runNsenter(nsPath string, args ...string) error {
	nsenterArgs := []string{"-n" + nsPath, "--"}
	nsenterArgs = append(nsenterArgs, args...)
	cmd := exec.Command("nsenter", nsenterArgs...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func maskSize(mask net.IPMask) int {
	ones, _ := mask.Size()
	return ones
}
