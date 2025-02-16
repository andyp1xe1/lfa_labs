# Laboratory Work 1

## Theory and Definitions

### Alphabet

A finite nonempty set of symbols. By convention denoted as $\Sigma$. For
example:

$$ \Sigma = \{0, 1\} \text{ binary alphabet} $$

$$ \Sigma = \{ a, b, \dots, z \} \text{ the set of lowercase letters} $$

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

$$ P = \{ \alpha \to \beta \mid \alpha, \beta \in (V_{N} \cup V_{T})^*, \alpha \ne \epsilon \} $$

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
3.  $S = \{q_0\}$
4.  For production $P$:

- $P = \emptyset$
- For all values:

$$ \delta(q, a) = (q_1, q_2, \dots, q_m) $$

we have

$$ P = P\cup \{ q \rightarrow aq_i \mid  i = 1\dots m \} $$

- for the values with

$$ F \cap \{ q_1, q_2, \dots, q_m\} \neq \emptyset $$

we have

$$ P = P\cup \{q \rightarrow a\} $$

## Objectives

## Implementation Description

## Results & Conclusions

## References
