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
	decryptFileFlags  = flag.NewFlagSet("decryptFile", flag.ExitOnError)
	decryptFilePath   = decryptFileFlags.String("path", "", "properties `file` to decrypt")
	decryptFilePass   = decryptFileFlags.String("password", "", "`password` to decrypt the values")
	decryptFileOutput = decryptFileFlags.String("output", "", "output `file` to write results (default is input file)")
)

func init() {
	decryptFileFlags.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "decryptFile: decrypt encrypted values in a property file\n")
		decryptFileFlags.PrintDefaults()
	}
}

func decryptFile() {
	if *decryptFilePath == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "the path parameter is required\n")
		decryptFileFlags.Usage()
		os.Exit(200)
	} else {
		stat, err := os.Stat(*decryptFilePath)
		if err != nil || stat.IsDir() {
			fmt.Fprintf(flag.CommandLine.Output(), "the path parameter must be an existing, readable file\n")
			decryptFileFlags.Usage()
			os.Exit(201)
		}
	}
	if *decryptFilePass == "" {
		*decryptFilePass = readPassword("Password:", decryptFileFlags, 202)
	}
	if *decryptFileOutput == "" {
		*decryptFileOutput = *decryptFilePath
	}

	f, err := os.Open(*decryptFilePath)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "unable to read property file: %v\n", err)
		os.Exit(203)
	}
	defer f.Close()

	found := 0
	var result bytes.Buffer
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, props.EncryptAESGCM) {
			found++
			i := strings.Index(line, props.EncryptAESGCM)
			val := line[i:]
			line := line[:i]
			enc, err := props.Decrypt(*decryptFilePass, val)
			if err != nil {
				fmt.Fprintf(flag.CommandLine.Output(), "unable to decrypt property: %v\n", err)
				os.Exit(204)
			}
			result.WriteString(line)
			result.WriteString(props.EncryptNone)
			result.WriteString(enc)
			result.WriteRune('\n')
		} else {
			result.WriteString(line)
			result.WriteRune('\n')
		}
	}
	f.Close()
	err = os.WriteFile(*decryptFileOutput, result.Bytes(), 0o644)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "unable to write output: %v\n", err)
		os.Exit(205)
	}
	fmt.Printf("%d properties decrypted\n", found)
}
