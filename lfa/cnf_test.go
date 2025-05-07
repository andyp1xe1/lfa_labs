package lfa

import (
	"fmt"
	"strings"
	"testing"
)

func assert(t *testing.T, condition bool, message string) {
	if !condition {
		t.Fatal(message)
	}
}

func cloneGrammar(g *GrammarV2) *GrammarV2 {
	clone := NewGrammarV2(g.StartSymbol, g.Epsilon, []string{}, []string{})
	for nt := range g.NonTerminals {
		clone.NonTerminals[nt] = true
	}
	for t := range g.Terminals {
		clone.Terminals[t] = true
	}
	for lhs, prods := range g.Productions {
		for _, rhs := range prods {
			rhsCopy := make([]string, len(rhs))
			copy(rhsCopy, rhs)
			clone.Productions[lhs] = append(clone.Productions[lhs], rhsCopy)
		}
	}
	return clone
}

func countProductions(g *GrammarV2) int {
	count := 0
	for _, prods := range g.Productions {
		count += len(prods)
	}
	return count
}

func makeTestGrammar() *GrammarV2 {
	g := NewGrammarV2("S", "ε",
		[]string{"S", "A", "B", "C", "D"},
		[]string{"d", "a", "b"})
	g.AddProduction("S", []string{"d", "B"})
	g.AddProduction("S", []string{"A"})
	g.AddProduction("A", []string{"d"})
	g.AddProduction("A", []string{"d", "S"})
	g.AddProduction("A", []string{"a", "B", "d", "B"})
	g.AddProduction("B", []string{"a"})
	g.AddProduction("B", []string{"a", "S"})
	g.AddProduction("B", []string{"A", "C"})
	g.AddProduction("D", []string{"A", "B"})
	g.AddProduction("C", []string{"b", "C"})
	g.AddProduction("C", []string{"ε"})
	return g
}

func TestGrammarNormalization(t *testing.T) {
	grammar := makeTestGrammar()

	t.Run("Original Grammar", func(t *testing.T) {
		t.Log("Original Grammar\n", grammar)
		assert(t, len(grammar.NonTerminals) == 5, "Should have 5 non-terminals")
		assert(t, len(grammar.Terminals) == 3, "Should have 3 terminals")
		assert(t, countProductions(grammar) == 11, "Should have 11 productions")
		assert(t, grammar.hasProduction("C", []string{"ε"}), "Should have C -> ε")
	})

	transformAndTest := func(name string, transform func(*GrammarV2), testFunc func(*testing.T, *GrammarV2)) {
		t.Run(name, func(t *testing.T) {
			g := cloneGrammar(grammar)
			transform(g)
			testFunc(t, g)
		})
	}

	transformAndTest("Eliminate Epsilon", func(g *GrammarV2) {
		g.EliminateEpsilon()
	}, func(t *testing.T, g *GrammarV2) {
		assert(t, !g.hasProduction("C", []string{"ε"}), "C -> ε should be eliminated")
		assert(t, g.hasProduction("B", []string{"A"}), "B -> A should be added")
		assert(t, g.hasProduction("S", []string{"d", "B"}), "S -> dB should still exist")
	})

	transformAndTest("Eliminate Renaming", func(g *GrammarV2) {
		g.EliminateEpsilon()
		g.EliminateRenaming()
	}, func(t *testing.T, g *GrammarV2) {
		assert(t, !g.hasProduction("S", []string{"A"}), "S -> A should be eliminated")
		assert(t, g.hasProduction("S", []string{"d"}), "S -> d should be added")
		assert(t, g.hasProduction("S", []string{"d", "S"}), "S -> dS should be added")
		assert(t, g.hasProduction("S", []string{"a", "B", "d", "B"}), "S -> aBdB should be added")
	})

	transformAndTest("Eliminate Inaccessible", func(g *GrammarV2) {
		g.EliminateEpsilon()
		g.EliminateRenaming()
		g.EliminateInaccessibleSymbols()
	}, func(t *testing.T, g *GrammarV2) {
		_, hasD := g.NonTerminals["D"]
		assert(t, !hasD, "D should be eliminated as inaccessible")
		assert(t, !g.hasProduction("D", []string{"A", "B"}), "D -> AB should be eliminated")
	})

	transformAndTest("Eliminate Non-Productive", func(g *GrammarV2) {
		g.EliminateEpsilon()
		g.EliminateRenaming()
		g.EliminateInaccessibleSymbols()
		g.EliminateNonProductiveSymbols()
	}, func(t *testing.T, g *GrammarV2) {
		for lhs, prods := range g.Productions {
			for _, rhs := range prods {
				for _, sym := range rhs {
					if sym != g.Epsilon {
						_, isNT := g.NonTerminals[sym]
						_, isT := g.Terminals[sym]
						assert(t, isNT || isT, fmt.Sprintf("Symbol %s in %s -> %s should be productive", sym, lhs, strings.Join(rhs, " ")))
					}
				}
			}
		}
	})

	transformAndTest("Convert to CNF", func(g *GrammarV2) {
		g.EliminateEpsilon()
		g.EliminateRenaming()
		g.EliminateInaccessibleSymbols()
		g.EliminateNonProductiveSymbols()
		g.ConvertToCNF()
	}, func(t *testing.T, g *GrammarV2) {
		validateCNF(t, g)
	})

	t.Run("Full Normalization", func(t *testing.T) {
		g := cloneGrammar(grammar)
		g.Normalize()

		t.Log("CNF Normalized Grammar\n", g)

		// no epsilons except maybe start symbol
		for lhs, prods := range g.Productions {
			for _, rhs := range prods {
				if len(rhs) == 1 && rhs[0] == g.Epsilon {
					assert(t, lhs == g.StartSymbol, "Only start symbol can have epsilon production")
				}
			}
		}

		// no unit productions
		for _, prods := range g.Productions {
			for _, rhs := range prods {
				if len(rhs) == 1 && g.NonTerminals[rhs[0]] {
					t.Fatal("Unit productions should be eliminated")
				}
			}
		}

		// must be CNF
		validateCNF(t, g)
	})
}

func validateCNF(t *testing.T, g *GrammarV2) {
	for lhs, prods := range g.Productions {
		for _, rhs := range prods {
			// skip S -> ε
			if lhs == g.StartSymbol && len(rhs) == 1 && rhs[0] == g.Epsilon {
				continue
			}
			isTerminal := len(rhs) == 1 && g.Terminals[rhs[0]]
			isNonTermPair := len(rhs) == 2 && g.NonTerminals[rhs[0]] && g.NonTerminals[rhs[1]]
			assert(t, isTerminal || isNonTermPair, fmt.Sprintf("Production %s -> %s is not in CNF", lhs, strings.Join(rhs, " ")))
		}
	}
}
