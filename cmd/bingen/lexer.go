package main

import (
	"bufio"
	"io"
	"strings"
)

type token struct {
	typ  MapType
	tok  string
	line int
	col  int
}

func lex(r io.Reader) ([]token, error) {
	var (
		tokens []token
	)
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		initLen := len(line)
		line = strings.TrimSpace(line)
		col := initLen - len(line) + 1
		words := strings.Split(line, " ")
		for i := 0; i < len(words); i++ {
			if i > 0 {
				col++
			}
			if len(words[i]) == 0 {
				continue
			}
			if strings.HasPrefix(words[i], "//") {
				// Line comment
				break
			}
			tokens = append(tokens, token{
				typ:  matchType(words[i]),
				tok:  words[i],
				line: lineNum,
				col:  col,
			})
			col += len(words[i])
		}
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			return nil, err
		}
	}
	return tokens, nil
}

type iterator = func(offset ...int) (*token, bool)

func tokenIterator(tokens []token) iterator {
	if len(tokens) == 0 {
		return func(offset ...int) (*token, bool) {
			return nil, false
		}
	}
	i := -1
	return func(offset ...int) (*token, bool) {
		if len(offset) > 0 {
			i += offset[0]
			return nil, false
		}
		i++
		if i < 0 {
			i = 0
		}
		if i >= len(tokens) {
			i--
			return nil, false
		}
		return &tokens[i], true
	}
}
