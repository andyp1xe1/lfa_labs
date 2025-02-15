package main

import (
	"log"
	"math/rand"
)

// Variant 5:
// VN={S, F, L},
// VT={a, b, c, d},
// P={
//     S → bS
//     S → aF
//     S → d
//     F → cF
//     F → dF
//     F → aL
//     L → aL
//     L → c
//     F → b
// }

var (
	uniqueWords = make(map[string]bool, 0)
)

type Grammar struct {
	vn []string
	vt []string
	p  map[string][]string
}

func NewGrammar() *Grammar {
	return &Grammar{
		vn: []string{"S", "F", "L"},
		vt: []string{"a", "b", "c", "d"},
		p: map[string][]string{
			"S": {"bS", "aF", "d"},
			"F": {"cF", "dF", "aL", "b"},
			"L": {"aL", "c"},
		},
	}
}

func (g *Grammar) getRandomWord() string {
	word := ""
	vn := g.vn[0]

	for {
		sufArr := g.p[vn]
		suf := sufArr[rand.Intn(len(sufArr))]

		if len(suf) == 1 {
			word += string(suf[0])
			return word
		}

		word += string(suf[0])
		vn = string(suf[1])
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
	dfa := NewDFA()

	dfa.Sigma = append(dfa.Sigma, g.vt...)
	dfa.Q = append(dfa.Q, g.vn...)

	for k, v := range g.p {
		dfa.delta[k] = make(map[string]string)

		for _, s := range v {
			if len(s) == 1 {
				dfa.delta[k][string(s[0])] = dfa.F
				continue
			}

			dfa.delta[k][string(s[0])] = string(s[1])
		}
	}

	return dfa
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

type DFA struct {
	Q     []string
	Sigma []string
	delta map[string]map[string]string
	q0    string
	F     string
}

func NewDFA() *DFA {
	dfa := &DFA{
		Q:     make([]string, 0),
		Sigma: make([]string, 0),
		delta: make(map[string]map[string]string),
		q0:    "S",
		F:     "X",
	}
	dfa.Q = append(dfa.Q, dfa.q0)
	dfa.Q = append(dfa.Q, dfa.F)

	return dfa
}

func (d *DFA) Accept(s string) bool {
	var ok bool
	q := d.q0

	for _, r := range s {

		if !contains(d.Sigma, string(r)) {
			log.Printf("Symbol %s not in Sigma", string(r))
			return false
		}

		if _, ok = d.delta[q]; !ok {
			log.Printf("State %s does not have outgoing transitions", q)
			return false
		}

		q, ok = d.delta[q][string(r)]
		if !ok {
			log.Printf("q = %s, r = %s", q, string(r))
			return false
		}
	}

	return q == d.F
}

func main() {
	g := NewGrammar()
	d := g.ToDFA()

	randWords := make([]string, 0)
	for i := 0; i < 10; i++ {
		randWords = append(randWords, g.GetUniqueRandomWord())
	}
	log.Println("Random words: ", randWords)

	for _, w := range randWords {
		if d.Accept(w) {
			log.Println("word: ", w, " Accepted")
		} else {
			log.Println("word: ", w, " Rejected")
		}
	}
}
