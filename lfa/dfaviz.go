package lfa

import (
	"fmt"
	"os"
)

func (d *DFA) ToDOT(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "digraph DFA {")
	fmt.Fprintln(file, "  rankdir=LR;")
	fmt.Fprintln(file, "  node [shape=circle];")

	// Mark final states
	for _, f := range d.F {
		fmt.Fprintf(file, "  \"%s\" [shape=doublecircle];\n", f)
	}

	// Mark initial state
	fmt.Fprintf(file, "  \"\" [shape=none];\n")
	fmt.Fprintf(file, "  \"\" -> \"%s\";\n", d.Q0)

	// Add transitions
	for state, trans := range d.Delta {
		for symbol, nextState := range trans {
			fmt.Fprintf(file, "  \"%s\" -> \"%s\" [label=\"%c\"];\n", state, nextState, symbol)
		}
	}

	fmt.Fprintln(file, "}")
	return nil
}
