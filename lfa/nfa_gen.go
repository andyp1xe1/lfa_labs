package lfa

import (
	"fmt"
	"math/rand"
)

// GenerateRandomWord generates a random word accepted by the NFA
// maxLength is used to prevent infinite loops in NFAs with cycles
func (n *NFA) GenerateRandomWord(maxLength int) (string, error) {
	// rand.Seed(time.Now().UnixNano())
	word := []byte{}

	// Start from one of the initial states
	currentState := n.Q0[rand.Intn(len(n.Q0))]

	// Keep track of visited state+length combinations to avoid infinite loops
	visited := make(map[string]map[int]bool)

	return n.generateFromState(currentState, word, maxLength, visited)
}

// generateFromState is a helper function for GenerateRandomWord
func (n *NFA) generateFromState(state State, currentWord []byte, maxLength int, visited map[string]map[int]bool) (string, error) {
	// Check if we've already visited this state with this word length
	if visited[state] == nil {
		visited[state] = make(map[int]bool)
	}
	if visited[state][len(currentWord)] {
		return "", fmt.Errorf("detected cycle in NFA")
	}
	visited[state][len(currentWord)] = true

	// If we're in a final state, we can choose to stop
	// The probability of stopping increases as the word gets longer
	inFinalState := false
	for _, finalState := range n.F {
		if state == finalState {
			inFinalState = true
			break
		}
	}

	if inFinalState && len(currentWord) > 0 && rand.Float32() > float32(len(currentWord))/float32(maxLength) {
		return string(currentWord), nil
	}

	// Get all possible transitions from the current state
	possibleTransitions := make(map[byte][]State)
	if transitions, ok := n.Delta[state]; ok {
		for symbol, states := range transitions {
			for nextState := range states {
				possibleTransitions[symbol] = append(possibleTransitions[symbol], nextState)
			}
		}
	}

	// If no transitions are available and we're not in a final state, backtrack
	if len(possibleTransitions) == 0 && !inFinalState {
		return "", fmt.Errorf("reached dead end")
	}

	// Build a list of all possible next moves (with and without consuming input)
	type move struct {
		symbol byte
		state  State
	}
	possibleMoves := []move{}

	// First prioritize epsilon transitions
	if epsilonTransitions, ok := possibleTransitions[Epsilon]; ok {
		for _, nextState := range epsilonTransitions {
			possibleMoves = append(possibleMoves, move{Epsilon, nextState})
		}
	}

	// Then add non-epsilon transitions
	for symbol, states := range possibleTransitions {
		if symbol != Epsilon {
			for _, nextState := range states {
				possibleMoves = append(possibleMoves, move{symbol, nextState})
			}
		}
	}

	// If we're in a final state and have no moves, return the current word
	if len(possibleMoves) == 0 && inFinalState {
		return string(currentWord), nil
	}

	// Randomly select moves until we find a valid path or exhaust options
	for len(possibleMoves) > 0 {
		// Select a random move
		moveIndex := rand.Intn(len(possibleMoves))
		nextMove := possibleMoves[moveIndex]

		// Remove the selected move from options
		possibleMoves = append(possibleMoves[:moveIndex], possibleMoves[moveIndex+1:]...)

		// If it's an epsilon transition
		if nextMove.symbol == Epsilon {
			// Try to generate from the next state without adding to the word
			if result, err := n.generateFromState(nextMove.state, currentWord, maxLength, visited); err == nil {
				return result, nil
			}
			// If that path failed, continue with other options
		} else {
			// Check if adding this symbol would exceed max length
			if len(currentWord) >= maxLength {
				continue // Try other options that might be shorter
			}

			// Try to generate from the next state with the symbol added
			newWord := append(append([]byte{}, currentWord...), nextMove.symbol)
			if result, err := n.generateFromState(nextMove.state, newWord, maxLength, visited); err == nil {
				return result, nil
			}
		}
	}

	// If we're in a final state, we can return the current word
	if inFinalState {
		return string(currentWord), nil
	}

	return "", fmt.Errorf("no valid word found")
}
