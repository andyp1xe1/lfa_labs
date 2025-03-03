package lfa

import (
	"log"
)

type State = string

func contains[T comparable](list []T, item T) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

type DeltaDFA map[State]map[byte]State

func (d DeltaDFA) Add(in State, r byte, out State) {
	key := in
	if _, ok := d[key]; !ok {
		d[key] = make(map[byte]State)
	}
	d[key][r] = out
}

func (d DeltaDFA) Lookup(in State, r byte) State {
	if q, ok := d[in]; !ok {
		return ""
	} else {
		return q[r]
	}

}

func (d DeltaDFA) LookupQ(in State) map[byte]State {
	if q, ok := d[in]; !ok {
		return nil
	} else {
		return q
	}
}

type DFA struct {
	Q     []State
	Sigma []byte
	Delta DeltaDFA
	Q0    State
	F     []State
}

func NewDFA(q []State, sigma []byte, delta DeltaDFA, q0 State, f []State) *DFA {
	dfa := &DFA{
		Q:     q,
		Sigma: sigma,
		Delta: delta,
		Q0:    q0,
		F:     f,
	}

	return dfa
}

func (d *DFA) Accept(s string) bool {
	q := d.Q0

	for _, char := range s {
		symbol := byte(char)

		if !contains(d.Sigma, symbol) {
			log.Printf("Symbol %v not in Sigma", symbol)
			return false
		}

		if m := d.Delta.LookupQ(q); m == nil {
			log.Printf("State %s does not have outgoing transitions", q)
			return false
		}

		q = d.Delta.Lookup(q, symbol)
		if len(q) == 0 {
			log.Printf("q = %v, r = %v", q, symbol)
			return false
		}
	}

	return contains(d.F, q)
}

func (d *DFA) ToGrammar() *Grammar {
	grammar := &Grammar{
		Vn: d.Q,
		Vt: d.Sigma,
		P:  make(map[State][]string),
		S:  d.Q0,
	}

	for state, transitions := range d.Delta {
		for symbol, nextState := range transitions {
			grammar.P[state] = append(grammar.P[state], string(symbol)+nextState)
		}
	}

	// for _, finalState := range d.F {
	// 	grammar.P[finalState] = append(grammar.P[finalState], "")
	// }

	return grammar
}
