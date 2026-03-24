// Package filesystem handles mount operations and rootfs setup.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// OverlayConfig holds configuration for an overlay mount.
type OverlayConfig struct {
	// LowerDirs are the read-only lower directories (bottom to top).
	LowerDirs []string
	// UpperDir is the writable upper directory.
	UpperDir string
	// WorkDir is the overlay work directory.
	WorkDir string
	// MergedDir is the merged view mount point.
	MergedDir string
}

// ValidateOverlayConfig validates an overlay configuration.
func ValidateOverlayConfig(config *OverlayConfig) error {
	if config == nil {
		return fmt.Errorf("overlay config is nil")
	}

	if len(config.LowerDirs) == 0 {
		return fmt.Errorf("at least one lower directory is required")
	}

	for _, lower := range config.LowerDirs {
		if _, err := os.Stat(lower); err != nil {
			return fmt.Errorf("lower directory %s: %w", lower, err)
		}
	}

	if config.UpperDir == "" {
		return fmt.Errorf("upper directory is required")
	}

	if config.WorkDir == "" {
		return fmt.Errorf("work directory is required")
	}

	if config.MergedDir == "" {
		return fmt.Errorf("merged directory is required")
	}

	return nil
}

// PrepareOverlayDirs creates the necessary directories for an overlay mount.
func PrepareOverlayDirs(config *OverlayConfig) error {
	// Create upper directory
	if err := os.MkdirAll(config.UpperDir, 0755); err != nil {
		return fmt.Errorf("create upper dir: %w", err)
	}

	// Create work directory
	if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}

	// Create merged directory
	if err := os.MkdirAll(config.MergedDir, 0755); err != nil {
		return fmt.Errorf("create merged dir: %w", err)
	}

	return nil
}

// BuildOverlayOptions builds the mount options string for overlayfs.
func BuildOverlayOptions(config *OverlayConfig) string {
	// Join lower directories with colons
	lowerDirs := strings.Join(config.LowerDirs, ":")

	// Build options string
	return fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		lowerDirs, config.UpperDir, config.WorkDir)
}

// OverlayFromLayers creates an overlay config from container layers.
// layerPaths should be ordered from base (index 0) to top.
func OverlayFromLayers(layerPaths []string, containerDir string) *OverlayConfig {
	return &OverlayConfig{
		LowerDirs: layerPaths,
		UpperDir:  filepath.Join(containerDir, "diff"),
		WorkDir:   filepath.Join(containerDir, "work"),
		MergedDir: filepath.Join(containerDir, "merged"),
	}
}
