package main_test

import (
	_ "embed"
	"strings"
	"testing"
)

var (
	//go:embed token
	tokenList string
)

func TestToken(t *testing.T) {
	tokens := strings.Split(strings.Trim(tokenList, "\n"), "\n")
	t.Log(len(tokens))
	for i := 0; i < len(tokens); i++ {
		token := strings.Split(tokens[i], ",")
		t.Log("network", token[0], "address", token[1])
	}
}
