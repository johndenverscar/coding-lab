package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	countBytes := flag.Bool("c", false, "Count bytes")

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
}
