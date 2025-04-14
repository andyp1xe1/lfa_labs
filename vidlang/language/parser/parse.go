package parser

import (
	"fmt"
	"log"
	"strconv"
)

type Parser struct {
	Expressions chan Node

	lex       *lexer
	currItem  item
	peekItem  item
	peek2Item item
}

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

// nextItem advances the Parser to the next token, and sets the peekItem
func (p *Parser) nextItem() {
	if p.currItem.typ == itemError {
		panic(NewAstError(
			"lexical error: "+p.currItem.val,
			p.currItem.line, p.currItem.pos))
	}
	p.currItem = p.peekItem
	p.peekItem = p.peek2Item
	p.peek2Item = <-p.lex.items

}

func (p *Parser) errorf(format string, args ...any) {
	panic(NewAstError(
		fmt.Sprintf("syntax error: "+format, args...),
		p.currItem.line, p.currItem.pos))
}

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

var validArgs = map[itemType]bool{
	itemLeftBrace:  true,
	itemIdentifier: true,
	itemNumber:     true,
	itemString:     true,
	itemBool:       true,
}

func (p *Parser) parseCommand() NodeCommand {
	var node NodeCommand
	node.Name = p.currItem.val
	node.Args = make([]NodeValue, 0)
	for validArgs[p.peekItem.typ] && p.peekItem.typ != itemNewline {
		p.nextItem()
		assert(
			validArgs[p.currItem.typ],
			"parseCommand's loop should be entered with a valid arg, but got %s",
			p.currItem)
		assert(
			p.currItem.typ != itemNewline,
			"parseCommand's loop should not process newlines",
		)
		node.Args = append(node.Args, p.parseValue())
	}
	return node
}

var validValues = map[itemType]bool{
	itemLeftBrace:  true,
	itemIdentifier: true,
	itemNumber:     true,
	itemString:     true,
	itemBool:       true,
	itemSelfStar:   true,
	itemStream:     true,
}

// if it's a list, parse it
// else if it's a value,
// if it is a number and peek is an operation, parse math expression
// else parse the simple value
// if a subexpression continues, parse it and give it the parameter list
// if a pipeline continues, parse it, adding the list or value as input
// return the value
// else if it's a command, parse the pipeline, form an expression, return it
func (p *Parser) parseValue() NodeValue {
	assert(validValues[p.currItem.typ],
		"parseValue should be invoked with currItem at a simple value, list or subexpression got %s",
		p.currItem)

	var n NodeValue

	if p.currItem.typ == itemLeftBrace {
		n = p.parseSimpleValueList()
		assert(
			p.currItem.typ == itemRightBrace, "assumed that the list was terminated successfully by a right brace, but got %s -> %s", p.currItem, p.peekItem)
		if p.peekItem.typ == itemLeftParen {
			n = p.parseSubExpr(n)
		}
	} else if (p.currItem.typ == itemNumber || p.currItem.typ == itemLeftParen) && p.peekItem.typ == mathSymbols[p.peekItem.val] {
		n = p.parseMathExpression()
	} else {
		n = p.parseSimpleValue()
	}

	return n
}

// TODO maybe split valeus and expressions logic
func (p *Parser) parseAssignable() NodeValue {
	assert(validValues[p.currItem.typ] || p.currItem.typ > itemCommand,
		"parseValue should be invoked with currItem at a simple value, list, subexpression or command, got %s",
		p.currItem)

	var n NodeValue

	if validValues[p.currItem.typ] {
		n = p.parseValue()

		if p.peekItem.typ == itemNewline && p.peek2Item.typ == itemPipe {
			p.nextItem()
		}

		if p.peekItem.typ == itemPipe { // TODO:
			p.nextItem()
		}

		if p.currItem.typ == itemPipe {
			p.nextItem()
			if n.ValueType() != ValueList {
				n = NodeList[NodeValue]{n}
			}
			n = NodeExpr{Input: n.(NodeList[NodeValue]), Pipeline: p.parsePipeline()}
		}
	} else if p.currItem.typ > itemCommand {
		n = NodeExpr{Pipeline: p.parsePipeline(), Input: nil}
	}

	return n
}

// TODO hangle empty body case `()`
func (p *Parser) parseSubExprBody() NodeValue {
	assert(p.currItem.typ == itemLeftParen,
		"parseSubExprBody should be invoked with currItem at left paren (the beginning of its body), got %s",
		p.currItem)
	p.nextItem()

	n := p.parseAssignable()
	// TODO: This is a temporary fix! Very bad should be done something else
	if p.currItem.typ != itemRightParen {
		p.nextItem()
	}

	if p.currItem.typ != itemRightParen {
		p.errorf("expected right paren at the end of subexpression body, got %s -> %s", p.currItem, p.peekItem)
	}
	//p.nextItem()

	return n
}

func (p *Parser) parseSubExpr(v NodeValue) NodeSubExpr {

	assert(v.ValueType() == ValueList,
		"parseSubExpr's argument is assumed to be a list, but got %s", p.currItem)

	argList := make(NodeList[NodeIdent], 0)
	for _, arg := range v.(NodeList[NodeValue]) {
		if arg.ValueType() != ValueIdentifier {
			p.errorf("a subexpression's argument list must be a list of identifiers, but got %s", arg)
		}
		argList = append(argList, arg.(NodeIdent))
	}

	var n NodeSubExpr
	n.Params = argList

	p.nextItem()
	n.Body = p.parseSubExprBody()

	return n
}

func (p *Parser) parsePipeline() NodePipeline {
	node := make(NodePipeline, 0)

	assert(p.currItem.typ > itemCommand,
		"parsePipeline should be invoked with currItem at a command, got %s", p.currItem)

	for p.currItem.typ > itemCommand {

		node = append(node, p.parseCommand())
		if p.peekItem.typ == itemNewline && p.peek2Item.typ == itemPipe {
			p.nextItem()
		}
		if p.peekItem.typ != itemPipe {
			break
		}
		p.nextItem()
		if p.peekItem.typ < itemCommand {
			p.errorf("expected command after pipe, got %s", p.peekItem)
		}
		p.nextItem()
	}

	return node
}

// TODO: Rewrite this. it does not handle trailing commas, empty lists, and lists spanning multiple lines
func (p *Parser) parseSimpleValueList() NodeList[NodeValue] {
	list := make(NodeList[NodeValue], 0)
	assert(p.currItem.typ == itemLeftBrace, "parseSimpleValueList should be invoked with currItem at left brace, got %s", p.currItem)
	for p.currItem.typ != itemRightBrace {
		p.nextItem()
		list = append(list, p.parseSimpleValue())
		p.nextItem()
		if p.currItem.typ != itemComma {
			if p.currItem.typ != itemRightBrace {
				p.errorf("list not terminated properly, expected comma or right brace, got %s", p.currItem)
			}
			break
		}
	}
	assert(
		p.currItem.typ == itemRightBrace,
		"it was assumed that the list was terminated by a right brace but got %s -> %s", p.currItem, p.peekItem)
	return list
}

func (p *Parser) parseSimpleValue() NodeValue {
	assert(
		validValues[p.currItem.typ],
		"parseSimpleValue should be invoked with a valid value, but got %s", p.currItem,
	)
	var n NodeValue
	switch p.currItem.typ {
	case itemIdentifier, itemSelfStar, itemStream:
		n = NodeIdent(p.currItem.val)
	case itemNumber:
		n = NodeLiteralNumber(strToLiteralNumber(p.currItem.val))
	case itemBool:
		n = NodeLiteralBool(strToLiteralBool(p.currItem.val))
	case itemString:
		n = NodeLiteralString(p.currItem.val)
	default:
		p.errorf("parseSimpleValue should be invoked with a valid value, but got %s", p.currItem)
	}
	//p.nextItem()
	return n
}

func strToLiteralBool(s string) NodeLiteralBool { return s == "true" }
func strToLiteralNumber(s string) NodeLiteralNumber {
	n, err := strconv.ParseFloat(s, 64)
	assert(err == nil, "lexer mus provided a valid number, failed to parse number %s", s)
	return NodeLiteralNumber(n)
}

func (p *Parser) parseAssignment() NodeAssign {
	var node NodeAssign

	node.Dest = p.parseIdentList()

	if p.currItem.typ == itemDeclare {
		node.Define = true
	} else if p.currItem.typ != itemAssign {
		p.errorf("expected assignment or declaration, got %s", p.peekItem)
	}

	p.nextItem()

	for p.currItem.typ == itemNewline {
		p.nextItem()
	}

	node.Value = p.parseAssignable()

	return node
}

func (p *Parser) parseIdentList() NodeList[NodeIdent] {
	var idents NodeList[NodeIdent]
	for p.currItem.typ == itemIdentifier {

		idents = append(idents, NodeIdent(p.currItem.val))
		p.nextItem()

		if p.currItem.typ == itemComma {
			p.nextItem()
		}

	}
	return idents
}

// TODO: Respect operator precedence
func (p *Parser) parseMathExpression() NodeValue {
	return p.parseTerm()
}

func (p *Parser) parseTerm() NodeValue {
	node := p.parseFactor()
	for p.currItem.typ == itemPlus || p.currItem.typ == itemMinus {
		op := p.currItem.typ
		p.nextItem()
		node = NodeExprMath{Left: node, Op: op, Right: p.parseFactor()}
	}
	return node
}

func (p *Parser) parseFactor() NodeValue {
	node := p.parsePrimary()
	for p.currItem.typ == itemMult || p.currItem.typ == itemDiv {
		op := p.currItem.typ
		p.nextItem()
		node = NodeExprMath{Left: node, Op: op, Right: p.parsePrimary()}
	}
	return node
}

func (p *Parser) parsePrimary() NodeValue {
	switch p.currItem.typ {
	case itemIdentifier:
		node := NodeIdent(p.currItem.val)
		p.nextItem()
		return node
	case itemNumber:
		node := strToLiteralNumber(p.currItem.val)
		p.nextItem()
		return node
	case itemLeftParen:
		p.nextItem()
		node := p.parseMathExpression()
		if p.currItem.typ != itemRightParen {
			p.errorf("missing closing parenthesis")
		}
		p.nextItem()
		return node
	default:
		p.errorf("unexpected token in expression %s", p.currItem)
		return nil
	}
}

type AssertError struct {
	Message string
}

func (e AssertError) Error() string {
	return "Assertion Failed: " + e.Message
}

func assert(condition bool, msg string, args ...any) {
	if !condition {
		panic(AssertError{Message: fmt.Sprintf(msg, args...)})
	}
}
