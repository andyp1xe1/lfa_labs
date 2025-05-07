#!/bin/bash
echo Generating PNG images...
dot -Tpng 1_original.dot -o 1_original.png
dot -Tpng 2_no_epsilon.dot -o 2_no_epsilon.png
dot -Tpng 3_no_renaming.dot -o 3_no_renaming.png
dot -Tpng 4_no_inaccessible.dot -o 4_no_inaccessible.png
dot -Tpng 5_no_nonproductive.dot -o 5_no_nonproductive.png
dot -Tpng 6_cnf.dot -o 6_cnf.png
echo Done.
