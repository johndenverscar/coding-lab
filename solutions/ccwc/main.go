package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

type Counts struct {
	Lines int
	Words int
	Bytes int
	Chars int
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == '\f' || b == '\v'
}

func getCounts(data []byte) Counts {
	counts := Counts{
		Bytes: len(data),
		Chars: utf8.RuneCount(data),
	}

	inWord := false
	for _, b := range data {
		if b == '\n' {
			counts.Lines++
		}

		if isWhitespace(b) {
			if inWord {
				counts.Words++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	if inWord {
		counts.Words++
	}

	return counts
}

func main() {
	countBytes := flag.Bool("c", false, "Count bytes")
	countLines := flag.Bool("l", false, "Count lines")
	countWords := flag.Bool("w", false, "Count words")
	countChars := flag.Bool("m", false, "Count characters")

	flag.Parse()

	args := flag.Args()

	var data []byte
	var err error
	var filename string

	if len(args) == 0 {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		filename = ""
	} else {
		filename = args[0]
		data, err = os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	}

	counts := getCounts(data)

	noFlags := !*countBytes && !*countLines && !*countWords && !*countChars

	if noFlags {
		fmt.Printf("%d %d %d %d %s\n", counts.Lines, counts.Words, counts.Bytes, counts.Chars, filename)
		return
	}

	var results []int

	if *countBytes {
		results = append(results, counts.Bytes)
	}

	if *countLines {
		results = append(results, counts.Lines)
	}

	if *countWords {
		results = append(results, counts.Words)
	}

	if *countChars {
		results = append(results, counts.Chars)
	}

	for i, result := range results {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(result)
	}

	if filename != "" {
		fmt.Printf(" %s", filename)
	}
	fmt.Println()
}
