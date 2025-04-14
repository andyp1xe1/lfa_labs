package main

import (
	"fmt"
	lfa "lfa_labs/lfa"
	"os/exec"
	"strings"
)

// Demonstrates the step-by-step process of converting a regex to NFA
func showRegexProcessing(pattern string) {
	fmt.Printf("\n=== Processing Regex: %s ===\n\n", pattern)

	// Step 1: Thompson's Construction steps
	fmt.Println("Step 1: Thompson's Construction")
	showThompsonSteps(pattern)

	// Step 2: Generate the NFA
	fmt.Println("\nStep 2: Generating NFA")
	nfa, err := lfa.CreateNFAFromRegex(pattern)
	if err != nil {
		fmt.Printf("Error creating NFA: %v\n", err)
		return
	}

	// Visualize the NFA
	dotFile := fmt.Sprintf("regex_nfa_%s.dot", sanitizeFilename(pattern))
	if err := nfa.ToDOT(dotFile); err != nil {
		fmt.Printf("Error generating DOT file: %v\n", err)
		return
	}

	// Try to generate PNG
	pngFile := fmt.Sprintf("regex_nfa_%s.png", sanitizeFilename(pattern))
	cmd := exec.Command("dot", "-Tpng", dotFile, "-o", pngFile)
	if err := cmd.Run(); err == nil {
		fmt.Printf("NFA visualization saved to %s\n", pngFile)
	}

	// Step 3: Generate sample strings
	fmt.Println("\nStep 3: Sample strings accepted by this regex:")
	uniqueWords := make(map[string]bool)
	attempts := 0

	for len(uniqueWords) < 5 && attempts < 100 {
		word, err := nfa.GenerateRandomWord(10)
		attempts++
		if err != nil {
			continue
		}

		if nfa.Accept(word) && !uniqueWords[word] {
			uniqueWords[word] = true
			fmt.Printf("  - \"%s\"\n", word)
		}
	}
}

// Show Thompson's Construction steps specific to the regex
func showThompsonSteps(pattern string) {
	// Define epsilon symbol for printing
	const epsilonSymbol = "ε"
	// Base NFAs for all characters
	for i, c := range pattern {
		if isOperator(byte(c)) {
			continue
		}
		fmt.Printf("  • Basic NFA for '%c' at position %d\n", c, i)
	}

	// Process specific regex constructs based on the pattern
	if strings.Contains(pattern, "|") {
		parts := findAlternationParts(pattern)
		fmt.Printf("  • Creating alternation NFA for %s\n", parts)
		fmt.Printf("    - New start state with ε-transitions to start states of both alternatives\n")
		fmt.Printf("    - New accept state with ε-transitions from accept states of both alternatives\n")
	}

	// Handle parenthesized groups
	groups := findGroups(pattern)
	for i, group := range groups {
		fmt.Printf("  • Subexpression NFA for group %d: (%s)\n", i+1, group)
	}

	// Handle quantifiers (* + ? {n})
	for i := 0; i < len(pattern); i++ {
		if i+1 < len(pattern) {
			c := pattern[i]
			next := pattern[i+1]
			if !isOperator(byte(c)) && isQuantifier(byte(next)) {
				switch next {
				case '*':
					fmt.Printf("  • Kleene star (*) applied to '%c': creates loops with ε-transitions\n", c)
				case '+':
					fmt.Printf("  • Plus (+) applied to '%c': at least one occurrence\n", c)
				case '?':
					fmt.Printf("  • Optional (?) applied to '%c': adds ε-transition bypass\n", c)
				case '{':
					// Find closing brace
					end := strings.Index(pattern[i+1:], "}") + i + 1
					if end > i {
						fmt.Printf("  • Repetition %s applied to '%c': duplicating the NFA\n",
							pattern[i+1:end+1], c)
					}
				}
			}
		}
	}

	// Handle concatenation
	fmt.Println("  • Concatenating NFAs by adding ε-transitions between accept and start states")

	// Add pattern-specific insights
	if strings.Contains(pattern, "*") {
		fmt.Println("  • Creating Kleene star NFAs: ε-transitions allow 0 or more repetitions")
	}

	if strings.Contains(pattern, "(") && strings.Contains(pattern, ")") {
		fmt.Println("  • Connecting group NFAs by ε-transitions to main NFA")
	}

	// Add pattern-specific final steps
	switch pattern {
	case "(a|b)(c|d)E+G?":
		fmt.Println("  • (a|b) alternation → (c|d) alternation → 'E+' (one or more) → 'G?' (optional)")
	case "P(Q|R|S)T(U|V|W|X)*Z+":
		fmt.Println("  • 'P' NFA → (Q|R|S) alternation → 'T' NFA → (U|V|W|X)* (zero or more) → 'Z+' (one or more)")
	case "1(0|1)*2(3|4){5}36":
		fmt.Println("  • '1' NFA → (0|1)* (zero or more) → '2' NFA → (3|4){5} (exactly 5 times) → '3' NFA → '6' NFA")
	}
}

// Helper functions
func isOperator(c byte) bool {
	return c == '(' || c == ')' || c == '|' || c == '*' || c == '+' || c == '?' || c == '{' || c == '}'
}

func isQuantifier(c byte) bool {
	return c == '*' || c == '+' || c == '?' || c == '{'
}

func findGroups(pattern string) []string {
	var groups []string
	depth := 0
	start := -1

	for i, c := range pattern {
		if c == '(' {
			if depth == 0 {
				start = i + 1
			}
			depth++
		} else if c == ')' {
			depth--
			if depth == 0 && start != -1 {
				groups = append(groups, pattern[start:i])
				start = -1
			}
		}
	}

	return groups
}

func findAlternationParts(pattern string) string {
	// Simplified - just returns the pattern with highlighted alternation
	return strings.Replace(pattern, "|", "←|→", 1)
}

// Helper function to create a valid filename from a regex pattern
func sanitizeFilename(pattern string) string {
	replacer := strings.NewReplacer(
		"(", "_",
		")", "_",
		"|", "_OR_",
		"*", "_STAR_",
		"+", "_PLUS_",
		"?", "_QUEST_",
		"{", "_",
		"}", "_",
		" ", "_",
	)
	return replacer.Replace(pattern)
}

func main() {
	// List of regex patterns to process
	patterns := []string{
		"(a|b)(c|d)E+G?",
		"P(Q|R|S)T(U|V|W|X)*Z+",
		"1(0|1)*2(3|4){5}36",
	}

	// Process each pattern
	for _, pattern := range patterns {
		showRegexProcessing(pattern)
	}
}
