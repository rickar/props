package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rickar/props"
)

var (
	encryptFileFlags  = flag.NewFlagSet("encryptFile", flag.ExitOnError)
	encryptFilePath   = encryptFileFlags.String("path", "", "properties `file` to encrypt")
	encryptFilePass   = encryptFileFlags.String("password", "", "`password` to encrypt the values")
	encryptFileAlg    = encryptFileFlags.String("alg", props.EncryptDefault, "encryption `algorithm` to use (see props.Encrypt*)")
	encryptFileOutput = encryptFileFlags.String("output", "", "output `file` to write results (default is input file)")
)

func init() {
	encryptFileFlags.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "encryptFile: encrypt plaintext values in a property file\n")
		encryptFileFlags.PrintDefaults()
	}
}

func encryptFile() {
	if *encryptFilePath == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "the path parameter is required\n")
		encryptFileFlags.Usage()
		os.Exit(400)
	} else {
		stat, err := os.Stat(*encryptFilePath)
		if err != nil || stat.IsDir() {
			fmt.Fprintf(flag.CommandLine.Output(), "the path parameter must be an existing, readable file\n")
			encryptFileFlags.Usage()
			os.Exit(401)
		}
	}
	if *encryptFilePass == "" {
		*encryptFilePass = readPassword("Password:", encryptFileFlags, 402)
	}
	if *encryptFileAlg != props.EncryptAESGCM {
		fmt.Fprintf(flag.CommandLine.Output(), "the alg parameter must be an encryption algorithm id from props\n")
		encryptFileFlags.Usage()
		os.Exit(403)
	}
	if *encryptFileOutput == "" {
		*encryptFileOutput = *encryptFilePath
	}

	f, err := os.Open(*encryptFilePath)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "unable to read property file: %v\n", err)
		os.Exit(404)
	}
	defer f.Close()

	found := 0
	var result bytes.Buffer
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, props.EncryptNone) {
			found++
			i := strings.Index(line, props.EncryptNone)
			val := line[i+len(props.EncryptNone):]
			line := line[:i]
			enc, err := props.Encrypt(*encryptFileAlg, *encryptFilePass, val)
			if err != nil {
				fmt.Fprintf(flag.CommandLine.Output(), "unable to encrypt property: %v\n", err)
				os.Exit(405)
			}
			result.WriteString(line)
			result.WriteString(enc)
			result.WriteRune('\n')
		} else {
			result.WriteString(line)
			result.WriteRune('\n')
		}
	}
	f.Close()
	err = os.WriteFile(*encryptFileOutput, result.Bytes(), 0o644)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "unable to write output: %v\n", err)
		os.Exit(406)
	}
	fmt.Printf("%d properties encrypted\n", found)
}
