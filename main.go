package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// The message to log.
	msg := strings.Join(os.Args[1:], " ")

	// The current time.
	now := time.Now()

	// The output file name.
	fileName := now.Format("./2006_01/02.md")

	if err := createOrAppendToFile(fileName, func(lastLine string) []byte {
		t := now.Format("03:04 PM")
		body := fmt.Sprintf("%s - %s", t, msg)

		isNew := len(lastLine) == 0
		if isNew {
			header := now.Format("Mon, 02 Jan 2006")
			body = fmt.Sprintf("# %s\n\n%s", header, body)
		} else {
			lastTs, _, ok := strings.Cut(lastLine, " - ")
			if ok {
				t, err := time.Parse("03:04 PM", lastTs)
				if err == nil {
					t1 := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
					body = fmt.Sprintf("%s (took %s)", body, now.Sub(t1))
				}
			}
		}

		return []byte(fmt.Sprintf("%s\n", body))
	}); err != nil {
		panic(err)
	}
}

func createOrAppendToFile(name string, fn func(lastLine string) []byte) error {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_APPEND, 0644)
	if errors.Is(err, os.ErrNotExist) {
		dir := filepath.Dir(name)

		if err := os.MkdirAll(dir, 0700); err != nil && !os.IsExist(err) {
			return err
		} // Create your file

		return os.WriteFile(name, fn(""), 0644)
	}
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(fn(seekLastLine(f)))
	return err
}

func seekLastLine(f *os.File) string {
	line := ""
	var cursor int64 = 0
	stat, _ := f.Stat()
	filesize := stat.Size()
	for {
		cursor -= 1
		f.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		f.Read(char)

		if cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line) // there is more efficient way

		if cursor == -filesize { // stop if we are at the begining
			break
		}
	}

	return strings.TrimSpace(line)
}
