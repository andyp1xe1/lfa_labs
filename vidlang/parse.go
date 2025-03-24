package main

import "fmt"

type valueType int

const (
	valueString valueType = iota
	valueNumber
	valueBool

	valueVideo
	valueAudio
	valueImage
)

type nodeValue struct {
	val string
	typ valueType
}

type node interface{}

type nodeIdent string

type nodeList[T node] []T

type nodeCommand struct {
	name string
	args []node
}

type nodePipeline []nodeCommand

type nodeExpr struct {
	input nodeList[node]
	expr  nodePipeline
}

func (n nodeExpr) inferType() valueType {
	return valueString // TODO
}

type nodeAssign struct {
	dest   nodeList[nodeIdent]
	expr   nodeExpr
	define bool
}

type identifier struct {
	name string
	typ  string
}

type parser struct {
	lex         *lexer
	expressions chan node
	currItem    item
	peekItem    item
	identSet    map[string]*identifier
}

func parse(input string) *parser {
	p := &parser{
		lex:         lex(input),
		expressions: make(chan node),
		identSet:    make(map[string]*identifier),
	}
	go p.run()
	return p
}

// nextItem advances the parser to the next token, and sets the peekItem
func (p *parser) nextItem() {
	p.currItem = p.peekItem
	item := <-p.lex.items
	if item.typ != itemError {
		p.peekItem = item
		return
	}

	err := fmt.Errorf(item.val)
	fmt.Println(err)
	panic(err)
}

func (p *parser) errorf(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	fmt.Println(err)
	panic(err)
}

func (p *parser) run() {
	var currExpr node

	for {
		p.nextItem()
		switch p.currItem.typ {
		case itemEOF:
			close(p.expressions)
			return
		case itemNewline:
			continue
		case itemIdentifier: // in the future check if expression or assignment(?)
			currExpr = p.parseAssignment()
		case itemCommand:
			currExpr = p.parseCommand()
		}
		p.expressions <- currExpr
	}
}

var validArgs = map[itemType]bool{
	itemIdentifier: true,
	itemNumber:     true,
	itemString:     true,
	itemBool:       true,
}

func (p *parser) parseCommand() nodeCommand {
	var node nodeCommand
	node.name = p.currItem.val
	for validArgs[p.peekItem.typ] {
		p.nextItem()
		node.args = append(node.args, p.parseValue())
	}
	return node
}

func (p *parser) parseExpr() nodeExpr {
	var node nodeExpr

	if p.peekItem.typ == itemLeftBrace {
		node.input = p.parseList()
	}

	node.expr = p.parsePipeline()

	return node
}

func (p *parser) parsePipeline() nodePipeline {
	node := make(nodePipeline, 0)

	for {
		node = append(node, p.parseCommand())
		if p.peekItem.typ != itemPipe {
			break
		}
		p.nextItem()
		if p.peekItem.typ != itemCommand {
			p.errorf("expected command after pipe, got %s", p.peekItem)
		}
		p.nextItem()
	}

	return node
}

func (p *parser) parseList() nodeList[node] {
	var list nodeList[node]
	for p.currItem.typ != itemRightBrace {
		p.nextItem()
		list = append(list, p.parseValue())
		if p.currItem.typ != itemComma {
			if p.currItem.typ != itemRightBrace {
				p.errorf("list not terminated properly, expected comma or right brace, got %s", p.currItem)
			}
			break
		}

	}
	return list
}

func (p *parser) parseValue() node {
	return nil // TODO
}

func (p *parser) parseAssignment() nodeAssign {
	var node nodeAssign

	if p.peekItem.typ == itemDeclare {
		node.define = true
	} else if p.peekItem.typ != itemAssign {
		p.errorf("expected assignment or declaration, got %s", p.peekItem)
	}

	node.dest = p.parseIdentList()
	node.expr = p.parseExpr()

	return node
}

func (p *parser) parseIdentList() nodeList[nodeIdent] {
	var idents nodeList[nodeIdent]
	for {
		idents = append(idents, nodeIdent(p.currItem.val))

		p.nextItem()
		if p.currItem.typ != itemComma && p.peekItem.typ != itemIdentifier {
			break
		}
		p.nextItem()
	}
	return idents
}
