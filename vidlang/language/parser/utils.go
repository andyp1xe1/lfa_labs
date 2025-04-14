package parser

import "fmt"

func PrintTree(n interface{}, indent string) {
	switch node := n.(type) {

	case NodeAssign:
		fmt.Printf("%sAssignment:\n", indent)
		fmt.Printf("%sDest:\n", indent+"  ")
		// node.dest is a nodeList[nodeIdent]
		for _, dest := range node.Dest {
			PrintTree(dest, indent+"    ")
		}
		fmt.Printf("%sValue:\n", indent+"  ")
		PrintTree(node.Value, indent+"    ")

	case NodeExprMath:
		fmt.Printf("%sMath Expression: (Operator: %s)\n", indent, node.Op)
		fmt.Printf("%sLeft:\n", indent+"  ")
		PrintTree(node.Left, indent+"    ")
		fmt.Printf("%sRight:\n", indent+"  ")
		PrintTree(node.Right, indent+"    ")

	case NodeExpr:
		fmt.Printf("%sExpression:\n", indent)
		if node.Input != nil {
			fmt.Printf("%sInput:\n", indent+"  ")
			PrintTree(node.Input, indent+"    ")
		}
		if len(node.Pipeline) > 0 {
			fmt.Printf("%sPipeline:\n", indent+"  ")
			for i, cmd := range node.Pipeline {
				fmt.Printf("%sCommand %d:\n", indent+"    ", i)
				fmt.Printf("%s  Name: %s\n", indent+"      ", cmd.Name)
				if len(cmd.Args) > 0 {
					fmt.Printf("%s  Args:\n", indent+"      ")
					for _, arg := range cmd.Args {
						PrintTree(arg, indent+"        ")
					}
				}
			}
		}

	case NodeSubExpr:
		fmt.Printf("%sSub-Expression:\n", indent)
		fmt.Printf("%sParams:\n", indent+"  ")
		for _, param := range node.Params {
			PrintTree(param, indent+"    ")
		}
		fmt.Printf("%sBody:\n", indent+"  ")
		PrintTree(node.Body, indent+"    ")

	case NodeList[NodeValue]:
		fmt.Printf("%sList (length %d):\n", indent, len(node))
		for _, item := range node {
			PrintTree(item, indent+"  ")
		}

	// For literal values and identifiers
	case NodeLiteralString, NodeLiteralNumber, NodeLiteralBool, NodeIdent, NodeSelfStar:
		fmt.Printf("%s%v\n", indent, node)

	default:
		// If the node implements nodeValue, use its String method.
		if nv, ok := n.(NodeValue); ok {
			fmt.Printf("%s%s\n", indent, nv.String())
		} else {
			fmt.Printf("%sUnknown node: %#v\n", indent, n)
		}
	}
}
