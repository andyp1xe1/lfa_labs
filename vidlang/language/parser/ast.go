package parser

import (
	"fmt"
	"strings"
)

type AstError struct {
	Message string
	Line    int
	Pos     int
}

func (e AstError) Error() string {
	return fmt.Sprintf("%s at %d:%d", e.Message, e.Line, e.Pos)
}

func NewAstError(message string, line, pos int) AstError {
	return AstError{
		Message: message,
		Line:    line,
		Pos:     pos,
	}
}

type ValueType int

const (
	ValueLiteralBool ValueType = iota
	ValueLiteralNumber
	ValueLiteralString

	ValueGlobalStream
	ValueIdentifier
	ValueSelfStar

	ValueExpr
	ValueList
	ValueSubExpr
)

type NodeValue interface {
	ValueType() ValueType
	String() string
}

type NodeLiteralString string

func (s NodeLiteralString) ValueType() ValueType { return ValueLiteralString }
func (s NodeLiteralString) String() string       { return string(s) }

type NodeLiteralNumber float64

func (n NodeLiteralNumber) ValueType() ValueType { return ValueLiteralNumber }
func (n NodeLiteralNumber) String() string {
	// Format number without trailing zeros when it's a whole number
	if float64(n) == float64(int64(n)) {
		return fmt.Sprintf("%d", int64(n))
	}
	return fmt.Sprintf("%.6g", float64(n)) // More readable number format
}

type NodeLiteralBool bool

func (b NodeLiteralBool) ValueType() ValueType { return ValueLiteralBool }
func (b NodeLiteralBool) String() string       { return fmt.Sprintf("%t", bool(b)) }

type NodeIdent string

func (n NodeIdent) ValueType() ValueType { return ValueIdentifier }
func (n NodeIdent) String() string       { return string(n) }

type NodeSelfStar struct{ self string }

func (s NodeSelfStar) ValueType() ValueType { return ValueSelfStar }
func (s NodeSelfStar) String() string       { return s.self }

type NodeSubExpr struct {
	Body   NodeValue
	Params NodeList[NodeIdent]
}

func (s NodeSubExpr) ValueType() ValueType { return ValueSubExpr }
func (s NodeSubExpr) String() string {
	var params []string
	for _, p := range s.Params {
		params = append(params, p.String())
	}
	return fmt.Sprintf("[%s] -> %s", strings.Join(params, ", "), s.Body.String())
}

type NodeExprMath struct {
	Left  NodeValue
	Op    itemType
	Right NodeValue
}

func (n NodeExprMath) ValueType() ValueType { return ValueExpr }
func (n NodeExprMath) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Op.String(), n.Right.String())
}

type Node interface{}

type NodeList[T Node] []T

func (n NodeList[T]) ValueType() ValueType { return ValueList }
func (n NodeList[T]) String() string {
	var elements []string

	if len(n) == 0 {
		return "[]"
	}

	for _, elem := range n {
		elements = append(elements, fmt.Sprintf("%v", elem))
	}

	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

type NodeCommand struct {
	Name string
	Args []NodeValue
}

func (n NodeCommand) String() string {
	var args []string
	for _, arg := range n.Args {
		args = append(args, arg.String())
	}

	if len(args) == 0 {
		return n.Name
	}
	return fmt.Sprintf("%s(%s)", n.Name, strings.Join(args, ", "))
}

type NodePipeline []NodeCommand

func (n NodePipeline) String() string {
	var commands []string
	for _, cmd := range n {
		commands = append(commands, cmd.String())
	}
	return strings.Join(commands, " |> ")
}

type NodeExpr struct {
	Input    NodeList[NodeValue]
	Pipeline NodePipeline
}

func (n NodeExpr) ValueType() ValueType { return ValueExpr }
func (n NodeExpr) String() string {
	inputStr := n.Input.String()
	if len(n.Pipeline) == 0 {
		return inputStr
	}
	return fmt.Sprintf("%s |> %s", inputStr, n.Pipeline.String())
}

type NodeAssign struct {
	Dest   NodeList[NodeIdent]
	Value  NodeValue // some simple value, expr or subexpr
	Define bool
}

func (n NodeAssign) String() string {
	destStr := n.Dest.String()
	valueStr := n.Value.String()

	if n.Define {
		return fmt.Sprintf("%s := %s", destStr, valueStr)
	}
	return fmt.Sprintf("%s = %s", destStr, valueStr)
}

// Helper function to pretty print the entire AST
func PrettyPrintNode(n Node, indent string) string {
	switch node := n.(type) {
	case NodeAssign:
		operator := "="
		if node.Define {
			operator = ":="
		}
		return fmt.Sprintf("%sAssign: %s %s %s", indent, node.Dest.String(), operator, node.Value.String())

	case NodeExpr:
		return fmt.Sprintf("%sExpr: %s", indent, node.String())

	case NodeList[NodeValue], NodeList[NodeIdent]:
		if stringer, ok := any(node).(fmt.Stringer); ok {
			return fmt.Sprintf("%sList: %s", indent, stringer.String())
		}
		return fmt.Sprintf("%sList: %v", indent, node)

	case NodeCommand:
		return fmt.Sprintf("%sCommand: %s", indent, node.String())

	default:
		return fmt.Sprintf("%s%T: %v", indent, node, node)
	}
}

// Enhanced version of PrintNode that provides better formatting
func PrintNode(n Node) {
	fmt.Println(PrettyPrintNode(n, ""))
}

// Recursive pretty printer for debug purposes
func PrintNodeTree(n Node, indent string) {
	fmt.Println(PrettyPrintNode(n, indent))

	// Recursively print children if they exist
	switch node := n.(type) {
	case NodeAssign:
		fmt.Println(indent + "  Dest:")
		PrintNodeTree(node.Dest, indent+"    ")
		fmt.Println(indent + "  Value:")
		PrintNodeTree(node.Value, indent+"    ")

	case NodeExpr:
		fmt.Println(indent + "  Input:")
		PrintNodeTree(node.Input, indent+"    ")
		fmt.Println(indent + "  Pipeline:")
		for i, cmd := range node.Pipeline {
			fmt.Printf("%s  Command[%d]:\n", indent, i)
			PrintNodeTree(cmd, indent+"    ")
		}

	case NodeList[NodeValue]:
		for i, item := range node {
			fmt.Printf("%s  Item[%d]:\n", indent, i)
			PrintNodeTree(item, indent+"    ")
		}

	case NodeList[NodeIdent]:
		for i, item := range node {
			fmt.Printf("%s  Ident[%d]: %s\n", indent, i, item.String())
		}

	case NodeSubExpr:
		fmt.Println(indent + "  Params:")
		PrintNodeTree(node.Params, indent+"    ")
		fmt.Println(indent + "  Body:")
		PrintNodeTree(node.Body, indent+"    ")
	}
}
