#!/bin/sh
go run . &&
  dot -Tpng dfa.dot -o dfa.png &&
  nsxiv dfa.png
