package lfa

import (
	"math/rand"
)

var (
	uniqueWords = make(map[string]bool, 0)
)

// NonTerminal implements the State interface
type NonTerminal = string

type Grammar struct {
	S  NonTerminal
	Vn []NonTerminal
	Vt []byte
	P  map[NonTerminal][]string
}

func NewGrammarV5() *Grammar {
	return &Grammar{
		S:  "S",
		Vn: []NonTerminal{"S", "F", "L"},
		Vt: []byte{'a', 'b', 'c', 'd'},
		P: map[NonTerminal][]string{
			"S": {"bS", "aF", "d"},
			"F": {"cF", "dF", "aL", "b"},
			"L": {"aL", "c"},
		},
	}
}

func (g *Grammar) getRandomWord() string {
	word := ""
	vn := g.S

	for {
		sufArr := g.P[vn]
		suf := sufArr[rand.Intn(len(sufArr))]

		if len(suf) == 1 {
			word += string(suf[0])
			return word
		}

		word += string(suf[0])
		vn = NonTerminal(suf[1])
	}
}

func (g *Grammar) GetUniqueRandomWord() string {
	word := ""
	for {
		word = g.getRandomWord()

		if _, ok := uniqueWords[word]; !ok {
			uniqueWords[word] = true
			return word
		}
	}
}

func (g *Grammar) ToDFA() *DFA {
	var finalState NonTerminal = "X"

	q := make([]State, 0)
	delta := make(DeltaDFA)

	for _, v := range g.Vn {
		q = append(q, v)
	}

	for k, v := range g.P {
		delta[k] = make(map[byte]State)

		for _, s := range v {
			terminal := s[0]
			if len(s) == 1 {
				delta[k][terminal] = finalState
				continue
			}

			delta[k][terminal] = NonTerminal(s[1])
		}
	}

	dfa := NewDFA(q, g.Vt, delta, g.S, []State{finalState})
	return dfa
}
