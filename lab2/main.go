package main

import (
	"fmt"
	lfa "lfa_labs/lfa"
)

// Variant 5
// Q = {q0,q1,q2,q3},
// Sigma = {a,b},
// F = {q3},
// delta(q0,a) = q1,
// delta(q0,b) = q0,
// delta(q1,a) = q2,
// delta(q1,a) = q3,
// delta(q2,a) = q3,
// delta(q2,b) = q0.

func main() {
	g := lfa.NewGrammarV5()
	separator()
	fmt.Println("grammar type of v5: ", g.ClassifyGrammar())

	q0, q1, q2, q3 := "q0", "q1", "q2", "q3"
	Q := []lfa.State{q0, q1, q2, q3}

	Sigma := []byte{'a', 'b'}

	Delta := make(lfa.DeltaNfa)
	Delta.Add(q0, 'a', lfa.NewSetState(q1))
	Delta.Add(q0, 'b', lfa.NewSetState(q0))
	Delta.Add(q1, 'a', lfa.NewSetState(q2, q3))
	Delta.Add(q2, 'a', lfa.NewSetState(q3))
	Delta.Add(q2, 'b', lfa.NewSetState(q0))

	Q0 := []lfa.State{q0}
	F := []lfa.State{q3}

	nfa := lfa.NewNFA(Q, Sigma, Delta, Q0, F)
	dfa := nfa.ToDFA()

	separator()
	fmt.Println("The resulted DFA:")
	printDFA(dfa)

	separator()
	fmt.Println("is DFA: ", nfa.IsDFA())

	separator()
	fmt.Println("The NFA Grammar:")
	nfaG := nfa.ToGrammar()
	nfaG.Print()

	err := dfa.ToDOT("dfa.dot")
	if err != nil {
		fmt.Println(err)
	}
}

func separator() {
	fmt.Println()
	fmt.Println("-------------------")
	fmt.Println()
}

func printDFA(dfa *lfa.DFA) {
	fmt.Println("Start: ", dfa.Q0)
	fmt.Println("Final: ", dfa.F)
	for _, q := range dfa.Q {
		for _, r := range dfa.Sigma {
			qr := dfa.Delta.Lookup(q, r)
			if len(qr) > 0 {
				fmt.Println("(", q, string(r), ") -> ", qr)
			}
		}
	}
}
