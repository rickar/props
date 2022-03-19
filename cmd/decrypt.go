package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rickar/props"
)

var (
	decryptFlags = flag.NewFlagSet("decrypt", flag.ExitOnError)
	decryptValue = decryptFlags.String("value", "", "`encrypted` value to decrypt (including algorithm prefix)")
	decryptPass  = decryptFlags.String("password", "", "`password` to decrypt the value")
)

func init() {
	decryptFlags.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "decrypt: decrypt an encrypted property value\n")
		decryptFlags.PrintDefaults()
	}
}

func decrypt() {
	if *decryptValue == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "the value parameter is required\n")
		decryptFlags.Usage()
		os.Exit(100)
	}
	if *decryptPass == "" {
		*decryptPass = readPassword("Password:", decryptFlags, 100)
	}

	dec, err := props.Decrypt(*decryptPass, *decryptValue)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "decrypt error: %v\n", err)
		os.Exit(101)
	}
	fmt.Println(dec)
}
