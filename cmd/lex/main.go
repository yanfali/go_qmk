package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	filename := os.Args[1]
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	_lexer := lex(filename, string(input))

	//fmt.Printf("test %s\n", test.name)
	tokens := []item{}
	for {
		token := _lexer.nextItem()

		fmt.Printf("%s\n", token)
		tokens = append(tokens, token)
		if token.typ == itemEOF {
			break
		}
	}

}
