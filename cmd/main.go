package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "a command is required (decrypt, decryptFile, encrypt, encryptFile, recrypt, recryptFile)\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "decrypt":
		decryptFlags.Parse(os.Args[2:])
		decrypt()
	case "decryptFile":
		decryptFileFlags.Parse(os.Args[2:])
		decryptFile()
	case "encrypt":
		encryptFlags.Parse(os.Args[2:])
		encrypt()
	case "encryptFile":
		encryptFileFlags.Parse(os.Args[2:])
		encryptFile()
	case "recrypt":
		recryptFlags.Parse(os.Args[2:])
		recrypt()
	case "recryptFile":
		recryptFileFlags.Parse(os.Args[2:])
		recryptFile()
	default:
		fmt.Fprintf(os.Stderr, "a command is required (decrypt, decryptFile, encrypt, encryptFile, recrypt, recryptFile)\n")
		os.Exit(2)
	}
}

func readPassword(prompt string, flags *flag.FlagSet, exitCode int) string {
	fmt.Print(prompt)
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "unable to read password\n")
		flags.Usage()
		os.Exit(exitCode)
	}
	pass := string(bytes)
	if pass == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "no password was provided\n")
		flags.Usage()
		os.Exit(exitCode)
	}
	return pass
}
