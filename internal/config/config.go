// Package config loads Barbarossa CLI configuration from .barbarossa.yaml
// and environment variables, providing sensible defaults.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the Barbarossa CLI.
type Config struct {
	Docker     DockerConfig     `yaml:"docker"`
	Containers ContainerConfig  `yaml:"containers"`
}

// DockerConfig holds Docker connection settings.
type DockerConfig struct {
	Host       string `yaml:"host"`
	APIVersion string `yaml:"api_version"`
}

// ContainerConfig holds the list of monitored container names.
type ContainerConfig struct {
	Names []string `yaml:"names"`
}

// DefaultContainerNames are the three workers monitored by default.
var DefaultContainerNames = []string{"charlie", "oscar", "papa"}

// DefaultDockerHost is the default Docker daemon socket path.
const DefaultDockerHost = "unix:///var/run/docker.sock"

// Load reads configuration from .barbarossa.yaml (if it exists), overlays
// environment variables, and fills in defaults for any missing values.
func Load() *Config {
	cfg := &Config{
		Docker: DockerConfig{
			Host: DefaultDockerHost,
		},
		Containers: ContainerConfig{
			Names: DefaultContainerNames,
		},
	}

	// Try to read from .barbarossa.yaml
	cfg.loadYAML()

	// Environment variable overrides
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		cfg.Docker.Host = host
	}
	if containers := os.Getenv("BARBAROSSA_CONTAINERS"); containers != "" {
		parts := strings.Split(containers, ",")
		trimmed := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				trimmed = append(trimmed, t)
			}
		}
		if len(trimmed) > 0 {
			cfg.Containers.Names = trimmed
		}
	}

	return cfg
}

// loadYAML attempts to read and parse a .barbarossa.yaml file.
// Search order: BARBAROSSA_CONFIG env var → ./ → ~/
func (cfg *Config) loadYAML() {
	paths := []string{os.Getenv("BARBAROSSA_CONFIG")}

	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".barbarossa.yaml"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".barbarossa.yaml"))
	}

	for _, p := range paths {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		// Merge file values over defaults (file only overrides non-empty values)
		var fileCfg Config
		if err := yaml.Unmarshal(data, &fileCfg); err != nil {
			continue
		}
		if fileCfg.Docker.Host != "" {
			cfg.Docker.Host = fileCfg.Docker.Host
		}
		if fileCfg.Docker.APIVersion != "" {
			cfg.Docker.APIVersion = fileCfg.Docker.APIVersion
		}
		if len(fileCfg.Containers.Names) > 0 {
			cfg.Containers.Names = fileCfg.Containers.Names
		}
		return // First valid file wins
	}
}
