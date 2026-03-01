package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Test struct {
	Name       string
	On         []string
	Condition  string
	SourceFile string // relative path to the YAML file
}

type testDefinition struct {
	On        []string `yaml:"on"`
	Condition string   `yaml:"condition"`
}

func Discover(testDir string) ([]Test, error) {
	entries, err := os.ReadDir(testDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("test directory %q not found — run 'axiom init' to create it", testDir)
		}
		return nil, fmt.Errorf("reading test directory %s: %w", testDir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			if name == "config.yml" || name == "config.yaml" {
				continue
			}
			files = append(files, name)
		}
	}
	sort.Strings(files)

	var tests []Test
	for _, file := range files {
		path := filepath.Join(testDir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		// Use yaml.Decoder with KnownFields to preserve order via ordered map
		var raw yaml.Node
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		if raw.Kind != yaml.DocumentNode || len(raw.Content) == 0 {
			continue
		}
		mapping := raw.Content[0]
		if mapping.Kind != yaml.MappingNode {
			continue
		}

		// Iterate key-value pairs to preserve order
		for i := 0; i < len(mapping.Content)-1; i += 2 {
			keyNode := mapping.Content[i]
			valNode := mapping.Content[i+1]

			var def testDefinition
			if err := valNode.Decode(&def); err != nil {
				return nil, fmt.Errorf("parsing test %q in %s: %w", keyNode.Value, path, err)
			}

			if def.Condition == "" {
				return nil, fmt.Errorf("test %q in %s: condition is required", keyNode.Value, path)
			}

			tests = append(tests, Test{
				Name:       keyNode.Value,
				On:         def.On,
				Condition:  def.Condition,
				SourceFile: file,
			})
		}
	}

	return tests, nil
}
