package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		fmt.Println("EDITOR is not set")
		os.Exit(1)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	filename := filepath.Join(home, ".j.md")

	var section string
	var toAdd []string

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-s":
			i++
			if i >= len(os.Args) {
				fmt.Println("Error: expected message argument", os.Args[i])
				os.Exit(1)
			}
			section = os.Args[i]
		case "-m":
			i++
			if i >= len(os.Args) {
				fmt.Println("Error: expected message argument", os.Args[i])
				os.Exit(1)
			}
			toAdd = append(toAdd, os.Args[i])
		case "-t":
			i++
			if i >= len(os.Args) {
				fmt.Println("Error: expected todo argument", os.Args[i])
				os.Exit(1)
			}
			toAdd = append(toAdd, "- [ ] "+os.Args[i])
		case "to", "todo":
			i++
			if i >= len(os.Args) {
				fmt.Println("Error: expected todo content", os.Args[i])
				os.Exit(1)
			}
			text := strings.Join(os.Args[i:], " ")
			toAdd = append(toAdd, "- [ ] "+text)
			i += len(os.Args[i:])
		case "note":
			i++
			if i >= len(os.Args) {
				fmt.Println("Error: expected todo content", os.Args[i])
				os.Exit(1)
			}
			text := strings.Join(os.Args[i:], " ")
			toAdd = append(toAdd, text)
			i += len(os.Args[i:])
		default:
			fmt.Println("Unknown arg", os.Args[i])
			os.Exit(1)
		}
	}

	if len(toAdd) > 0 {
		j, err := parseJournalFile(filename)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		// j.prependLines(toAdd)
		if section == "" {
			j.appendLinesToDateSection(time.Now(), toAdd)
		} else {
			j.appendLinesToSection(section, toAdd)
		}
		err = j.writeFile(filename)
	} else {
		err = openEditor(editor, filename)
	}
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func openEditor(editorCmd, filename string) error {
	c := exec.Command(editorCmd, filename)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Env = os.Environ()
	return c.Run()
}

type journal struct {
	lines   []string
	headers map[string]int
}

func dateHeader(d time.Time) string {
	return d.Format("# (Mon) Jan 2 2006")
}

func (j *journal) appendLinesToDateSection(d time.Time, lines []string) {
	j.appendLinesToSection(dateHeader(d), lines)
}

func (j *journal) appendLinesToSection(s string, lines []string) {
	dateHeaderLine, exists := j.headers[s]
	if !exists {
		newLines := append([]string{s}, lines...)
		j.prependLines(newLines)
	} else {
		nextHeaderLine := len(j.lines)
		for _, v := range j.headers {
			if v > dateHeaderLine && v < nextHeaderLine {
				nextHeaderLine = v
			}
		}
		j.prependLinesTo(nextHeaderLine, lines)
	}
}

func (j *journal) prependLines(lines []string) {
	j.prependLinesTo(0, lines)
}
func (j *journal) prependLinesTo(n int, lines []string) {
	j.lines = append(j.lines[:n], append(lines, j.lines[n:]...)...)
	for k, v := range j.headers {
		if v >= n {
			j.headers[k] = v + len(j.lines)
		}
	}
}

func (j journal) writeFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, l := range j.lines {
		_, err = f.WriteString(l + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

var headerRegex = regexp.MustCompile("^#+ +.*")

func parseJournalFile(filename string) (journal, error) {
	var j journal
	j.headers = make(map[string]int)

	f, err := os.Open(filename)
	if err != nil {
		return j, err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	var line string
	line, err = r.ReadString('\n')

	for err == nil {
		line = strings.TrimRight(line, " \t\n\r")
		if headerRegex.MatchString(line) {
			j.headers[line] = len(j.lines)
		}
		j.lines = append(j.lines, line)
		line, err = r.ReadString('\n')
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return j, nil
}
