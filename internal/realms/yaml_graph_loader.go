package realms

import (
	"fmt"
	"goiam/internal/auth/graph"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type yamlGraphNode struct {
	Use          string            `yaml:"use"`
	Next         map[string]string `yaml:"next"`
	CustomConfig map[string]string `yaml:"custom_config"`
}

type yamlFlow struct {
	Name  string                   `yaml:"name"`
	Route string                   `yaml:"route"`
	Start string                   `yaml:"start"`
	Nodes map[string]yamlGraphNode `yaml:"nodes"`
}

// Converts parsed yamlFlow into a graph.FlowWithRoute
func convertToFlowWithRoute(yf *yamlFlow) *FlowWithRoute {
	nodes := make(map[string]*graph.GraphNode, len(yf.Nodes))
	for name, yn := range yf.Nodes {
		nodes[name] = &graph.GraphNode{
			Name:         name,
			Use:          yn.Use,
			Next:         yn.Next,
			CustomConfig: yn.CustomConfig,
		}
	}

	return &FlowWithRoute{
		Route: yf.Route,
		Flow: &graph.FlowDefinition{
			Name:  yf.Name,
			Start: yf.Start,
			Nodes: nodes,
		},
	}
}

func LoadFlowFromYAMLString(content string) (*FlowWithRoute, error) {
	var yflow yamlFlow
	if err := yaml.Unmarshal([]byte(content), &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML content: %w", err)
	}
	return convertToFlowWithRoute(&yflow), nil
}

func LoadFlowFromYAML(path string) (*graph.FlowDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %w", path, err)
	}

	var yflow yamlFlow
	if err := yaml.Unmarshal(data, &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", path, err)
	}

	flowWithRoute := convertToFlowWithRoute(&yflow)
	return flowWithRoute.Flow, nil
}

func LoadFlowsFromDir(dir string) ([]FlowWithRoute, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read flows directory %s: %w", dir, err)
	}

	var flows []FlowWithRoute

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read flow file %s: %w", file, err)
		}

		var y yamlFlow
		if err := yaml.Unmarshal(data, &y); err != nil {
			return nil, fmt.Errorf("invalid YAML in flow file %s: %w", file, err)
		}

		flows = append(flows, *convertToFlowWithRoute(&y))
	}

	return flows, nil
}

// GetAllRealms returns a copy of the current loaded realms map.
func GetAllRealms() map[string]*LoadedRealm {
	loadedRealmsMu.RLock()
	defer loadedRealmsMu.RUnlock()

	// Shallow copy to prevent external mutation
	copy := make(map[string]*LoadedRealm, len(loadedRealms))
	for k, v := range loadedRealms {
		copy[k] = v
	}
	return copy
}
