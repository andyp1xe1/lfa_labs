package lfa

import (
	"fmt"
	"os"
	"strings"
)

type GrammarV2 struct {
	NonTerminals map[string]bool
	Terminals    map[string]bool
	Productions  map[string][][]string
	StartSymbol  string
	Epsilon      string

	counter int
}

func NewGrammarV2(startSymbol, epsilon string, nt, t []string) *GrammarV2 {
	g := &GrammarV2{
		NonTerminals: make(map[string]bool),
		Terminals:    make(map[string]bool),
		Productions:  make(map[string][][]string),
		StartSymbol:  startSymbol,
		Epsilon:      epsilon,
		counter:      0,
	}
	for _, s := range nt {
		g.NonTerminals[s] = true
	}
	for _, s := range t {
		g.Terminals[s] = true
	}
	return g
}

func (g *GrammarV2) AddProduction(lhs string, rhs []string) {
	if !g.hasProduction(lhs, rhs) {
		g.Productions[lhs] = append(g.Productions[lhs], rhs)
	}
}

func (g *GrammarV2) hasProduction(lhs string, rhs []string) bool {
	for _, p := range g.Productions[lhs] {
		if slicesEqual(p, rhs) {
			return true
		}
	}
	return false
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// EliminateEpsilon removes epsilon productions and generates new productions
func (g *GrammarV2) EliminateEpsilon() {
	// Find nullable symbols
	nullable := make(map[string]bool)

	// Initial nullable symbols (direct ε productions)
	for lhs, prods := range g.Productions {
		for _, rhs := range prods {
			if len(rhs) == 1 && rhs[0] == g.Epsilon {
				nullable[lhs] = true
				break
			}
		}
	}

	// Propagate nullable symbols
	changed := true
	for changed {
		changed = false
		for lhs, prods := range g.Productions {
			if nullable[lhs] {
				continue
			}

			for _, rhs := range prods {
				allNullable := true
				for _, sym := range rhs {
					if sym != g.Epsilon && (!nullable[sym] || g.Terminals[sym]) {
						allNullable = false
						break
					}
				}

				if allNullable {
					nullable[lhs] = true
					changed = true
					break
				}
			}
		}
	}

	// Generate new productions without epsilon
	newProds := make(map[string][][]string)
	for lhs, prods := range g.Productions {
		for _, rhs := range prods {
			if len(rhs) == 1 && rhs[0] == g.Epsilon {
				continue // Skip direct epsilon productions
			}

			combinations := g.generateCombinations(rhs, nullable)
			for _, combo := range combinations {
				if len(combo) > 0 || lhs == g.StartSymbol {
					g.addProductionToMap(newProds, lhs, combo)
				}
			}
		}
	}

	// If start symbol is nullable, add S -> ε
	if nullable[g.StartSymbol] {
		g.addProductionToMap(newProds, g.StartSymbol, []string{g.Epsilon})
	}

	g.Productions = newProds
}

func (g *GrammarV2) generateCombinations(symbols []string, nullable map[string]bool) [][]string {
	result := [][]string{{}}

	for _, sym := range symbols {
		if sym == g.Epsilon {
			continue
		}

		newResult := [][]string{}
		for _, combo := range result {
			// Always add the symbol
			withSymbol := make([]string, len(combo))
			copy(withSymbol, combo)
			withSymbol = append(withSymbol, sym)
			newResult = append(newResult, withSymbol)

			// If nullable, also add the combination without it
			if nullable[sym] {
				withoutSymbol := make([]string, len(combo))
				copy(withoutSymbol, combo)
				newResult = append(newResult, withoutSymbol)
			}
		}
		result = newResult
	}

	return result
}

func (g *GrammarV2) addProductionToMap(prods map[string][][]string, lhs string, rhs []string) {
	if _, exists := prods[lhs]; !exists {
		prods[lhs] = [][]string{}
	}

	for _, existing := range prods[lhs] {
		if slicesEqual(existing, rhs) {
			return
		}
	}
	prods[lhs] = append(prods[lhs], rhs)
}

// EliminateRenaming removes unit productions (A -> B)
func (g *GrammarV2) EliminateRenaming() {
	// Find all unit pairs
	unitPairs := make(map[string]map[string]bool)

	// Initialize with direct unit productions
	for lhs := range g.NonTerminals {
		unitPairs[lhs] = make(map[string]bool)
		unitPairs[lhs][lhs] = true // A -> A is a unit pair
	}

	for lhs, prods := range g.Productions {
		for _, rhs := range prods {
			if len(rhs) == 1 && g.NonTerminals[rhs[0]] {
				unitPairs[lhs][rhs[0]] = true
			}
		}
	}

	// Compute transitive closure
	changed := true
	for changed {
		changed = false
		for a := range unitPairs {
			for b := range unitPairs[a] {
				for c := range unitPairs[b] {
					if !unitPairs[a][c] {
						unitPairs[a][c] = true
						changed = true
					}
				}
			}
		}
	}

	// Create new productions
	newProds := make(map[string][][]string)
	for lhs := range g.NonTerminals {
		newProds[lhs] = [][]string{}
	}

	for a := range unitPairs {
		for b := range unitPairs[a] {
			for _, rhs := range g.Productions[b] {
				if len(rhs) != 1 || !g.NonTerminals[rhs[0]] {
					g.addProductionToMap(newProds, a, rhs)
				}
			}
		}
	}

	g.Productions = newProds
}

// EliminateInaccessibleSymbols removes unreachable symbols
func (g *GrammarV2) EliminateInaccessibleSymbols() {
	accessible := make(map[string]bool)

	// Start with the start symbol
	accessible[g.StartSymbol] = true
	changed := true

	for changed {
		changed = false
		for lhs := range accessible {
			for _, rhs := range g.Productions[lhs] {
				for _, sym := range rhs {
					if g.NonTerminals[sym] && !accessible[sym] {
						accessible[sym] = true
						changed = true
					}
				}
			}
		}
	}

	// Keep only productions with accessible symbols
	newProds := make(map[string][][]string)
	for lhs := range accessible {
		if prods, exists := g.Productions[lhs]; exists {
			newProds[lhs] = prods
		}
	}

	g.Productions = newProds

	// Update non-terminals
	for nt := range g.NonTerminals {
		if !accessible[nt] {
			delete(g.NonTerminals, nt)
		}
	}
}

// EliminateNonProductiveSymbols removes symbols that don't derive terminals
func (g *GrammarV2) EliminateNonProductiveSymbols() {
	productive := make(map[string]bool)

	// All terminals are productive
	for t := range g.Terminals {
		productive[t] = true
	}

	// Find productive non-terminals
	changed := true
	for changed {
		changed = false
		for lhs, prods := range g.Productions {
			if productive[lhs] {
				continue
			}

			for _, rhs := range prods {
				allProductive := true
				for _, sym := range rhs {
					if sym != g.Epsilon && !productive[sym] {
						allProductive = false
						break
					}
				}

				if allProductive {
					productive[lhs] = true
					changed = true
					break
				}
			}
		}
	}

	// Keep only productive productions
	newProds := make(map[string][][]string)
	for lhs, prods := range g.Productions {
		if !productive[lhs] {
			continue
		}

		newProds[lhs] = [][]string{}
		for _, rhs := range prods {
			allProductive := true
			for _, sym := range rhs {
				if sym != g.Epsilon && !productive[sym] {
					allProductive = false
					break
				}
			}

			if allProductive {
				newProds[lhs] = append(newProds[lhs], rhs)
			}
		}
	}

	g.Productions = newProds

	// Update non-terminals
	for nt := range g.NonTerminals {
		if !productive[nt] {
			delete(g.NonTerminals, nt)
		}
	}
}

// ConvertToCNF converts grammar to Chomsky Normal Form
func (g *GrammarV2) ConvertToCNF() {
	// Step 1: Create terminal wrappers
	terminalWrappers := make(map[string]string)
	terminalProds := make(map[string][][]string)

	// First pass: collect terminal productions
	for _, prods := range g.Productions {
		for _, rhs := range prods {
			if len(rhs) > 1 {
				for _, sym := range rhs {
					if g.Terminals[sym] && terminalWrappers[sym] == "" {
						wrapper := fmt.Sprintf("T_%s", sym)
						terminalWrappers[sym] = wrapper
						g.NonTerminals[wrapper] = true
						terminalProds[wrapper] = [][]string{{sym}}
					}
				}
			}
		}
	}

	// Second pass: replace terminals in long productions
	newProds := make(map[string][][]string)
	for lhs, prods := range g.Productions {
		newProds[lhs] = [][]string{}

		for _, rhs := range prods {
			if len(rhs) <= 1 {
				// Keep A -> a and A -> B as is
				newProds[lhs] = append(newProds[lhs], rhs)
				continue
			}

			// Replace terminals in productions with length > 1
			newRhs := make([]string, len(rhs))
			for i, sym := range rhs {
				if g.Terminals[sym] {
					newRhs[i] = terminalWrappers[sym]
				} else {
					newRhs[i] = sym
				}
			}

			// Now break down long productions
			if len(newRhs) > 2 {
				currentLhs := lhs
				for i := 0; i < len(newRhs)-2; i++ {
					newNT := fmt.Sprintf("X%d", g.counter)
					g.counter++
					g.NonTerminals[newNT] = true
					newProds[currentLhs] = append(newProds[currentLhs], []string{newRhs[i], newNT})
					currentLhs = newNT
				}
				// Add the last production with the last two symbols
				newProds[currentLhs] = append(newProds[currentLhs], []string{newRhs[len(newRhs)-2], newRhs[len(newRhs)-1]})
			} else {
				// If length is 2, just add it
				newProds[lhs] = append(newProds[lhs], newRhs)
			}
		}
	}

	// Add terminal wrapper productions
	for wrapper, prods := range terminalProds {
		newProds[wrapper] = prods
	}

	g.Productions = newProds
}

// String returns a string representation of the grammar
func (g *GrammarV2) String() string {
	var sb strings.Builder

	sb.WriteString("Grammar:\n")
	sb.WriteString(fmt.Sprintf("  Start Symbol: %s\n", g.StartSymbol))
	sb.WriteString("  Non-Terminals: ")
	for nt := range g.NonTerminals {
		sb.WriteString(nt + " ")
	}
	sb.WriteString("\n  Terminals: ")
	for t := range g.Terminals {
		sb.WriteString(t + " ")
	}
	sb.WriteString("\n  Productions:\n")

	for lhs, prods := range g.Productions {
		for _, rhs := range prods {
			sb.WriteString(fmt.Sprintf("\t%s\t->\t", lhs))
			if len(rhs) == 0 {
				sb.WriteString(g.Epsilon)
			} else {
				sb.WriteString(strings.Join(rhs, " "))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Normalize applies all transformations to convert the grammar to CNF
func (g *GrammarV2) Normalize() {
	g.EliminateEpsilon()
	g.EliminateRenaming()
	g.EliminateInaccessibleSymbols()
	g.EliminateNonProductiveSymbols()
	g.ConvertToCNF()
}

// ToDOT generates a Graphviz DOT representation of the grammar
func (g *GrammarV2) ToDOT(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "digraph Grammar {")
	fmt.Fprintln(file, "  rankdir=LR;")
	fmt.Fprintln(file, "  node [fontname=\"Helvetica\", fontsize=12];")
	fmt.Fprintln(file, "  edge [fontname=\"Helvetica\", fontsize=10];")

	// Special start symbol
	fmt.Fprintln(file, "  \"start\" [shape=none, label=\"\"];")
	fmt.Fprintf(file, "  \"start\" -> \"%s\";\n", g.StartSymbol)

	// Nodes for non-terminals
	for nt := range g.NonTerminals {
		if nt == g.StartSymbol {
			fmt.Fprintf(file, "  \"%s\" [shape=circle, style=bold, peripheries=2];\n", nt)
		} else {
			fmt.Fprintf(file, "  \"%s\" [shape=circle];\n", nt)
		}
	}

	// Collect terminals
	terminals := make(map[string]bool)
	for _, prods := range g.Productions {
		for _, rhs := range prods {
			for _, sym := range rhs {
				if !g.NonTerminals[sym] && sym != g.Epsilon {
					terminals[sym] = true
				}
			}
		}
	}

	// Add terminal nodes
	for t := range terminals {
		fmt.Fprintf(file, "  \"%s\" [shape=plaintext];\n", t)
	}

	// Add epsilon
	if g.UsesEpsilon() {
		fmt.Fprintln(file, "  \"ε\" [shape=plaintext];")
	}

	// Productions as boxes
	prodCount := 0
	for lhs, productions := range g.Productions {
		for _, rhs := range productions {
			prodID := fmt.Sprintf("prod_%d", prodCount)
			prodCount++

			var rhsStr string
			if len(rhs) == 1 && rhs[0] == g.Epsilon {
				rhsStr = "ε"
			} else {
				rhsStr = strings.Join(rhs, " ")
			}
			productionLabel := fmt.Sprintf("%s → %s", lhs, rhsStr)

			// Production node
			fmt.Fprintf(file, "  \"%s\" [shape=box, label=\"%s\"];\n", prodID, productionLabel)

			// Connect lhs -> production
			fmt.Fprintf(file, "  \"%s\" -> \"%s\" [color=blue];\n", lhs, prodID)

			// Connect production -> each RHS symbol
			for _, sym := range rhs {
				if sym == g.Epsilon {
					fmt.Fprintf(file, "  \"%s\" -> \"ε\" [color=red, style=dashed];\n", prodID)
				} else {
					fmt.Fprintf(file, "  \"%s\" -> \"%s\" [color=red, style=dashed];\n", prodID, sym)
				}
			}
		}
	}

	// Optional: Legend
	fmt.Fprintln(file, "  subgraph cluster_legend {")
	fmt.Fprintln(file, "    label=\"Legend\"; style=dotted; fontsize=10;")
	fmt.Fprintln(file, "    \"legend_nt\" [shape=circle, label=\"Non-Terminal\"];")
	fmt.Fprintln(file, "    \"legend_start\" [shape=circle, style=bold, peripheries=2, label=\"Start Symbol\"];")
	fmt.Fprintln(file, "    \"legend_term\" [shape=plaintext, label=\"Terminal\"];")
	fmt.Fprintln(file, "    \"legend_prod\" [shape=box, label=\"Production\"];")
	fmt.Fprintln(file, "    \"legend_deriv\" [shape=plaintext, label=\"Blue: Derives\"];")
	fmt.Fprintln(file, "    \"legend_refs\" [shape=plaintext, label=\"Red: References\"];")
	fmt.Fprintln(file, "  }")

	fmt.Fprintln(file, "}")
	return nil
}

// helper
func (g *GrammarV2) UsesEpsilon() bool {
	for _, prods := range g.Productions {
		for _, rhs := range prods {
			if len(rhs) == 1 && rhs[0] == g.Epsilon {
				return true
			}
		}
	}
	return false
}
