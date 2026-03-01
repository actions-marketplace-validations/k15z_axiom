package config

import (
	"fmt"
	"os"
	"strings"
)

type CacheConfig struct {
	Enabled bool
	Dir     string
}

type Config struct {
	Model   string
	TestDir string
	Cache   CacheConfig
	APIKey  string
}

func Default() Config {
	return Config{
		Model:   "claude-haiku-4-5-20251001",
		TestDir: ".axiom/",
		Cache: CacheConfig{
			Enabled: true,
			Dir:     ".axiom/.cache/",
		},
	}
}

func Load(testDir string) (Config, error) {
	// Load .env before anything else so vars are available via os.Getenv
	loadDotEnv()

	cfg := Default()

	if testDir != "" {
		cfg.TestDir = testDir
	}

	cfg.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	if cfg.APIKey == "" {
		return cfg, fmt.Errorf("ANTHROPIC_API_KEY is not set (set it in the environment or a .env file)")
	}

	return cfg, nil
}

// loadDotEnv reads a .env file from the current directory and sets any
// environment variables that are not already set. Silently ignores missing files.
func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
