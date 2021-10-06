package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"notetest/notes"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	_ "embed"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/ssh/terminal"
)

func readLine(r *bufio.Reader) []byte {
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

	return buf.Bytes()
}

//go:embed usage.txt
var usage string

func main() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "/usr/bin/nvim"
	}

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
	// b := readLine(r)
	b, err := terminal.ReadPassword(0)
	if err != nil {
		panic(err)
	}
	_, _ = os.Stdout.WriteString("\n")

	key := argon2.IDKey(b, make([]byte, 32), 1, 64*1024, 2, 32)

	var nm *notes.Notes

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening DB file: %s\nCreating a new db\n", err.Error())
		nm = notes.NewEmptyDB()
		goto skipOpen
	}

	nm, err = notes.UnlockDB(file, key)
	if err != nil {
		panic(err)
	}

	if err := file.Close(); err != nil {
		panic(err)
	}

skipOpen:

	const tmpFilePath = "/dev/shm/notetesttmpfile.txt"

	defer os.Remove(tmpFilePath)

	for {
		_, _ = os.Stdout.WriteString("Notes:\n")

		noteList := nm.ListNotes()
		sort.Strings(noteList)
		for i, noteName := range noteList {
			fmt.Printf("%d | %s\n", i, noteName)
		}

		noteN := 0
		newNoteName := ""

		for {
			_, _ = os.Stdout.WriteString("Enter number (type 'c' to create new): ")
			line := readLine(r)
			_, _ = os.Stdout.WriteString("\n")

			strLine := string(line)

			if strLine == "c" {
				_, _ = os.Stdout.WriteString("Enter new note name: ")
				newNoteName = string(readLine(r))
				_, _ = os.Stdout.WriteString("\n")
				break
			}

			n, err := strconv.ParseInt(strLine, 10, 64)
			if err != nil {
				fmt.Printf("Error parsing int: %s\n", err.Error())
				continue
			}

			if n < 0 || n >= int64(len(noteList)) {
				_, _ = os.Stdout.WriteString("Number is either below zero or too high. Try again.\n")
				continue
			}

			noteN = int(n)
			break
		}

		noteName := newNoteName

		if newNoteName == "" {
			noteName = noteList[noteN]
		}

		noteContent, found := nm.ViewNote(noteName)
		if !found {
			noteContent = ""
		}

		tmpFile, err := os.Create(tmpFilePath)
		if err != nil {
			panic(err)
		}

		if _, err := tmpFile.WriteString(noteContent); err != nil {
			panic(err)
		}

		if err := tmpFile.Close(); err != nil {
			panic(err)
		}

		// cmd := exec.Cmd{
		// 	Path:   editor,
		// 	Args:   []string{tmpFilePath},
		// 	Stdin:  os.Stdin,
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stderr,
		// }

		cmd := exec.Command(editor, tmpFilePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			panic(err)
		}

		if err := cmd.Wait(); err != nil {
			panic(err)
		}

		tmpFileR, err := os.Open(tmpFilePath)
		if err != nil {
			panic(err)
		}

		sb := &strings.Builder{}

		if _, err := io.Copy(sb, tmpFileR); err != nil {
			panic(err)
		}

		if err := tmpFileR.Close(); err != nil {
			panic(err)
		}

		nm.UpdateNote(notes.NotesUpdate{
			Name:    noteName,
			Content: sb.String(),
			Delete:  false,
		})

		dbFile, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}

		if err := nm.SaveDB(dbFile, key); err != nil {
			panic(err)
		}

		if err := dbFile.Close(); err != nil {
			panic(err)
		}
	}
}
