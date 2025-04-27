package service

import (
	"fmt"
	"goiam/internal/model"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type yamlFlow struct {
	FlowId     string             `yaml:"flow_id"`
	Route      string             `yaml:"route"`
	Active     bool               `yaml:"active"`
	Definition yamlFlowdefinition `yaml:"definition"`
}

type yamlFlowdefinition struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Start       string                   `yaml:"start"`
	Nodes       map[string]yamlGraphNode `yaml:"nodes"`
}

type yamlGraphNode struct {
	Use          string            `yaml:"use"`
	Next         map[string]string `yaml:"next"`
	CustomConfig map[string]string `yaml:"custom_config"`
}

// Converts parsed yamlFlow into a graph.FlowWithRoute
func convertToFlow(yf *yamlFlow) *model.Flow {

	nodes := make(map[string]*model.GraphNode, len(yf.Definition.Nodes))
	for name, yn := range yf.Definition.Nodes {
		nodes[name] = &model.GraphNode{
			Name:         name,
			Use:          yn.Use,
			Next:         yn.Next,
			CustomConfig: yn.CustomConfig,
		}
	}

	return &model.Flow{
		Id:     yf.FlowId,
		Route:  yf.Route,
		Active: yf.Active,
		Definition: &model.FlowDefinition{
			Name:        yf.Definition.Name,
			Description: yf.Definition.Description,
			Start:       yf.Definition.Start,
			Nodes:       nodes,
		},
	}
}

func LoadFlowFromYAMLString(content string) (*model.Flow, error) {
	var yflow yamlFlow
	if err := yaml.Unmarshal([]byte(content), &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML content: %w", err)
	}
	return convertToFlow(&yflow), nil
}

func LoadFlowFromYAML(path string) (*model.Flow, error) {
	data, err := os.ReadFile(path) // #nosec G304 (the path is trusted as it is not meant to be used with user input)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %w", path, err)
	}

	var yflow yamlFlow
	if err := yaml.Unmarshal(data, &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", path, err)
	}

	flow := convertToFlow(&yflow)
	return flow, nil
}

func LoadFlowsFromDir(dir string) ([]*model.Flow, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read flows directory %s: %w", dir, err)
	}

	var flows []*model.Flow

	for _, file := range files {
		data, err := os.ReadFile(file) // #nosec G304 (the dir load is trusted as it is not mean to used with user input)
		if err != nil {
			return nil, fmt.Errorf("failed to read flow file %s: %w", file, err)
		}

		var y yamlFlow
		if err := yaml.Unmarshal(data, &y); err != nil {
			return nil, fmt.Errorf("invalid YAML in flow file %s: %w", file, err)
		}

		flows = append(flows, convertToFlow(&y))
	}

	return flows, nil
}

// ConvertFlowToYAML converts a FlowDefinition to a YAML string
func ConvertFlowToYAML(flow *model.FlowDefinition) (string, error) {
	yamlBytes, err := yaml.Marshal(flow)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}
