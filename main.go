package main

import (
	"bufio"
	"bytes"
	"notetest/notes"
	"os"

	_ "embed"

	"golang.org/x/crypto/argon2"
)

func main() {
	//go:embed usage.txt
	var usage string

	args := os.Args

	if len(args) != 2 {
		// too little args
		_, _ = os.Stdout.WriteString("Too little or too many arguments.\n")
		_, _ = os.Stdout.WriteString(usage)
		os.Exit(2)
	}

	filePath := args[1]

	_, _ = os.Stdout.WriteString("Enter password: ")

	r := bufio.NewReader(os.Stdin)
	buf := &bytes.Buffer{}

	for {
		b, err := r.ReadByte()
		if err != nil {
			panic(err)
		}

		if b != '\n' {
			buf.WriteByte(b)
		} else {
			break
		}
	}

	key := argon2.IDKey(buf.Bytes(), make([]byte, 32), 1, 64*1024, 2, 32)

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	nm, err := notes.UnlockDB(file, key)
	if err != nil {
		panic(err)
	}

}
