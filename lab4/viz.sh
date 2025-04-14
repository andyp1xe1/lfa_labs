#!/bin/sh
go run .

# Generate PNG from DOT files
for file in regex_nfa_*.dot; do
  png_file="${file%.dot}.png"
  dot -Tpng "$file" -o "$png_file" && echo "Generated $png_file"
done

# Try to display images with available viewer
if command -v nsxiv >/dev/null 2>&1; then
  nsxiv regex_nfa_*.png
elif command -v xdg-open >/dev/null 2>&1; then
  xdg-open regex_nfa_*.png
fi
