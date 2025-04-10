package internal

import (
	"goiam/internal/auth/graph"
	"goiam/internal/auth/graph/yaml"
	"log"
)

var FlowsDir = "../config/flows"

var FlowRegistry = map[string]*graph.FlowWithRoute{}

func InitializeFlows() {

	flows, err := yaml.LoadFlowsFromDir(FlowsDir)
	if err != nil {
		log.Fatalf("failed to load flows: %v", err)
	}

	for _, flow := range flows {
		FlowRegistry[flow.Flow.Name] = &flow
	}

}
