package interpreter

import (
	"fmt"

	"github.com/andyp1xe1/vidlang/language/parser"
)

var count = 0

type Env struct {
	debug  bool
	memory map[uint64]any
	vars   map[string]uint64
}

func NewEnv(debug bool) *Env {
	return &Env{debug, make(map[uint64]any), make(map[string]uint64)}
}

func store(e *Env, v parser.NodeValue) uint64 {
	p := uint64(count)
	count++
	e.memory[p] = v
	return p
}

func assign(e *Env, ident parser.NodeIdent, value parser.NodeValue) {
	e.vars[string(ident)] = store(e, value)
}

func (e *Env) Eval(script string) error {
	p := parser.Parse(script)

	for exp := range p.Expressions {
		if err, ok := exp.(parser.AstError); ok {
			return err
		}
		if err := eval(e, exp); err != nil {
			return err
		}
	}
	return nil
}

func eval(e *Env, exp parser.Node) (err error) {
	switch v := exp.(type) {
	case parser.NodeAssign:
		err = evalAssign(e, v)
	case parser.NodeExpr:
		err = evalExpr(e, v)
	default:
		if e.debug {
			fmt.Println("Ignoring expression: ", exp)
		}
	}
	return
}

func evalAssign(e *Env, exp parser.NodeAssign) error {
	if e.debug {
		fmt.Println("Evaluating assignment: ", exp)
	}
	return nil
}

func evalValue(e *Env, exp parser.NodeValue) error {
	switch v := exp.(type) {
	case parser.NodeExpr:
		return evalExpr(e, v)
	}
	return nil
}

func evalExpr(e *Env, exp parser.NodeExpr) error {
	if e.debug {
		fmt.Println("Evaluating expression: ", exp)
	}
	return nil
}
