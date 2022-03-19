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
	recryptFileFlags   = flag.NewFlagSet("recryptFile", flag.ExitOnError)
	recryptFilePath    = recryptFileFlags.String("path", "", "properties `file` to re-encrypt")
	recryptFilePass    = recryptFileFlags.String("newpass", "", "new `password` to re-encrypt the values")
	recryptFileOldPass = recryptFileFlags.String("oldpass", "", "old `password` to decrypt the values")
	recryptFileAlg     = recryptFileFlags.String("alg", props.EncryptDefault, "encryption `algorithm` to use (see props.Encrypt*)")
	recryptFileOutput  = recryptFileFlags.String("output", "", "output `file` to write results (default is input file)")
)

func init() {
	recryptFileFlags.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "recryptFile: re-encrypt values in a property file\n")
		recryptFileFlags.PrintDefaults()
	}
}

func recryptFile() {
	if *recryptFilePath == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "the path parameter is required\n")
		recryptFileFlags.Usage()
		os.Exit(600)
	} else {
		stat, err := os.Stat(*recryptFilePath)
		if err != nil || stat.IsDir() {
			fmt.Fprintf(flag.CommandLine.Output(), "the path parameter must be an existing, readable file\n")
			recryptFileFlags.Usage()
			os.Exit(601)
		}
	}
	if *recryptFileAlg != props.EncryptAESGCM {
		fmt.Fprintf(flag.CommandLine.Output(), "the alg parameter must be an encryption algorithm id from props\n")
		recryptFileFlags.Usage()
		os.Exit(602)
	}
	if *recryptFileOldPass == "" {
		*recryptFileOldPass = readPassword("Old Password:", recryptFileFlags, 603)
	}
	if *recryptFilePass == "" {
		*recryptFilePass = readPassword("New Password:", recryptFileFlags, 604)
	}
	if *recryptFileOutput == "" {
		*recryptFileOutput = *recryptFilePath
	}

	f, err := os.Open(*recryptFilePath)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "unable to read property file: %v\n", err)
		os.Exit(605)
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
			dec, err := props.Decrypt(*recryptFileOldPass, val)
			if err != nil {
				fmt.Fprintf(flag.CommandLine.Output(), "unable to decrypt property: %v\n", err)
				os.Exit(100)
			}

			enc, err := props.Encrypt(*recryptFileAlg, *recryptFilePass, dec)
			if err != nil {
				fmt.Fprintf(flag.CommandLine.Output(), "unable to encrypt property: %v\n", err)
				os.Exit(606)
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
	err = os.WriteFile(*recryptFileOutput, result.Bytes(), 0o644)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "unable to write output: %v\n", err)
		os.Exit(607)
	}
	fmt.Printf("%d properties re-encrypted\n", found)
}
