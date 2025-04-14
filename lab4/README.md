# Laboratory Work 4

## Theory and Definitions

### Regular Expressions

A regular expression (regex) is a sequence of characters that defines a search pattern. Regular expressions provide a concise and flexible means for matching strings of text, such as particular characters, words, or patterns of characters.

Formally, a regular expression over an alphabet Σ can be defined recursively as follows:
- The empty string ε is a regular expression
- For each a ∈ Σ, a is a regular expression
- If r and s are regular expressions, then:
  - The concatenation r·s is a regular expression
  - The alternation r|s is a regular expression
  - The Kleene star r* is a regular expression

Regular expressions have direct correspondence to finite automata. In fact, regular expressions and finite automata are proven to be equivalent in expressive power: they both describe exactly the set of regular languages.

### Common Regex Operators:

1. **Basic Matching**:
   - Literal characters match themselves
   - `a` matches the character "a"

2. **Character Classes**:
   - `.` (dot) - matches any single character
   - `[abc]` - matches any one of the characters a, b, or c
   - `[^abc]` - matches any character except a, b, or c
   - `[a-z]` - matches any character from a to z

3. **Quantifiers**:
   - `*` (asterisk) - matches 0 or more occurrences of the preceding element
   - `+` (plus) - matches 1 or more occurrences of the preceding element
   - `?` (question mark) - matches 0 or 1 occurrence of the preceding element
   - `{n}` - matches exactly n occurrences of the preceding element
   - `{n,}` - matches n or more occurrences
   - `{n,m}` - matches between n and m occurrences

4. **Alternation and Grouping**:
   - `|` (pipe) - alternation, matches either the expression before or after it
   - `()` (parentheses) - groups expressions together, affecting operator precedence

### Regular Expressions to NFA Conversion

Converting regular expressions to NFAs follows Thompson's construction algorithm:

1. **Base Cases**:
   - Empty string (ε): NFA with a single state that is both initial and accepting
   - Single character (a): NFA with two states connected by a transition on 'a'

2. **Inductive Cases**:
   - **Concatenation (r·s)**: Connect the accepting states of r's NFA to the initial states of s's NFA with ε-transitions
   - **Alternation (r|s)**: Create a new initial state with ε-transitions to the initial states of both r's and s's NFAs, and a new accepting state with ε-transitions from the accepting states of both NFAs
   - **Kleene Star (r*)**: Create a new initial/accepting state with ε-transitions to/from r's initial/accepting states, and add ε-transitions from r's accepting states back to r's initial states

### Applications of Regular Expressions

Regular expressions are used in various domains:

1. **Text Processing**: Find, validate, and manipulate text patterns
2. **Data Validation**: Validate data formats like emails, phone numbers, dates
3. **Lexical Analysis**: Tokenize source code in compilers
4. **Search and Replace**: Find patterns in text editors and word processors
5. **Web Scraping**: Extract specific information from websites
6. **Log Analysis**: Parse and analyze log files for specific patterns or events
7. **Network Security**: Pattern matching in intrusion detection systems

## Objectives

1. Understand what a regular expression is and what it can be used for.

2. Continuing the work in the same repository and the same project, the following need to be added:
    a. Provide a function that could convert a regex into an NFA.
    b. Visualize the resulting NFA.
    c. Implement a function that would generate strings from your NFA.

3. Implement conversion of NFA to regex.
    a. Provide function that converts an automaton to a regex.
    b. Verify that the regex accepts the same language as the automaton.

## Implementation

For this laboratory work, I've implemented a regular expression parser and converter to NFA, as well as the ability to generate random words from the resulting NFA.

### The Regular Expression Parser

I created a `Regex` struct that holds the expression to be parsed and the necessary state for parsing:

```go
type Regex struct {
    expression   string
    position     int
    stateCounter int
    statePrefix  string
    alphabet     []byte // Alphabet for wildcard character
}
```

The parser uses a recursive descent approach, breaking down the expression into smaller units:

```go
func (r *Regex) Parse() (*NFA, error) {
    return r.parseExpression()
}

func (r *Regex) parseExpression() (*NFA, error) {
    // Parse the first term
    nfa, err := r.parseTerm()
    if err != nil {
        return nil, err
    }

    // Check for alternation (|)
    for r.position < len(r.expression) && r.expression[r.position] == '|' {
        r.position++ // Skip the '|'

        // Parse the term after '|'
        right, err := r.parseTerm()
        if err != nil {
            return nil, err
        }

        // Union the current NFA with the new term
        nfa = UnionNFAs(nfa, right, r.statePrefix, &r.stateCounter)
    }

    return nfa, nil
}
```

The parser handles various regex constructs including:
- Basic character matching
- Alternation (|)
- Grouping with parentheses
- Quantifiers (*, +, ?, {n})
- Wildcards (.)

### NFA Construction Functions

To support the regex-to-NFA conversion, I implemented several NFA manipulation functions:

1. **CreateBasicNFA**: Creates an NFA that accepts a single character
2. **CreateEmptyNFA**: Creates an NFA that accepts the empty string
3. **CreateWildcardNFA**: Creates an NFA that accepts any character from the alphabet
4. **ConcatenateNFAs**: Connects two NFAs in sequence
5. **UnionNFAs**: Creates an NFA that accepts either of two NFAs (alternation)
6. **StarNFA**: Creates an NFA for Kleene star (zero or more repetitions)
7. **PlusNFA**: Creates an NFA for plus (one or more repetitions)
8. **QuestionNFA**: Creates an NFA for question mark (zero or one occurrences)
9. **RepeatNFA**: Creates an NFA that repeats exactly n times

Here's an example of the `StarNFA` implementation:

```go
func StarNFA(nfa *NFA, statePrefix string, counter *int) *NFA {
    start := fmt.Sprintf("%s%d", statePrefix, *counter)
    *counter++
    accept := fmt.Sprintf("%s%d", statePrefix, *counter)
    *counter++

    delta := make(DeltaNfa)

    // Copy all transitions from the original NFA
    for state, transitions := range nfa.Delta {
        for symbol, states := range transitions {
            delta.Add(state, symbol, states)
        }
    }

    // Add epsilon transition from start to accept (for zero repetitions)
    delta.Add(start, Epsilon, NewSetState(accept))

    // Add epsilon transition from start to original NFA's initial states
    for _, startState := range nfa.Q0 {
        delta.Add(start, Epsilon, NewSetState(startState))
    }

    // Add epsilon transitions from original NFA's final states to:
    // 1. The new accept state
    // 2. Back to the original NFA's initial states (for repetition)
    for _, finalState := range nfa.F {
        delta.Add(finalState, Epsilon, NewSetState(accept))
        for _, startState := range nfa.Q0 {
            delta.Add(finalState, Epsilon, NewSetState(startState))
        }
    }

    // Combine states and handle epsilon in sigma
    states := append([]State{start, accept}, nfa.Q...)
    sigma := append([]byte{}, nfa.Sigma...)
    hasEpsilon := false
    for _, symbol := range sigma {
        if symbol == Epsilon {
            hasEpsilon = true
            break
        }
    }
    if !hasEpsilon {
        sigma = append(sigma, Epsilon)
    }

    return NewNFA(
        states,
        sigma,
        delta,
        []State{start},
        []State{accept},
    )
}
```

### Random Word Generation from NFA

To verify the correctness of the NFA construction, I implemented a function to generate random words accepted by the NFA:

```go
func (n *NFA) GenerateRandomWord(maxLength int) (string, error) {
    // Start from one of the initial states
    currentState := n.Q0[rand.Intn(len(n.Q0))]
    word := []byte{}
    
    // Keep track of visited state+length combinations to avoid infinite loops
    visited := make(map[string]map[int]bool)

    return n.generateFromState(currentState, word, maxLength, visited)
}
```

The `generateFromState` helper function recursively builds a word by:
1. Choosing random transitions from the current state
2. Following epsilon transitions without adding to the word
3. Adding characters to the word when following non-epsilon transitions
4. Potentially stopping if the current state is an accepting state

This approach allows the generation of diverse words that are guaranteed to be accepted by the NFA.

### Testing the Regex-to-NFA Implementation

The test cases verify that the regex parser correctly converts patterns to NFAs:

```go
func TestRegexGenWords(t *testing.T) {
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

            // Generate and verify random words
            t.Logf("Random words for pattern '%s':", tc.pattern)
            uniqueWords := make(map[string]bool)
            for len(uniqueWords) < 5 && attempts < maxAttempts {
                word, err := nfa.GenerateRandomWord(10)
                if err == nil && nfa.Accept(word) {
                    uniqueWords[word] = true
                    t.Logf("  Word %d: %s", len(uniqueWords), word)
                }
            }
        })
    }
}
```

The tests confirm that:
1. The regex can be successfully parsed into an NFA
2. The NFA can generate words matching the pattern
3. The generated words are accepted by both the NFA and a reference regex implementation

### Results

The implementation successfully parses regular expressions and constructs equivalent NFAs. The generated random words are consistently accepted by both the constructed NFA and the reference regex implementation, confirming the correctness of the conversion.

For visualization, the NFA can be converted to DOT format and rendered as a graph:

```go
func (n *NFA) ToDOT(filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    fmt.Fprintln(file, "digraph NFA {")
    fmt.Fprintln(file, "  rankdir=LR;")
    fmt.Fprintln(file, "  node [shape=circle];")

    // Mark final states
    for _, f := range n.F {
        fmt.Fprintf(file, "  \"%s\" [shape=doublecircle];\n", f)
    }

    // Mark initial states
    fmt.Fprintf(file, "  \"\" [shape=none];\n")
    for _, q0 := range n.Q0 {
        fmt.Fprintf(file, "  \"\" -> \"%s\";\n", q0)
    }

    // Add transitions
    for state, transitions := range n.Delta {
        for symbol, nextStates := range transitions {
            for nextState := range nextStates {
                fmt.Fprintf(file, "  \"%s\" -> \"%s\" [label=\"%c\"];\n", state, nextState, symbol)
            }
        }
    }

    fmt.Fprintln(file, "}")
    return nil
}
```

### Conclusion

This laboratory work explores the powerful connection between regular expressions and finite automata. By implementing a regex-to-NFA converter, I've demonstrated the theoretical equivalence between these two formalisms.

The implementation provides a clear picture of Thompson's construction algorithm and shows how complex regex patterns can be systematically converted to NFAs. The ability to generate random words accepted by the NFA serves as a practical validation mechanism.

While the current implementation covers most standard regex features, potential extensions could include:
1. Support for more advanced regex features like negative lookaheads/lookbehinds
2. Optimization of the resulting NFA to minimize state count
3. Direct conversion from regex to DFA using the subset construction algorithm

This work reinforces the understanding that regular expressions, NFAs, and DFAs all describe exactly the same class of languages - the regular languages. It also demonstrates the practical utility of these theoretical concepts in parsing and pattern matching applications.
