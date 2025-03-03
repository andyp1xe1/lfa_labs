#!/bin/sh
go run . &&
  dot -Tpng dfa.dot -o dfa.png &&
  dot -Tpng nfa.dot -o nfa.png &&
  nsxiv dfa.png
  nsxiv nfa.png
