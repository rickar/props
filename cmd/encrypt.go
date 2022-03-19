package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rickar/props"
)

var (
	encryptFlags = flag.NewFlagSet("encrypt", flag.ExitOnError)
	encryptValue = encryptFlags.String("value", "", "`plaintext` value to encrypt")
	encryptPass  = encryptFlags.String("password", "", "`password` to encrypt the value")
	encryptAlg   = encryptFlags.String("alg", props.EncryptDefault, "encryption `algorithm` to use (see props.Encrypt*)")
)

func init() {
	encryptFlags.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "encrypt: encrypt a value for use in a property file\n")
		encryptFlags.PrintDefaults()
	}
}

func encrypt() {
	if *encryptValue == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "the value parameter is required\n")
		encryptFlags.Usage()
		os.Exit(300)
	}
	if *encryptPass == "" {
		*encryptPass = readPassword("Password:", encryptFlags, 301)
	}
	if *encryptAlg != props.EncryptAESGCM {
		fmt.Fprintf(flag.CommandLine.Output(), "the alg parameter must be an encryption algorithm id from props\n")
		encryptFlags.Usage()
		os.Exit(302)
	}

	switch len(*encryptPass) {
	case 16, 24, 32:
	default:
		fmt.Fprintf(flag.CommandLine.Output(), "the password parameter must be 16, 24, or 32 bytes\n")
		encryptFlags.Usage()
		os.Exit(303)
	}

	enc, err := props.Encrypt(*encryptAlg, *encryptPass, *encryptValue)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "encrypt error: %v\n", err)
		os.Exit(304)
	}
	fmt.Println(enc)
}
