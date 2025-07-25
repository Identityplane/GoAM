package visual

import (
	"fmt"
	"log"
	"strings"

	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/pkg/model"
)

// RenderDOTGraph generates a Graphviz DOT representation of a flow.
func RenderDOTGraph(flow *model.FlowDefinition) string {
	var b strings.Builder
	b.WriteString("digraph Flow {\n")
	b.WriteString(`  rankdir=LR;` + "\n")
	b.WriteString(fmt.Sprintf(`  label="%s"; labelloc=top; fontsize=20;`+"\n", flow.Description))

	for name, node := range flow.Nodes {
		def := graph.GetNodeDefinitionByName(node.Use)
		style := `shape=box`

		if def == nil {
			log.Panic("No node Definition found")
		}

		switch def.Type {
		case model.NodeTypeInit:
			style = `shape=diamond, style=filled, fillcolor=lightgreen`
		case model.NodeTypeLogic:
			style = `shape=ellipse, style=filled, fillcolor=lightyellow`
		case model.NodeTypeQuery:
			style = `shape=rect, style=filled, fillcolor=lightblue`
		case model.NodeTypeResult:
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
