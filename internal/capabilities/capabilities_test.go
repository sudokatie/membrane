package capabilities

import (
	"testing"
)

func TestParseCapability(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    uint
		wantErr bool
	}{
		{"CAP_CHOWN", "CAP_CHOWN", 0, false},
		{"chown lowercase", "chown", 0, false},
		{"CAP_NET_ADMIN", "CAP_NET_ADMIN", 12, false},
		{"unknown", "CAP_UNKNOWN", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCapability(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCapability() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseCapability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToBitset(t *testing.T) {
	caps := []string{"CAP_CHOWN", "CAP_DAC_OVERRIDE", "CAP_FOWNER"}
	bits, err := ToBitset(caps)
	if err != nil {
		t.Fatalf("ToBitset() error = %v", err)
	}

	// CAP_CHOWN=0, CAP_DAC_OVERRIDE=1, CAP_FOWNER=3
	expected := uint64(1<<0 | 1<<1 | 1<<3)
	if bits != expected {
		t.Errorf("ToBitset() = %x, want %x", bits, expected)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	if len(config.Bounding) == 0 {
		t.Error("DefaultConfig() has empty bounding set")
	}
}

func TestFromSpec(t *testing.T) {
	// Test nil input
	if FromSpec(nil) != nil {
		t.Error("FromSpec(nil) should return nil")
	}
}
