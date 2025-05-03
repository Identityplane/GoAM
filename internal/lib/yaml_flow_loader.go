package lib

import (
	"fmt"
	"goiam/internal/model"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type yamlFlowDefinition struct {
	Description string                   `yaml:"description"`
	Start       string                   `yaml:"start"`
	Nodes       map[string]yamlGraphNode `yaml:"nodes"`
}

type yamlGraphNode struct {
	Name         string            `yaml:"name"`
	Use          string            `yaml:"use"`
	Next         map[string]string `yaml:"next"`
	CustomConfig map[string]string `yaml:"custom_config"`
}

func LoadFlowDefinitonFromString(content string) (*model.FlowDefinition, error) {
	var yflow yamlFlowDefinition
	if err := yaml.Unmarshal([]byte(content), &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML content: %w", err)
	}
	return yflow.convertToFlowDefinition(), nil
}

func (y *yamlFlowDefinition) convertToFlowDefinition() *model.FlowDefinition {
	nodes := make(map[string]*model.GraphNode, len(y.Nodes))
	for name, yn := range y.Nodes {
		nodes[name] = &model.GraphNode{
			Name:         yn.Name,
			Use:          yn.Use,
			Next:         yn.Next,
			CustomConfig: yn.CustomConfig,
		}
	}

	return &model.FlowDefinition{
		Description: y.Description,
		Start:       y.Start,
		Nodes:       nodes,
	}
}

func LoadFlowDefinitonsFromDir(dir string) ([]*model.FlowDefinition, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read flows directory %s: %w", dir, err)
	}

	var definitions []*model.FlowDefinition

	for _, file := range files {
		data, err := os.ReadFile(file) // #nosec G304 (the dir load is trusted as it is not mean to used with user input)
		if err != nil {
			return nil, fmt.Errorf("failed to read flow file %s: %w", file, err)
		}

		flow, err := LoadFlowDefinitonFromString(string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to load flow from YAML file %s: %w", file, err)
		}

		definitions = append(definitions, flow)
	}

	return definitions, nil
}

// ConvertFlowToYAML converts a FlowDefinition to a YAML string
func ConvertFlowToYAML(flow *model.FlowDefinition) (string, error) {
	yamlBytes, err := yaml.Marshal(flow)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}
