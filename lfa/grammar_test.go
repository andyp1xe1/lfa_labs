package lfa

import (
	"testing"
)

func TestRandomWordsAreUnique(t *testing.T) {
	g := NewGrammarV5()

	randWords := make(map[string]bool, 0)
	for i := 0; i < 100; i++ {
		w := g.GetUniqueRandomWord()

		if _, ok := randWords[w]; ok {
			t.Fatal("word: ", w, " was already generated")
		}

		randWords[w] = true
	}

	t.Logf("Generated successfully %v unique words", len(randWords))
}
