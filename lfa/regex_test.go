package lfa

import (
	"regexp"
	"testing"
)

func TestRegexGenWords(t *testing.T) {
	// Test regex patterns from the image
	testCases := []struct {
		name    string
		pattern string
	}{
		{
			name:    "Pattern 1: (a|b)(c|d)E+G?",
			pattern: "(a|b)(c|d)E+G?",
		},
		{
			name:    "Pattern 2: P(Q|R|S)T(U|V|W|X)*Z+",
			pattern: "P(Q|R|S)T(U|V|W|X)*Z+",
		},
		{
			name:    "Pattern 3: 1(0|1)*2(3|4){5}36",
			pattern: "1(0|1)*2(3|4){5}36",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the regex pattern into an NFA
			nfa, err := CreateNFAFromRegex(tc.pattern)
			if err != nil {
				t.Fatalf("Failed to create NFA from regex '%s': %v", tc.pattern, err)
			}

			// Compile the same pattern using the built-in regex library for verification.
			re, err := regexp.Compile("^" + tc.pattern + "$")
			if err != nil {
				t.Fatalf("Failed to compile regex pattern '%s': %v", tc.pattern, err)
			}

			t.Logf("Random words for pattern '%s':", tc.pattern)
			uniqueWords := make(map[string]bool)
			maxAttempts := 100
			attempts := 0

			for len(uniqueWords) < 5 && attempts < maxAttempts {
				word, err := nfa.GenerateRandomWord(10)
				attempts++
				if err != nil {
					continue // Try again if generation failed
				}

				if !nfa.Accept(word) {
					t.Errorf("Generated word '%s' is not accepted by the NFA", word)
					continue
				}

				if !re.MatchString(word) {
					t.Errorf("Generated word '%s' does not match the regex pattern '%s'", word, tc.pattern)
					continue
				}

				if !uniqueWords[word] {
					uniqueWords[word] = true
					t.Logf("  Word %d: %s", len(uniqueWords), word)
				}
			}

			if len(uniqueWords) < 5 {
				t.Logf("Could only generate %d unique words after %d attempts", len(uniqueWords), attempts)
			}
		})
	}
}
