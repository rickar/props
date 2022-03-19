package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rickar/props"
)

var (
	recryptFlags   = flag.NewFlagSet("recrypt", flag.ExitOnError)
	recryptValue   = recryptFlags.String("value", "", "`encrypted` value to re-encrypt")
	recryptPass    = recryptFlags.String("newpass", "", "new `password` to re-encrypt the value")
	recryptOldPass = recryptFlags.String("oldpass", "", "old `password` to decrypt the value")
	recryptAlg     = recryptFlags.String("alg", props.EncryptDefault, "encryption `algorithm` to use (see props.Encrypt*)")
)

func init() {
	recryptFlags.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "recrypt: re-encrypt a property value with a new password\n")
		recryptFlags.PrintDefaults()
	}
}

func recrypt() {
	if *recryptValue == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "the value parameter is required\n")
		recryptFlags.Usage()
		os.Exit(500)
	}
	if *recryptOldPass == "" {
		*recryptOldPass = readPassword("Old Password:", recryptFlags, 501)
	}
	if *recryptPass == "" {
		*recryptPass = readPassword("New Password:", recryptFlags, 502)
	}
	if *recryptAlg != props.EncryptAESGCM {
		fmt.Fprintf(flag.CommandLine.Output(), "the alg parameter must be an encryption algorithm id from props\n")
		recryptFlags.Usage()
		os.Exit(503)
	}

	switch len(*recryptPass) {
	case 16, 24, 32:
	default:
		fmt.Fprintf(flag.CommandLine.Output(), "the newpass parameter must be 16, 24, or 32 bytes\n")
		recryptFlags.Usage()
		os.Exit(504)
	}

	p := props.NewProperties()
	p.Set("val", *recryptValue)
	c := props.Configuration{Props: p}
	dec, err := c.Decrypt(*recryptOldPass, "val", "")
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "decrypt error: %v\n", err)
		os.Exit(505)
	}

	enc, err := props.Encrypt(*recryptAlg, *recryptPass, dec)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "encrypt error: %v\n", err)
		os.Exit(506)
	}
	fmt.Println(enc)
}
