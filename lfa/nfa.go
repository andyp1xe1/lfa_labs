package lfa

import (
	"sort"
	"strings"
)

type DeltaNfa map[State]map[byte]setState

func (d DeltaNfa) Add(in State, r byte, out setState) {
	if _, ok := d[in]; !ok {
		d[in] = make(map[byte]setState)
	}
	d[in][r] = out
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

func (n *NFA) ToDFA() *DFA {
	nfaQueue := newNFAQueue()
	deltaPrime := make(DeltaDFA)

	qPrime := make([]State, 0)
	fPrime := make([]State, 0)

	q0 := make(setState)
	for _, q := range n.Q0 {
		q0.Add(q)
	}

	nfaQueue.enqueue(q0)
	qPrime = append(qPrime, q0.toState())

	for !nfaQueue.done() {
		currentSetState := nfaQueue.dequeue()
		//fmt.Println("curr: ", currentSetState)
		//time.Sleep(200 * time.Millisecond)

		for _, f := range n.F {
			if currentSetState[f] {
				fPrime = append(fPrime, currentSetState.toState())
				break
			}
		}
		//fmt.Println("Appended to final: ", currentSetState.toState())

		for _, r := range n.Sigma {
			out := make(setState)
			for state := range currentSetState {
				//fmt.Printf("\tcurr2: (%v, %s); ", state, string(r))
				if nextStates := n.Delta.Lookup(state, r); nextStates != nil {
					//fmt.Println("adding: ", nextStates)
					out.Union(nextStates)
				} else {
					//fmt.Println()
				}
			}

			if len(out) > 0 {
				if !nfaQueue.wasProcessed(out) {
					//fmt.Println("enqueuing new state: ", out)
					nfaQueue.enqueue(out)
					qPrime = append(qPrime, out.toState())
				}

				deltaPrime.Add(currentSetState.toState(), r, out.toState())
				//fmt.Printf("Registered: %v %v %v\n", currentSetState.toState(), string(r), out.toState())
			}
		}
	}

	return NewDFA(qPrime, n.Sigma, deltaPrime, q0.toState(), fPrime)
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
