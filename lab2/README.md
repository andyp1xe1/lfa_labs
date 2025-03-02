# Laboratory Work 2

## Theory and Definitions

### Alphabet

A finite nonempty set of symbols. By convention denoted as $\Sigma$. For
example:

$$ \Sigma = \set{0, 1} \text{ binary alphabet} $$

$$ \Sigma = \set{ a, b, \dots, z } \text{ the set of lowercase letters} $$

### Grammar

A grammar is what defines the rules of a language, it is defined as a 5
sized toople:

$$ G = (V_{N}, V_{T}, P, S) $$

where:
- $V_{N}$ a finite set of *non-terminal symbols*  
- $V_T$ - a is a finite set of *terminal symbols*  
- $S$ - is the starting point  
- $P$ - is a finite set of productions of rules  

Additionally *terminals* and *non-terminals* don't have common symbols

$$V_N \cap V_T = \emptyset $$

### Production

A production is the set of rules a grammar follows. It is defined as:

$$ P = \set{ \alpha \to \beta \mid \alpha, \beta \in (V_{N} \cup V_{T})^*, \alpha \ne \epsilon } $$

### Deterministic Finite Automata

A 5-tuple:

$$ DFA = (Q, \Sigma, \delta, q_{0}, F) $$

where:
- $Q$ - a finite set of states $\Sigma$ - an alphabet
- $\delta : Q \times \Sigma \rightarrow Q$ (a transition function)
- $q_{0} \in Q$ - the initial state $F \subset Q$ - a set of accepting/final states

### Language of a DFA

the language of a DFA is the set of all strings over $\Sigma$ that
start from $q_{0}$, follow the transitions as the string is read
left to right and reach some accepting/final state.

### Equivalence of DFA with Regular Grammar

For the given $DFA = (Q, \Sigma, \delta, q_0, F)$ can be obtained an
equivalent regular grammar $G = (V_N, V_T, P, S)$.

Algorithm:

1.  $V_N = Q$
2.  $V_T = \Sigma$
3.  $S = \set{q_0}$
4.  For production $P$:
    - $P = \emptyset$
    - For all values:  
      $\delta(q, a) = (q_1, q_2, \dots, q_m)$
      we have  
      $P = P\cup \set{ q \rightarrow aq_i \mid  i = 1\dots m }$
    - for the values with  
      $F \cap \set{ q_1, q_2, \dots, q_m} \neq \emptyset$
      we have  
      $P = P\cup \set{q \rightarrow a}$


## Objectives

1. Discover what a language is and what it needs to have in order to be considered a formal one;

2. Provide the initial setup for the evolving project that you will work on during this semester. You can deal with each laboratory work as a separate task or project to demonstrate your understanding of the given themes, but you also can deal with labs as stages of making your own big solution, your own project. Do the following:

    a. Create GitHub repository to deal with storing and updating your project;

    b. Choose a programming language. Pick one that will be easiest for dealing with your tasks, you need to learn how to solve the problem itself, not everything around the problem (like setting up the project, launching it correctly and etc.);

    c. Store reports separately in a way to make verification of your work simpler (duh)

3. According to your variant number, get the grammar definition and do the following:

    a. Implement a type/class for your grammar;

    b. Add one function that would generate 5 valid strings from the language expressed by your given grammar;

    c. Implement some functionality that would convert and object of type Grammar to one of type Finite Automaton;

    d. For the Finite Automaton, please add a method that checks if an input string can be obtained via the state transition from it;

## Implementation

My variant is

```
Variant 5:
VN={S, F, L},
VT={a, b, c, d},
P={
    S → bS
    S → aF
    S → d
    F → cF
    F → dF
    F → aL
    L → aL
    L → c
    F → b
}
```

I chose Golang as the language, since I am most familiar with it at the moment, it is fast, and it has all the primitives I need for such a task: maps, lists, "methods" etc, while also being simple and not getting in my way with opinions on how classes should be done (there are no classes)

### The Structs

For the grammar and the DFA, I implemented structs closely to the definition in the theory.

The grammar struct:

```go
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
```

and the DFA struct:

```go
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
```

For the *delta* function I have used a "bidimentional" map, representing the transitions in the form of a map of maps. The key of the outer map is the state, and the key of the inner map is the symbol. The value of the inner map is the state it transitions to. This allows for fast lookups and validations.

### The Methods

First, the grammar has a method that converts itself to a DFA.
It implememnts the algorithm from the theory in the most straightforward way possible.
It assingns the states to the `Q` list, symbols to the `Sigma` list, and the transitions to the `delta` map.
Additionally it constructs the final state transitions:

```go
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
```

Then, the DFA has a method that checks if the input string can be obtained via the state transition from it.
It iterates through the string, and checks if the current state has a transition to the next symbol.
If the current state has a transition to the next symbol, it updates the current state to the next state.
If the current state does not have a transition to the next symbol, it returns false.
At the end, if the current state is the final state, it returns true.

```go
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
```

### Testing

In order to test my implememnation, I first need a a way to randomly generate strings from the grammar:

```go
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
```

Then, to the test:

```go
func TestRandomWordsAccepted(t *testing.T) {
	g := NewGrammar()
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
```

Here I have tested the algorithm with 10 random words, and checked if they are accepted by the DFA.

### Results:

Here are the results of running the tests:

![results](./img/results.png)

### Conclusion

In conclusion I can say that I don't regret my choice of using Go, as it allowed me to itarate fast and test my implementation without having to rely on external tools.

Additionally, I enjoyed learning and understanding more concretely what a State Machine is, and I definitely see its usecases.
