package lfa

import (
	"testing"
)

func TestWordsRejected(t *testing.T) {
	g := NewGrammarV5()
	d := g.ToDFA()

	for _, w := range []string{"baas", "addc", "bacbbda"} {
		if d.Accept(w) {
			t.Fatal("word: ", w, " Accepted")
		} else {
			t.Log("word: ", w, " Rejected")
		}
	}
}

func TestRandomWordsAccepted(t *testing.T) {
	g := NewGrammarV5()
	d := g.ToDFA()

	randWords := make([]string, 0)
	for i := 0; i < 10; i++ {
		randWords = append(randWords, g.GetUniqueRandomWord())
	}
	t.Log("Random words: ", randWords)

	for _, w := range randWords {
		if d.Accept(w) {
			t.Log("word: ", w, " Accepted")
		} else {
			t.Fatal("word: ", w, " Rejected")
		}
	}
}
