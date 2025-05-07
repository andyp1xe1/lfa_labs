# Laboratory work 6

## Parser Theory and Implementation

### What is a Parser?

The parser takes the stream of tokens produced by the lexer and constructs an Abstract Syntax Tree (AST). The AST represents the grammatical structure of the input script, making it easier to analyze, optimize, and execute. In our VidLang DSL, the parser is responsible for recognizing language constructs like assignments, expressions, commands, and pipelines.

### Lexer Token Types

Before diving into the parser implementation, let's understand the token types that the lexer produces. The lexer generates tokens with specific types defined as an `itemType` enum in `item.go`:

```go
type itemType int

const (
	itemError itemType = iota
	itemEOF

	itemIdentifier
	itemVariable

	// predefined identifiers
	itemSelfStar
	itemStream
	itemUnderscore

	// number operators
	itemDiv
	itemMinus
	itemMult
	itemPlus

	// stream operators
	itemAssign
	itemDeclare
	itemPipe

	// list operators
	itemConcatOp

	// literals
	itemNumber
	itemString
	itemBool

	// delimiters
	itemComma
	itemLeftBrace
	itemLeftParen
	itemNewline
	itemRightBrace
	itemRightParen

	// comment
	itemComment

	// commands
	itemCommand // to delimit commands
	itemBrigtness
	itemConcat
	itemContrast
	itemCrossfade
	itemCut
	itemExport
	itemFade
	itemHue
	itemMap
	itemOpen
	itemPitch
	itemSaturation
	itemSpeed
	itemTrackLine
	itemVolume
)
```

The lexer categorizes tokens into different types:
- Basic tokens like identifiers, variables, and literals
- Operators for numbers, streams, and lists
- Delimiters like parentheses, braces, and commas
- Command tokens for video processing operations

These token types are crucial for the parser to understand the structure of the input and build the AST accordingly.

### Recursive Descent Parsing

We've implemented a recursive descent parser. This approach mirrors the grammar of the language directly in the code, with each non-terminal in the grammar having a corresponding parsing function.

```go
func (p *Parser) parseCommand() NodeCommand { ... }
func (p *Parser) parseValue() NodeValue { ... }
func (p *Parser) parseAssignable() NodeValue { ... }
```

This direct mapping makes the parser relatively easy to understand and maintain. Each function is responsible for parsing a specific grammatical construct, and it calls other parsing functions to handle sub-constructs. Error handling can be localized to specific parsing functions, which simplifies debugging.

I chose recursive descent because it's a good balance between simplicity and expressiveness. It's easier to implement and debug than more complex parsing algorithms like LR or LALR, while still being powerful enough to handle our DSL.

### Parser Structure

The parser maintains several fields to track its progress through the token stream and manage its state:

```go
type Parser struct {
	Expressions chan Node      // channel for parsed expressions
	lex         *lexer         // lexer that produces tokens
	currItem    item           // current item
	peekItem    item           // next item
	peek2Item   item           // item after next
}
```

The `lex` field holds the lexer, which provides the stream of tokens. The `Expressions` channel is where the parser sends the constructed AST nodes. The `currItem`, `peekItem`, and `peek2Item` fields allow the parser to look ahead in the token stream. This lookahead is crucial for making parsing decisions, such as distinguishing between assignments and expressions.

I opted for a three-token lookahead (`currItem`, `peekItem`, `peek2Item`) because it was necessary to handle certain syntactic ambiguities in the language. The `Expressions` channel enables asynchronous processing of the script, potentially improving performance.

### Node Types

The AST is composed of different node types that implement the `Node` interface:

```go
type Node interface {
	// Common interface for all AST nodes
}

type NodeValue interface {
	Node
	ValueType() ValueType
}

type NodeLiteralString string
type NodeLiteralNumber float64
type NodeLiteralBool bool
type NodeIdent string
type NodeList[T any] []T
type NodeSubExpr struct { Body NodeValue; Params NodeList[NodeIdent] }
type NodeExprMath struct { Left NodeValue; Op OpType; Right NodeValue }
type NodeCommand struct { Name string; Args []NodeValue }
type NodePipeline []NodeCommand
type NodeExpr struct { Input NodeList[NodeValue]; Pipeline NodePipeline }
type NodeAssign struct { Dest NodeList[NodeIdent]; Value NodeValue; Define bool }
```

These node types represent the different values, expressions, commands, and control structures in VidLang.

### Parser Initialization

The `Parse` function initializes the parser and starts the parsing process in a goroutine:

```go
func Parse(input string) *Parser {
	p := &Parser{
		Expressions: make(chan Node),
		lex:         lex(input),
	}
	p.currItem = <-p.lex.items
	p.peekItem = <-p.lex.items
	p.peek2Item = <-p.lex.items
	go p.run()
	return p
}
```

The `Parse` function creates a new parser, initializes its fields, and starts the `run` method in a goroutine. The initial consumption of three items from the lexer channel primes the lookahead buffer.

### Main Parsing Loop

The `run` method is the heart of the parser:

```go
func (p *Parser) run() {
	defer func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case AstError:
				p.Expressions <- val
			default:
				log.Fatal(val)
			}
		}
	}()

	for i := 0; ; i++ {
		p.nextItem()
		switch p.currItem.typ {
		case itemEOF:
			close(p.Expressions)
			return
		case itemIdentifier:
			switch p.peekItem.typ {
			case itemAssign, itemDeclare, itemComma:
				p.Expressions <- p.parseAssignment()
			case itemPipe:
				p.Expressions <- p.parseAssignable()
			}
		case itemLeftBrace, itemNumber, itemString, itemBool:
			p.Expressions <- p.parseAssignable()
		default:
			if p.currItem.typ > itemCommand {
				p.Expressions <- p.parseAssignable()
			} else {
				// idk
			}
		}
	}
}
```

The `run` method iterates through the token stream, dispatching to different parsing functions based on the current token type. The lookahead is crucial for deciding which parsing function to call. An important feature is the use of deferred error handling, which recovers from panics thrown during parsing and sends the error as a node through the `Expressions` channel.

### Parsing Functions

The parser includes several parsing functions, each responsible for parsing a specific grammatical construct:

#### Error Handling

```go
func (p *Parser) errorf(format string, args ...any) {
	panic(NewAstError(
		fmt.Sprintf("syntax error: "+format, args...),
		p.currItem.line, p.currItem.pos))
}
```

The `errorf` function reports a parsing error by creating and panicking with an `AstError`. The error includes the line and position information, which helps users locate the error in their source code.

#### Assignments

```go
func (p *Parser) parseAssignment() NodeAssign {
	var node NodeAssign
	node.Dest = p.parseIdentList()
	// ...
	node.Value = p.parseAssignable()
	return node
}
```

The `parseAssignment` function parses assignment statements, including both variable declarations and regular assignments. It handles newlines between the assignment operator and the assigned value, which allows for multi-line assignments.

#### Values

```go
func (p *Parser) parseValue() NodeValue {
	// ...
	if p.currItem.typ == itemLeftBrace {
		n = p.parseSimpleValueList()
		// ...
	} else if (p.currItem.typ == itemNumber || p.currItem.typ == itemLeftParen) && p.peekItem.typ == mathSymbols[p.peekItem.val] {
		n = p.parseMathExpression()
	} else {
		n = p.parseSimpleValue()
	}
	return n
}
```

The `parseValue` function is responsible for parsing simple values, lists, and subexpressions.

#### Assignable Values

```go
func (p *Parser) parseAssignable() NodeValue {
	// ...
	if validValues[p.currItem.typ] {
		n = p.parseValue()
		// ...
		if p.currItem.typ == itemPipe {
			// ...
			n = NodeExpr{Input: n.(NodeList[NodeValue]), Pipeline: p.parsePipeline()}
		}
	} else if p.currItem.typ > itemCommand {
		n = NodeExpr{Pipeline: p.parsePipeline(), Input: nil}
	}
	return n
}
```

The `parseAssignable` function is responsible for parsing values that can be assigned to variables or used as inputs to pipelines.

#### Pipelines

```go
func (p *Parser) parsePipeline() NodePipeline {
	node := make(NodePipeline, 0)
	// ...
	for p.currItem.typ > itemCommand {
		node = append(node, p.parseCommand())
		// ...
	}
	return node
}
```

The `parsePipeline` function parses a pipeline of commands, which are chained together using the pipe operator (`|>`).

#### Commands

```go
func (p *Parser) parseCommand() NodeCommand {
	var node NodeCommand
	node.Name = p.currItem.val
	node.Args = make([]NodeValue, 0)
	// ...
	return node
}
```

The `parseCommand` function parses a command, including its name and arguments.

#### Math Expressions

```go
func (p *Parser) parseMathExpression() NodeValue {
	return p.parseTerm()
}

func (p *Parser) parseTerm() NodeValue {
	node := p.parseFactor()
	// ...
	return node
}

func (p *Parser) parseFactor() NodeValue {
	node := p.parsePrimary()
	// ...
	return node
}

func (p *Parser) parsePrimary() NodeValue {
	switch p.currItem.typ {
	case itemIdentifier:
		// ...
	case itemNumber:
		// ...
	case itemLeftParen:
		// ...
	default:
		// ...
	}
}
```

The `parseMathExpression`, `parseTerm`, `parseFactor`, and `parsePrimary` functions are responsible for parsing mathematical expressions. These functions implement a standard recursive descent parser for arithmetic expressions with operator precedence.

### Helper Functions

The parser includes several helper functions:

```go
func (p *Parser) nextItem() {
	// ...
}

func assert(condition bool, msg string, args ...any) {
	// ...
}
```

### Command Mapping

The parser uses a mapping from command names to item types to identify and process commands:

```go
var commands = map[string]itemType{
	"brightness": itemBrigtness,
	"concat":     itemConcat,
	"contrast":   itemContrast,
	// ...
}

func isCommand(s string) bool {
	_, ok := commands[s]
	return ok
}
```

This mapping allows the parser to check if a string represents a command and to get the corresponding item type.

## Design Considerations

### Three-Token Lookahead

The parser uses a three-token lookahead to handle syntactic ambiguities. This is essential for distinguishing between different language constructs, especially multi-line pipelines and complex expressions.

### Concurrent Operation

The parser operates concurrently with the lexer, using channels to communicate. This design allows the parser to process tokens as they are produced by the lexer, potentially improving performance for large scripts.

### Error Handling with Location Information

The parser includes robust error handling with detailed error messages that include source code locations. This makes it easier for users to locate and fix errors in their scripts.

## Usage Example

Let's examine how the parser processes a VidLang script snippet:

```go
func main() {
    // ...
    testScript := `
videoTrack =*
    |> brightness 1.3
    |> contrast 1.1
`
	p := parser.Parse(testScript)
	for expr := range p.Expressions {
		parser.PrintTree(expr, "")
	}
}
```

When running this example, you'll see each node printed with its type and value information. This output is invaluable for debugging the parser itself and understanding how the language is interpreted.
The generated output consists of the printed representation of the parsed nodes:

```
Assignment:
  Dest:
    videoTrack
  Value:
    Expression:
      Input:
        List (length 1):
          *
      Pipeline:
        Command 0:
            Name: brightness
            Args:
            1.3
        Command 1:
            Name: contrast
            Args:
            1.1
```

## Conclusions

The parser implementation for our VidLang DSL demonstrates several important principles of parsing:

1. **Well-defined Token Types**: The lexer provides a clear set of token types that the parser can use to understand the structure of the input.

2. **Recursive Descent Parsing**: Using a recursive descent parser makes the code easier to understand and maintain, with each parsing function handling a specific grammatical construct.

3. **Three-Token Lookahead**: The parser uses a three-token lookahead to handle syntactic ambiguities, such as distinguishing between assignments, expressions, and commands.

4. **Concurrent Operation**: The parser operates concurrently with the lexer, using channels to communicate, which can improve performance for large scripts.

5. **Robust Error Handling**: The parser includes detailed error messages with source code locations, making it easier for users to fix errors in their scripts.

The parser transforms the stream of tokens produced by the lexer into an AST, which represents the grammatical structure of the input script. This AST can then be used for further analysis, optimization, and execution by the interpreter or compiler.

Future improvements could include:
- Implementing more sophisticated semantic analysis, such as type checking and inference
- Enhancing the AST with additional information for optimization
