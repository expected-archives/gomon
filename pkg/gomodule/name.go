package gomodule

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path"
)

var ErrUnableToGetGoMod = errors.New("unable to get go gomodule")

func GetName() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", ErrUnableToGetGoMod
	}

	goModPath := path.Join(dir, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return "", ErrUnableToGetGoMod
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := getFullLine(reader)
		if err != nil && err != io.EOF {
			return "", ErrUnableToGetGoMod
		}

		if bytes.Contains(line, []byte("gomodule")) {
			return getModuleName(line), nil
		}

		if err != nil && err == io.EOF {
			return "", ErrUnableToGetGoMod
		}
	}
}

func getModuleName(line []byte) string {
	lineSplit := bytes.SplitN(line, []byte("gomodule"), 2)
	if len(lineSplit) != 2 {
		return ""
	}

	moduleName := string(bytes.TrimSpace(bytes.TrimSpace(lineSplit[1])))

	return moduleName
}

func getFullLine(reader *bufio.Reader) ([]byte, error) {
	var (
		line     []byte
		tmpLine  []byte
		isPrefix bool
		err      error
	)

	for {
		tmpLine, isPrefix, err = reader.ReadLine()
		if err == io.EOF {
			return line, io.EOF
		} else if err != nil {
			return nil, err
		}

		line = append(line, tmpLine...)
		if !isPrefix {
			break
		}
	}

	return line, nil
}
