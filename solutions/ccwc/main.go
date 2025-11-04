package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"unicode/utf8"
)

func main() {
	countBytes := flag.Bool("c", false, "Count bytes")
	countLines := flag.Bool("l", false, "Count lines")
	countWords := flag.Bool("w", false, "Count words")
	countChars := flag.Bool("m", false, "Count characters")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: ccwc [options] <filename>")
		os.Exit(1)
	}

	filename := args[0]

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if *countBytes {
		byteCount := len(data)
		fmt.Printf("%d %s\n", byteCount, filename)
	}

	if *countLines {
		lineCount := countLinesInData(data)
		fmt.Printf("%d %s\n", lineCount, filename)
	}

	if *countWords {
		wordCount := countWordsInData(data)
		fmt.Printf("%d %s\n", wordCount, filename)
	}

	if *countChars {
		charCount := utf8.RuneCount(data)
		fmt.Printf("%d %s\n", charCount, filename)
	}
}

func countLinesInData(data []byte) int {
	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	// If the file does not end with a newline, count the last line
	if len(data) > 0 && data[len(data)-1] != '\n' {
		count++
	}
	return count
}

func countWordsInData(data []byte) int {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanWords)

	count := 0
	for scanner.Scan() {
		count++
	}

	return count
}
