package lfa

import (
	"fmt"
	"sort"
	"strings"
)

// Epsilon represents an epsilon (ε) transition in an NFA
const Epsilon = byte(255)

type DeltaNfa map[State]map[byte]setState

func (d DeltaNfa) Add(in State, r byte, out setState) {
	if _, ok := d[in]; !ok {
		d[in] = make(map[byte]setState)
	}

	if existing, ok := d[in][r]; ok {
		// If there are already transitions for this symbol, merge them
		for state := range out {
			existing[state] = true
		}
	} else {
		// Otherwise, add the new transition
		d[in][r] = out
	}
}

func (d DeltaNfa) Lookup(in State, r byte) setState {
	if q, ok := d[in]; !ok {
		return nil
	} else {
		return q[r]
	}
}

type setState map[string]bool

func NewSetState(states ...State) setState {
	s := make(setState)
	for _, state := range states {
		s[state] = true
	}
	return s
}

func (s setState) toState() State {
	buff := make([]string, 0)
	for q := range s {
		buff = append(buff, q)
	}

	sort.Slice(buff, func(i, j int) bool {
		return buff[i] < buff[j]
	})

	return "{" + strings.Join(buff, ",") + "}"
}

func (s setState) Add(state State) {
	s[state] = true
}

func (s setState) Union(s2 setState) {
	for k := range s2 {
		s[k] = true
	}
}

func (s setState) Equals(s2 setState) bool {
	if len(s) != len(s2) {
		return false
	}
	for k := range s {
		if _, ok := s2[k]; !ok {
			return false
		}
	}
	return true
}

type NFA struct {
	Q     []State
	Sigma []byte
	Delta DeltaNfa
	Q0    []State
	F     []State
}

func NewNFA(Q []State, Sigma []byte, Delta DeltaNfa, Q0 []State, F []State) *NFA {
	// Ensure that all symbols used in Delta are in Sigma
	symbolSet := make(map[byte]bool)
	for _, symbol := range Sigma {
		symbolSet[symbol] = true
	}

	for _, transitions := range Delta {
		for symbol := range transitions {
			if !symbolSet[symbol] && symbol != Epsilon {
				Sigma = append(Sigma, symbol)
				symbolSet[symbol] = true
			}
		}
	}

	nfa := &NFA{
		Q:     Q,
		Sigma: Sigma,
		Delta: Delta,
		Q0:    Q0,
		F:     F,
	}

	return nfa
}

func (n *NFA) IsDFA() bool {
	for _, transitions := range n.Delta {
		symbolsSeen := make(map[byte]bool)

		for symbol, destinations := range transitions {
			if symbolsSeen[symbol] {
				return false
			}
			symbolsSeen[symbol] = true

			if len(destinations) != 1 {
				return false
			}
		}

		for _, symbol := range n.Sigma {
			if !symbolsSeen[symbol] {
				return false
			}
		}
	}

	for _, state := range n.Q {
		if _, exists := n.Delta[state]; !exists {
			return false
		}
	}

	return true
}

// EpsilonClosure computes the set of states reachable from a state through epsilon transitions
func (n *NFA) EpsilonClosure(state State) setState {
	closure := NewSetState(state)
	stack := []State{state}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Find all states reachable through epsilon transitions
		if nextStates := n.Delta.Lookup(current, Epsilon); nextStates != nil {
			for nextState := range nextStates {
				if _, exists := closure[nextState]; !exists {
					closure[nextState] = true
					stack = append(stack, nextState)
				}
			}
		}
	}

	return closure
}

// EpsilonClosureSet computes the epsilon closure for a set of states
func (n *NFA) EpsilonClosureSet(states setState) setState {
	closure := make(setState)

	for state := range states {
		stateClosure := n.EpsilonClosure(state)
		closure.Union(stateClosure)
	}

	return closure
}

// Accept checks if an input string is accepted by the NFA
func (n *NFA) Accept(s string) bool {
	currentStates := make(setState)
	for _, state := range n.Q0 {
		currentStates[state] = true
	}
	currentStates = n.EpsilonClosureSet(currentStates)

	for _, r := range s {
		symbol := byte(r)

		nextStates := make(setState)
		for state := range currentStates {
			if reachable := n.Delta.Lookup(state, symbol); reachable != nil {
				nextStates.Union(reachable)
			}
		}

		currentStates = n.EpsilonClosureSet(nextStates)

		if len(currentStates) == 0 {
			return false
		}
	}

	for _, acceptState := range n.F {
		if currentStates[acceptState] {
			return true
		}
	}

	return false
}

func (n *NFA) ToDFA() *DFA {
	nfaQueue := newNFAQueue()
	deltaPrime := make(DeltaDFA)

	qPrime := make([]State, 0)
	fPrime := make([]State, 0)

	// Start with epsilon closure of initial states
	q0 := make(setState)
	for _, q := range n.Q0 {
		q0.Add(q)
	}
	q0 = n.EpsilonClosureSet(q0)

	nfaQueue.enqueue(q0)
	qPrime = append(qPrime, q0.toState())

	for _, f := range n.F {
		if q0[f] {
			fPrime = append(fPrime, q0.toState())
			break
		}
	}

	dfa_sigma := make([]byte, 0)
	for _, symbol := range n.Sigma {
		if symbol != Epsilon {
			dfa_sigma = append(dfa_sigma, symbol)
		}
	}

	for !nfaQueue.done() {
		currentSetState := nfaQueue.dequeue()

		for _, r := range dfa_sigma {
			out := make(setState)

			for state := range currentSetState {
				if nextStates := n.Delta.Lookup(state, r); nextStates != nil {
					out.Union(nextStates)
				}
			}

			out = n.EpsilonClosureSet(out)

			if len(out) > 0 {
				outState := out.toState()

				if !nfaQueue.wasProcessed(out) {
					nfaQueue.enqueue(out)
					qPrime = append(qPrime, outState)

					for _, f := range n.F {
						if out[f] {
							fPrime = append(fPrime, outState)
							break
						}
					}
				}

				deltaPrime.Add(currentSetState.toState(), r, outState)
			}
		}
	}

	return NewDFA(qPrime, dfa_sigma, deltaPrime, q0.toState(), fPrime)
}

type nfaQueue struct {
	queue []setState
	reg   map[string]bool
}

func newNFAQueue() nfaQueue {
	return nfaQueue{
		queue: make([]setState, 0),
		reg:   make(map[string]bool),
	}
}

func (n nfaQueue) wasProcessed(s setState) bool {
	return n.reg[s.toState()]
}

func (n *nfaQueue) enqueue(s setState) {
	if !n.wasProcessed(s) {
		n.queue = append(n.queue, s)
	}
	n.reg[s.toState()] = true
}

func (n *nfaQueue) dequeue() setState {
	res := n.queue[0]
	n.queue = n.queue[1:]
	return res
}

func (n nfaQueue) done() bool {
	return len(n.queue) == 0

	// for _, q := range n.queue {
	// 	if !n.reg[q.toState()] {
	// 		return false
	// 	}
	// }
}

func (n *NFA) ToGrammar() *Grammar {
	grammar := &Grammar{
		Vn: n.Q,
		Vt: n.Sigma,
		P:  make(map[State][]string),
		S:  n.Q0[0],
	}

	for state, transitions := range n.Delta {
		for symbol, nextStates := range transitions {
			for nextState := range nextStates {
				grammar.P[state] = append(grammar.P[state], string(symbol)+nextState)
			}
		}
	}

	for _, finalState := range n.F {
		grammar.P[finalState] = append(grammar.P[finalState], "eps")
	}

	return grammar
}

// CreateBasicNFA creates a basic NFA that accepts a single character
func CreateBasicNFA(symbol byte, statePrefix string, counter *int) *NFA {
	start := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++
	accept := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++

	delta := make(DeltaNfa)
	delta.Add(start, symbol, NewSetState(accept))

	return NewNFA(
		[]State{start, accept},
		[]byte{symbol},
		delta,
		[]State{start},
		[]State{accept},
	)
}

// CreateEmptyNFA creates an NFA that accepts the empty string (epsilon)
func CreateEmptyNFA(statePrefix string, counter *int) *NFA {
	state := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++

	return NewNFA(
		[]State{state},
		[]byte{},
		make(DeltaNfa),
		[]State{state},
		[]State{state},
	)
}

// CreateWildcardNFA creates an NFA that accepts any single character from the alphabet
func CreateWildcardNFA(alphabet []byte, statePrefix string, counter *int) *NFA {
	start := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++
	accept := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++

	delta := make(DeltaNfa)

	// Create a transition for each character in the alphabet
	for _, symbol := range alphabet {
		if symbol != Epsilon {
			delta.Add(start, symbol, NewSetState(accept))
		}
	}

	return NewNFA(
		[]State{start, accept},
		alphabet,
		delta,
		[]State{start},
		[]State{accept},
	)
}

// ConcatenateNFAs connects two NFAs in sequence
func ConcatenateNFAs(first, second *NFA) *NFA {
	delta := make(DeltaNfa)

	for state, transitions := range first.Delta {
		for symbol, states := range transitions {
			delta.Add(state, symbol, states)
		}
	}

	for state, transitions := range second.Delta {
		for symbol, states := range transitions {
			delta.Add(state, symbol, states)
		}
	}

	for _, acceptState := range first.F {
		for _, startState := range second.Q0 {
			delta.Add(acceptState, Epsilon, NewSetState(startState))
		}
	}

	sigma := make([]byte, 0)
	symbolSet := make(map[byte]bool)

	for _, symbol := range first.Sigma {
		if !symbolSet[symbol] {
			sigma = append(sigma, symbol)
			symbolSet[symbol] = true
		}
	}

	for _, symbol := range second.Sigma {
		if !symbolSet[symbol] {
			sigma = append(sigma, symbol)
			symbolSet[symbol] = true
		}
	}

	if !symbolSet[Epsilon] {
		sigma = append(sigma, Epsilon)
	}

	states := append([]State{}, first.Q...)
	for _, state := range second.Q {
		states = append(states, state)
	}

	return NewNFA(
		states,
		sigma,
		delta,
		first.Q0,
		second.F,
	)
}

// UnionNFAs creates a new NFA that accepts either of two NFAs
func UnionNFAs(first, second *NFA, statePrefix string, counter *int) *NFA {
	start := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++
	accept := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++

	delta := make(DeltaNfa)

	for state, transitions := range first.Delta {
		for symbol, states := range transitions {
			delta.Add(state, symbol, states)
		}
	}

	for state, transitions := range second.Delta {
		for symbol, states := range transitions {
			delta.Add(state, symbol, states)
		}
	}

	for _, startState := range first.Q0 {
		delta.Add(start, Epsilon, NewSetState(startState))
	}

	for _, startState := range second.Q0 {
		delta.Add(start, Epsilon, NewSetState(startState))
	}

	for _, finalState := range first.F {
		delta.Add(finalState, Epsilon, NewSetState(accept))
	}

	for _, finalState := range second.F {
		delta.Add(finalState, Epsilon, NewSetState(accept))
	}

	sigma := make([]byte, 0)
	symbolSet := make(map[byte]bool)

	for _, symbol := range first.Sigma {
		if !symbolSet[symbol] {
			sigma = append(sigma, symbol)
			symbolSet[symbol] = true
		}
	}

	for _, symbol := range second.Sigma {
		if !symbolSet[symbol] {
			sigma = append(sigma, symbol)
			symbolSet[symbol] = true
		}
	}

	if !symbolSet[Epsilon] {
		sigma = append(sigma, Epsilon)
	}

	states := append([]State{start, accept}, first.Q...)
	for _, state := range second.Q {
		states = append(states, state)
	}

	return NewNFA(
		states,
		sigma,
		delta,
		[]State{start},
		[]State{accept},
	)
}

// StarNFA creates a Kleene star NFA (zero or more repetitions)
func StarNFA(nfa *NFA, statePrefix string, counter *int) *NFA {
	start := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++
	accept := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++

	delta := make(DeltaNfa)

	for state, transitions := range nfa.Delta {
		for symbol, states := range transitions {
			delta.Add(state, symbol, states)
		}
	}

	delta.Add(start, Epsilon, NewSetState(accept))

	for _, startState := range nfa.Q0 {
		delta.Add(start, Epsilon, NewSetState(startState))
	}

	for _, finalState := range nfa.F {
		delta.Add(finalState, Epsilon, NewSetState(accept))

		for _, startState := range nfa.Q0 {
			delta.Add(finalState, Epsilon, NewSetState(startState))
		}
	}

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

// PlusNFA creates a plus NFA (one or more repetitions)
func PlusNFA(nfa *NFA, statePrefix string, counter *int) *NFA {
	firstCopy := &NFA{
		Q:     append([]State{}, nfa.Q...),
		Sigma: append([]byte{}, nfa.Sigma...),
		Delta: make(DeltaNfa),
		Q0:    append([]State{}, nfa.Q0...),
		F:     append([]State{}, nfa.F...),
	}

	for state, transitions := range nfa.Delta {
		for symbol, states := range transitions {
			firstCopy.Delta.Add(state, symbol, states)
		}
	}

	starVersion := StarNFA(nfa, statePrefix, counter)

	return ConcatenateNFAs(firstCopy, starVersion)
}

// QuestionNFA creates a question mark NFA (zero or one occurrences)
func QuestionNFA(nfa *NFA, statePrefix string, counter *int) *NFA {
	start := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++
	accept := fmt.Sprintf("%s%d", statePrefix, *counter)
	*counter++

	delta := make(DeltaNfa)

	for state, transitions := range nfa.Delta {
		for symbol, states := range transitions {
			delta.Add(state, symbol, states)
		}
	}

	delta.Add(start, Epsilon, NewSetState(accept))

	for _, startState := range nfa.Q0 {
		delta.Add(start, Epsilon, NewSetState(startState))
	}

	for _, finalState := range nfa.F {
		delta.Add(finalState, Epsilon, NewSetState(accept))
	}

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

// RepeatNFA creates an NFA that repeats exactly n times
func RepeatNFA(nfa *NFA, n int) *NFA {
	if n <= 0 {
		// For n <= 0, return an NFA that accepts the empty string
		counter := 0
		return CreateEmptyNFA("q", &counter)
	}

	// Special case for n=1, just return a copy
	if n == 1 {
		return deepCopyNFA(nfa)
	}

	// For exact repetition, concatenate n copies
	stateCounter := 0

	// Create a new start and final state
	start := fmt.Sprintf("repeat_start_%d", stateCounter)
	stateCounter++
	final := fmt.Sprintf("repeat_final_%d", stateCounter)
	stateCounter++

	copies := make([]*NFA, n)
	for i := 0; i < n; i++ {
		copy := deepCopyNFA(nfa)

		// Rename the states to make them unique
		stateMap := make(map[State]State)
		for _, state := range copy.Q {
			// Only rename if not already renamed (to handle multiple references)
			if !strings.Contains(state, "_copy") {
				stateMap[state] = fmt.Sprintf("%s_copy%d_%d", state, i, stateCounter)
				stateCounter++
			}
		}

		// Apply the renaming to all components
		for j, state := range copy.Q {
			if newState, ok := stateMap[state]; ok {
				copy.Q[j] = newState
			}
		}

		for j, state := range copy.Q0 {
			if newState, ok := stateMap[state]; ok {
				copy.Q0[j] = newState
			}
		}

		for j, state := range copy.F {
			if newState, ok := stateMap[state]; ok {
				copy.F[j] = newState
			}
		}

		// Update the delta function
		newDelta := make(DeltaNfa)
		for state, transitions := range copy.Delta {
			newState := state
			if mapped, ok := stateMap[state]; ok {
				newState = mapped
			}

			for symbol, targetStates := range transitions {
				newTargets := make(setState)
				for target := range targetStates {
					if mapped, ok := stateMap[target]; ok {
						newTargets[mapped] = true
					} else {
						newTargets[target] = true
					}
				}
				newDelta.Add(newState, symbol, newTargets)
			}
		}
		copy.Delta = newDelta

		copies[i] = copy
	}

	// Connect the copies with epsilon transitions
	delta := make(DeltaNfa)
	states := []State{start, final}

	// Add epsilon transition from start to first copy
	delta.Add(start, Epsilon, NewSetState(copies[0].Q0...))

	// Add transitions from each copy
	for i, copy := range copies {
		for state, transitions := range copy.Delta {
			for symbol, targetStates := range transitions {
				delta.Add(state, symbol, targetStates)
			}
		}

		states = append(states, copy.Q...)

		if i < n-1 {
			for _, finalState := range copy.F {
				delta.Add(finalState, Epsilon, NewSetState(copies[i+1].Q0...))
			}
		} else {
			for _, finalState := range copy.F {
				delta.Add(finalState, Epsilon, NewSetState(final))
			}
		}
	}

	sigma := make([]byte, 0)
	symbolSet := make(map[byte]bool)
	for _, copy := range copies {
		for _, symbol := range copy.Sigma {
			if !symbolSet[symbol] {
				sigma = append(sigma, symbol)
				symbolSet[symbol] = true
			}
		}
	}

	if !symbolSet[Epsilon] {
		sigma = append(sigma, Epsilon)
	}

	return NewNFA(
		states,
		sigma,
		delta,
		[]State{start},
		[]State{final},
	)
}

// Helper function to create a deep copy of an NFA
func deepCopyNFA(nfa *NFA) *NFA {
	q := make([]State, len(nfa.Q))
	sigma := make([]byte, len(nfa.Sigma))
	q0 := make([]State, len(nfa.Q0))
	f := make([]State, len(nfa.F))

	copy(q, nfa.Q)
	copy(sigma, nfa.Sigma)
	copy(q0, nfa.Q0)
	copy(f, nfa.F)

	// Deep copy the transition function
	delta := make(DeltaNfa)
	for state, transitions := range nfa.Delta {
		for symbol, states := range transitions {
			// Create a new set of states for the destination
			newStates := make(setState)
			for s := range states {
				newStates[s] = true
			}
			delta.Add(state, symbol, newStates)
		}
	}

	return &NFA{
		Q:     q,
		Sigma: sigma,
		Delta: delta,
		Q0:    q0,
		F:     f,
	}
}
