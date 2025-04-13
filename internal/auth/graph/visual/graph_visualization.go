package visual

import (
	"fmt"
	"goiam/internal/auth/graph"
	"log"
	"strings"
)

// RenderDOTGraph generates a Graphviz DOT representation of a flow.
func RenderDOTGraph(flow *graph.FlowDefinition) string {
	var b strings.Builder
	b.WriteString("digraph Flow {\n")
	b.WriteString(`  rankdir=LR;` + "\n")
	b.WriteString(fmt.Sprintf(`  label="%s"; labelloc=top; fontsize=20;`+"\n", flow.Name))

	for name, node := range flow.Nodes {
		def := graph.GetNodeDefinitionByName(node.Use)
		style := `shape=box`

		if def == nil {
			log.Panic("No node definiton found")
		}

		switch def.Type {
		case graph.NodeTypeInit:
			style = `shape=diamond, style=filled, fillcolor=lightgreen`
		case graph.NodeTypeLogic:
			style = `shape=ellipse, style=filled, fillcolor=lightyellow`
		case graph.NodeTypeQuery:
			style = `shape=rect, style=filled, fillcolor=lightblue`
		case graph.NodeTypeResult:
			style = `shape=doublecircle, style=filled, fillcolor=lightgray`
		}

		label := fmt.Sprintf("%s\n[%s]", node.Name, node.Use)
		b.WriteString(fmt.Sprintf(`  "%s" [label="%s", %s];`+"\n", name, label, style))

		for cond, next := range node.Next {
			b.WriteString(fmt.Sprintf(`  "%s" -> "%s" [label="%s"];`+"\n", name, next, cond))
		}
	}

	b.WriteString("}\n")
	return b.String()
}
