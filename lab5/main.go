package main

import (
	"fmt"
	"lfa_labs/lfa"
	"log"
	"os"
	"path/filepath"
)

// makeTestGrammar creates a test grammar as defined in the test file
func makeTestGrammar() *lfa.GrammarV2 {
	g := lfa.NewGrammarV2("S", "ε",
		[]string{"S", "A", "B", "C", "D"},
		[]string{"d", "a", "b"})
	g.AddProduction("S", []string{"d", "B"})
	g.AddProduction("S", []string{"A"})
	g.AddProduction("A", []string{"d"})
	g.AddProduction("A", []string{"d", "S"})
	g.AddProduction("A", []string{"a", "B", "d", "B"})
	g.AddProduction("B", []string{"a"})
	g.AddProduction("B", []string{"a", "S"})
	g.AddProduction("B", []string{"A", "C"})
	g.AddProduction("D", []string{"A", "B"})
	g.AddProduction("C", []string{"b", "C"})
	g.AddProduction("C", []string{"ε"})
	return g
}

func main() {
	// Create output directory if it doesn't exist
	outputDir := "grammar_viz"
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Original grammar
	original := makeTestGrammar()
	fmt.Println("Original Grammar:")
	fmt.Println(original)

	// Generate visualization for original grammar
	err = original.ToDOT(filepath.Join(outputDir, "1_original.dot"))
	if err != nil {
		fmt.Printf("Error creating DOT file for original grammar: %v\n", err)
		return
	}

	// Step 1: Eliminate Epsilon
	g := makeTestGrammar()
	g.EliminateEpsilon()
	fmt.Println("\nAfter Epsilon Elimination:")
	fmt.Println(g)
	err = g.ToDOT(filepath.Join(outputDir, "2_no_epsilon.dot"))
	if err != nil {
		fmt.Printf("Error creating DOT file: %v\n", err)
		return
	}

	// Step 2: Eliminate Renaming
	g.EliminateRenaming()
	fmt.Println("\nAfter Renaming Elimination:")
	fmt.Println(g)
	err = g.ToDOT(filepath.Join(outputDir, "3_no_renaming.dot"))
	if err != nil {
		fmt.Printf("Error creating DOT file: %v\n", err)
		return
	}

	// Step 3: Eliminate Inaccessible Symbols
	g.EliminateInaccessibleSymbols()
	fmt.Println("\nAfter Inaccessible Symbols Elimination:")
	fmt.Println(g)
	err = g.ToDOT(filepath.Join(outputDir, "4_no_inaccessible.dot"))
	if err != nil {
		fmt.Printf("Error creating DOT file: %v\n", err)
		return
	}

	// Step 4: Eliminate Non-Productive Symbols
	g.EliminateNonProductiveSymbols()
	fmt.Println("\nAfter Non-Productive Symbols Elimination:")
	fmt.Println(g)
	err = g.ToDOT(filepath.Join(outputDir, "5_no_nonproductive.dot"))
	if err != nil {
		fmt.Printf("Error creating DOT file: %v\n", err)
		return
	}

	// Step 5: Convert to Chomsky Normal Form
	g.ConvertToCNF()
	fmt.Println("\nAfter CNF Conversion:")
	fmt.Println(g)
	err = g.ToDOT(filepath.Join(outputDir, "6_cnf.dot"))
	if err != nil {
		fmt.Printf("Error creating DOT file: %v\n", err)
		return
	}

	// Full normalization in one step
	fmt.Println("\nDOT files generated in 'grammar_viz' directory.")

	// Generate a batch script to create all PNGs at once
	generateScript(outputDir)
}

// generateScript creates a shell script to generate all PNG files
func generateScript(outputDir string) {
	steps := []string{
		"original", "no_epsilon", "no_renaming", "no_inaccessible", "no_nonproductive", "cnf"}

	shFile, err := os.Create(filepath.Join(outputDir, "generate_images.sh"))
	if err != nil {
		log.Fatal(err)
	}
	defer shFile.Close()

	fmt.Fprintln(shFile, "#!/bin/bash")
	fmt.Fprintln(shFile, "echo Generating PNG images...")
	for i, stepName := range steps {
		fmt.Fprintf(shFile, "dot -Tpng %d_%s.dot -o %d_%s.png\n", i+1, stepName, i+1, stepName)
	}
	fmt.Fprintln(shFile, "echo Done.")

	// Make shell script executable
	os.Chmod(filepath.Join(outputDir, "generate_images.sh"), 0755)

	fmt.Println("To create PNG images, run:")
	fmt.Println("cd grammar_viz")
	fmt.Println("./generate_images.sh")
}
