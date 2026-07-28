.PHONY: compile

compile:
	mkdir -p result
	pdflatex -output-directory=./result ./main.tex
	bibtex ./result/main
	pdflatex -output-directory=./result ./main.tex
	pdflatex -output-directory=./result ./main.tex