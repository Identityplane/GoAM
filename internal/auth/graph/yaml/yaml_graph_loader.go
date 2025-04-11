package yaml

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

func LoadFlowFromYAMLString(content string) (*graph.FlowWithRoute, error) {
	var yflow yamlFlow
	if err := yaml.Unmarshal([]byte(content), &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	nodes := make(map[string]*graph.GraphNode)
	for name, yn := range yflow.Nodes {
		nodes[name] = &graph.GraphNode{
			Name:         name,
			Use:          yn.Use,
			Next:         yn.Next,
			CustomConfig: yn.CustomConfig,
		}
	}

	flow := &graph.FlowWithRoute{
		Route: yflow.Route,
		Flow: &graph.FlowDefinition{
			Name:  yflow.Name,
			Start: yflow.Start,
			Nodes: nodes,
		},
	}

	return flow, nil

}

func LoadFlowFromYAML(path string) (*graph.FlowDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var yflow yamlFlow
	if err := yaml.Unmarshal(data, &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	nodes := make(map[string]*graph.GraphNode)
	for name, yn := range yflow.Nodes {
		nodes[name] = &graph.GraphNode{
			Name:         name,
			Use:          yn.Use,
			Next:         yn.Next,
			CustomConfig: yn.CustomConfig,
		}
	}

	return &graph.FlowDefinition{
		Name:  yflow.Name,
		Start: yflow.Start,
		Nodes: nodes,
	}, nil
}

func LoadFlowsFromDir(dir string) ([]graph.FlowWithRoute, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read dir: %w", err)
	}

	var flows []graph.FlowWithRoute

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", file, err)
		}

		var y yamlFlow
		if err := yaml.Unmarshal(data, &y); err != nil {
			return nil, fmt.Errorf("invalid YAML in %s: %w", file, err)
		}

		nodes := make(map[string]*graph.GraphNode)
		for name, yn := range y.Nodes {
			nodes[name] = &graph.GraphNode{
				Name:         name,
				Use:          yn.Use,
				Next:         yn.Next,
				CustomConfig: yn.CustomConfig,
			}
		}

		flows = append(flows, graph.FlowWithRoute{
			Route: y.Route,
			Flow: &graph.FlowDefinition{
				Name:  y.Name,
				Start: y.Start,
				Nodes: nodes,
			},
		})
	}

	return flows, nil
}
