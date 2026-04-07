package network

import (
	"net"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.Mode != ModeBridge {
		t.Errorf("expected ModeBridge, got %s", config.Mode)
	}
	
	if config.BridgeName != "membrane0" {
		t.Errorf("expected membrane0, got %s", config.BridgeName)
	}
	
	if config.Gateway == nil {
		t.Error("expected gateway to be set")
	}
	
	if config.Subnet == nil {
		t.Error("expected subnet to be set")
	}
	
	if len(config.DNS) == 0 {
		t.Error("expected DNS servers to be set")
	}
}

func TestNew(t *testing.T) {
	config := DefaultConfig()
	config.ContainerIP = net.ParseIP("172.17.0.2")
	
	n := New("test-container-123456", config)
	
	if n.containerID != "test-container-123456" {
		t.Errorf("expected container ID to be set")
	}
	
	if n.vethHost != "vethtest-con" {
		t.Errorf("expected veth name vethtest-con, got %s", n.vethHost)
	}
	
	if n.vethContainer != "eth0" {
		t.Errorf("expected container interface eth0, got %s", n.vethContainer)
	}
}

func TestNewWithNilConfig(t *testing.T) {
	n := New("test", nil)
	
	if n.config == nil {
		t.Error("expected default config to be created")
	}
	
	if n.config.Mode != ModeBridge {
		t.Errorf("expected default mode to be bridge")
	}
}

func TestPortMapping(t *testing.T) {
	pm := PortMapping{
		HostPort:      8080,
		ContainerPort: 80,
		Protocol:      "tcp",
	}
	
	if pm.HostPort != 8080 {
		t.Errorf("expected host port 8080")
	}
	
	if pm.ContainerPort != 80 {
		t.Errorf("expected container port 80")
	}
}

func TestNetworkMode(t *testing.T) {
	tests := []struct {
		mode NetworkMode
		want string
	}{
		{ModeNone, "none"},
		{ModeHost, "host"},
		{ModeBridge, "bridge"},
	}
	
	for _, tt := range tests {
		if string(tt.mode) != tt.want {
			t.Errorf("mode %v: got %s, want %s", tt.mode, string(tt.mode), tt.want)
		}
	}
}

func TestMaskSize(t *testing.T) {
	tests := []struct {
		cidr string
		want int
	}{
		{"172.17.0.0/16", 16},
		{"10.0.0.0/8", 8},
		{"192.168.1.0/24", 24},
	}
	
	for _, tt := range tests {
		_, subnet, err := net.ParseCIDR(tt.cidr)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", tt.cidr, err)
		}
		
		got := maskSize(subnet.Mask)
		if got != tt.want {
			t.Errorf("maskSize(%s): got %d, want %d", tt.cidr, got, tt.want)
		}
	}
}

func TestGetContainerIP(t *testing.T) {
	config := DefaultConfig()
	config.ContainerIP = net.ParseIP("172.17.0.5")
	
	n := New("test", config)
	
	ip := n.GetContainerIP()
	if !ip.Equal(config.ContainerIP) {
		t.Errorf("expected %s, got %s", config.ContainerIP, ip)
	}
}

func TestGetGateway(t *testing.T) {
	config := DefaultConfig()
	n := New("test", config)
	
	gw := n.GetGateway()
	if !gw.Equal(config.Gateway) {
		t.Errorf("expected %s, got %s", config.Gateway, gw)
	}
}

func TestTeardownNotSetup(t *testing.T) {
	n := New("test", nil)
	
	// Should not error if not setup
	if err := n.Teardown(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNetworkConfigPortMappings(t *testing.T) {
	config := &NetworkConfig{
		Mode: ModeBridge,
		PortMappings: []PortMapping{
			{HostPort: 80, ContainerPort: 80, Protocol: "tcp"},
			{HostPort: 443, ContainerPort: 443, Protocol: "tcp"},
			{HostPort: 53, ContainerPort: 53, Protocol: "udp"},
		},
	}
	
	if len(config.PortMappings) != 3 {
		t.Errorf("expected 3 port mappings, got %d", len(config.PortMappings))
	}
}

func TestNetworkConfigDNS(t *testing.T) {
	config := &NetworkConfig{
		DNS: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
	}
	
	if len(config.DNS) != 3 {
		t.Errorf("expected 3 DNS servers, got %d", len(config.DNS))
	}
}

func TestNetworkModeNone(t *testing.T) {
	config := &NetworkConfig{
		Mode: ModeNone,
	}
	
	n := New("test", config)
	
	// Setup with mode none should succeed without doing anything
	err := n.Setup(1)
	if err != nil {
		t.Errorf("unexpected error for mode none: %v", err)
	}
}

func TestNetworkModeHost(t *testing.T) {
	config := &NetworkConfig{
		Mode: ModeHost,
	}
	
	n := New("test", config)
	
	// Setup with mode host should succeed without doing anything
	err := n.Setup(1)
	if err != nil {
		t.Errorf("unexpected error for mode host: %v", err)
	}
}

func TestShortContainerID(t *testing.T) {
	// Short ID should be used as-is
	n := New("abc", nil)
	if n.vethHost != "vethabc" {
		t.Errorf("expected vethabc, got %s", n.vethHost)
	}
}
