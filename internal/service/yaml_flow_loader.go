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
	Definition yamlFlowDefinition `yaml:"definition"`
}

type yamlFlowDefinition struct {
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

func (y *yamlFlowDefinition) ConvertToFlowDefinition() *model.FlowDefinition {

	nodes := make(map[string]*model.GraphNode, len(y.Nodes))
	for name, yn := range y.Nodes {
		nodes[name] = &model.GraphNode{
			Name:         name,
			Use:          yn.Use,
			Next:         yn.Next,
			CustomConfig: yn.CustomConfig,
		}
	}

	return &model.FlowDefinition{
		Name:        y.Name,
		Description: y.Description,
		Start:       y.Start,
		Nodes:       nodes,
	}
}

// Converts parsed yamlFlow into a graph.FlowWithRoute
func convertToFlow(yf *yamlFlow) *model.Flow {

	flowDefinition := yf.Definition.ConvertToFlowDefinition()

	return &model.Flow{
		Id:         yf.FlowId,
		Route:      yf.Route,
		Active:     yf.Active,
		Definition: flowDefinition,
	}
}

func LoadFlowFromYAMLString(content string) (*model.Flow, error) {
	var yflow yamlFlow
	if err := yaml.Unmarshal([]byte(content), &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML content: %w", err)
	}

	flow := convertToFlow(&yflow)

	// The flow yaml is not the same as the flow definition yaml
	// we need to store the definition yaml but as we don't have that here, we create it
	definitionYaml, err := yaml.Marshal(flow.Definition)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow definition: %w", err)
	}
	flow.DefintionYaml = string(definitionYaml)

	return flow, nil
}

func LoadFlowDefinitonFromString(content string) (*model.FlowDefinition, error) {
	var yflow yamlFlowDefinition
	if err := yaml.Unmarshal([]byte(content), &yflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML content: %w", err)
	}
	return yflow.ConvertToFlowDefinition(), nil
}

func LoadFlowFromYAML(path string) (*model.Flow, error) {
	data, err := os.ReadFile(path) // #nosec G304 (the path is trusted as it is not meant to be used with user input)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %w", path, err)
	}

	return LoadFlowFromYAMLString(string(data))
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

		flow, err := LoadFlowFromYAMLString(string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to load flow from YAML file %s: %w", file, err)
		}

		flows = append(flows, flow)
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
