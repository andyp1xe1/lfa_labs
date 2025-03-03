package lfa

func (g *Grammar) ClassifyGrammar() string {
	if isRegularGrammar(g) {
		return "Type 3: Regular Grammar"
	}

	if isContextFreeGrammar(g) {
		return "Type 2: Context-Free Grammar"
	}

	if isContextSensitiveGrammar(g) {
		return "Type 1: Context-Sensitive Grammar"
	}

	return "Type 0: Unrestricted Grammar"
}

// - A → a (terminal)
// - A → aB (terminal followed by non-terminal)
// - A → ε (epsilon, empty string)
func isRegularGrammar(g *Grammar) bool {
	// Check if it's right-linear or left-linear
	return isRightLinearGrammar(g) || isLeftLinearGrammar(g)
}

// isRightLinearGrammar checks if grammar is right-linear (A → a or A → aB)
func isRightLinearGrammar(g *Grammar) bool {
	for _, productions := range g.P {
		for _, prod := range productions {
			// Empty production (A → ε) is allowed
			if len(prod) == 0 {
				continue
			}

			// Check each production
			if len(prod) > 2 {
				return false // More than one non-terminal or too long
			}

			// First symbol must be terminal
			if !isTerminal(g, prod[0]) {
				return false
			}

			// If there's a second symbol, it must be a non-terminal
			if len(prod) == 2 && !isNonTerminal(g, string(prod[1])) {
				return false
			}
		}
	}
	return true
}

// isLeftLinearGrammar checks if grammar is left-linear (A → a or A → Ba)
func isLeftLinearGrammar(g *Grammar) bool {
	for _, productions := range g.P {
		for _, prod := range productions {
			// Empty production (A → ε) is allowed
			if len(prod) == 0 {
				continue
			}

			// Check each production
			if len(prod) > 2 {
				return false // More than one non-terminal or too long
			}

			// Last symbol must be terminal
			if !isTerminal(g, prod[len(prod)-1]) {
				return false
			}

			// If there's a first symbol (in a 2-char production), it must be a non-terminal
			if len(prod) == 2 && !isNonTerminal(g, string(prod[0])) {
				return false
			}
		}
	}
	return true
}

// isContextFreeGrammar checks if a grammar is context-free (Type 2)
// A context-free grammar has productions in the form:
// - A → α (where A is a single non-terminal and α is any string of terminals and non-terminals)
func isContextFreeGrammar(g *Grammar) bool {
	for nonTerminal, _ := range g.P {
		// The left side must be a single non-terminal
		if len(nonTerminal) != 1 || !isNonTerminal(g, nonTerminal) {
			return false
		}

		// In a context-free grammar, we don't need to check the right side
		// as any string of terminals and non-terminals is allowed,
		// as long as the left side is a single non-terminal
	}
	return true
}

// isContextSensitiveGrammar checks if a grammar is context-sensitive (Type 1)
// A context-sensitive grammar has productions in the form:
// - αAβ → αγβ (where A is a single non-terminal, α and β are strings, and γ is a non-empty string)
// This means the length of the right side is at least as long as the left side (except for S → ε)
func isContextSensitiveGrammar(g *Grammar) bool {
	for nonTerminal, productions := range g.P {
		for _, prod := range productions {
			// Special case: S → ε is allowed in CSG if S doesn't appear on the right side of any production
			if nonTerminal == g.S && len(prod) == 0 {
				if !nonTerminalAppearsOnRightSide(g, g.S) {
					continue
				}
				return false
			}

			// For all other productions, the right side must be at least as long as the left side
			if len(prod) < len(nonTerminal) {
				return false
			}
		}
	}

	// The grammar structure in this implementation makes it challenging to fully verify
	// the context-sensitive constraint (αAβ → αγβ) since we don't have the complete left side
	// of production rules stored with contexts. We're only checking length constraints.

	// A more comprehensive implementation would need a different Grammar structure
	// that can represent productions like αAβ → αγβ explicitly.

	return true
}

// Helper function to check if a character is a terminal
func isTerminal(g *Grammar, char byte) bool {
	for _, t := range g.Vt {
		if t == char {
			return true
		}
	}
	return false
}

// Helper function to check if a string is a non-terminal
func isNonTerminal(g *Grammar, s string) bool {
	for _, nt := range g.Vn {
		if nt == s {
			return true
		}
	}
	return false
}

func nonTerminalAppearsOnRightSide(g *Grammar, nonTerminal string) bool {
	for _, productions := range g.P {
		for _, prod := range productions {
			for i := 0; i < len(prod); i++ {
				// Check if this could be the start of a non-terminal
				if i <= len(prod)-len(nonTerminal) {
					if prod[i:i+len(nonTerminal)] == nonTerminal {
						return true
					}
				}
			}
		}
	}
	return false
}
