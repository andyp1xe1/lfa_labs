package lfa

import (
	"fmt"
	"strconv"
)

// Regex represents a regular expression parser
type Regex struct {
	expression   string
	position     int
	stateCounter int
	statePrefix  string
	alphabet     []byte // Alphabet for wildcard character
}

// NewRegex creates a new regex parser
func NewRegex(expression string) *Regex {
	// Default alphabet includes lowercase letters, uppercase letters, and digits
	defaultAlphabet := []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	}

	return &Regex{
		expression:   expression,
		position:     0,
		stateCounter: 0,
		statePrefix:  "q",
		alphabet:     defaultAlphabet,
	}
}

// SetAlphabet allows setting the alphabet used for wildcard characters
func (r *Regex) SetAlphabet(alphabet []byte) {
	r.alphabet = alphabet
}

// Parse converts a regular expression to an NFA
func (r *Regex) Parse() (*NFA, error) {
	return r.parseExpression()
}

// parseExpression parses a full expression (possibly with alternation)
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

// parseTerm parses a sequence of factors
func (r *Regex) parseTerm() (*NFA, error) {
	// Handle empty term
	if r.position >= len(r.expression) ||
		r.expression[r.position] == ')' ||
		r.expression[r.position] == '|' {
		return CreateEmptyNFA(r.statePrefix, &r.stateCounter), nil
	}

	// Parse the first factor
	nfa, err := r.parseFactor()
	if err != nil {
		return nil, err
	}

	// Continue parsing factors and concatenate them
	for r.position < len(r.expression) &&
		r.expression[r.position] != ')' &&
		r.expression[r.position] != '|' {

		// Parse the next factor
		next, err := r.parseFactor()
		if err != nil {
			return nil, err
		}

		// Concatenate the current NFA with the new factor
		nfa = ConcatenateNFAs(nfa, next)
	}

	return nfa, nil
}

// parseFactor parses a basic unit with potential repetition (*,+,?,{n},^)
func (r *Regex) parseFactor() (*NFA, error) {
	var nfa *NFA
	var err error

	// Parse the basic unit
	if r.position >= len(r.expression) {
		return nil, fmt.Errorf("unexpected end of expression")
	}

	switch r.expression[r.position] {
	case '(':
		r.position++ // Skip '('
		nfa, err = r.parseExpression()
		if err != nil {
			return nil, err
		}

		if r.position >= len(r.expression) || r.expression[r.position] != ')' {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		r.position++ // Skip ')'

	case '.': // Dot matches any character
		r.position++
		nfa = CreateWildcardNFA(r.alphabet, r.statePrefix, &r.stateCounter)

	default:
		// Standard character
		char := r.expression[r.position]
		r.position++
		nfa = CreateBasicNFA(char, r.statePrefix, &r.stateCounter)
	}

	// Check for repetition operators
	if r.position < len(r.expression) {
		switch r.expression[r.position] {
		case '*':
			r.position++
			nfa = StarNFA(nfa, r.statePrefix, &r.stateCounter)

		case '+':
			r.position++
			nfa = PlusNFA(nfa, r.statePrefix, &r.stateCounter)

		case '?':
			r.position++
			nfa = QuestionNFA(nfa, r.statePrefix, &r.stateCounter)

		case '^': // Power operator for repetition (similar to {n})
			r.position++
			if r.position >= len(r.expression) {
				return nil, fmt.Errorf("expected number after ^ operator")
			}

			// Expect a number after ^ operator
			count, err := r.parseNumber()
			if err != nil {
				return nil, err
			}

			nfa = RepeatNFA(nfa, count)

		case '{':
			r.position++
			count, err := r.parseNumber()
			if err != nil {
				return nil, err
			}

			if r.position >= len(r.expression) || r.expression[r.position] != '}' {
				return nil, fmt.Errorf("missing closing brace")
			}
			r.position++ // Skip '}'

			nfa = RepeatNFA(nfa, count)
		}
	}

	return nfa, nil
}

// parseNumber parses a number in curly braces {n}
func (r *Regex) parseNumber() (int, error) {
	start := r.position

	// Find the end of the number
	for r.position < len(r.expression) &&
		r.expression[r.position] >= '0' &&
		r.expression[r.position] <= '9' {
		r.position++
	}

	if start == r.position {
		return 0, fmt.Errorf("expected number in quantifier")
	}

	// Parse the number
	number, err := strconv.Atoi(r.expression[start:r.position])
	if err != nil {
		return 0, fmt.Errorf("invalid number in quantifier: %v", err)
	}

	return number, nil
}

// CreateNFAFromRegex creates an NFA that accepts the language described by the regular expression
func CreateNFAFromRegex(pattern string) (*NFA, error) {
	regex := NewRegex(pattern)
	return regex.Parse()
}
