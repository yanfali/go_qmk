package main

import (
	"sync/atomic"
	"testing"
)

func TestLexer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []item
	}{
		{name: "a comment", input: "#comment", expected: []item{{typ: itemComment, val: "#comment"}, {typ: itemEOF, val: ""}}},
		{
			name:  "a multline",
			input: "#comment \\\nsome more comment",
			expected: []item{
				{typ: itemComment, val: "#comment \\\nsome more comment"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name:  "a multline with one space before next line",
			input: "#comment \\ \nsome more comment\n#a new comment",
			expected: []item{
				{typ: itemComment, val: "#comment \\ \nsome more comment\n"},
				{typ: itemComment, val: "#a new comment"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name:  "a multline with 3 spaces before next line",
			input: "#comment \\   \nsome more comment\n",
			expected: []item{
				{typ: itemComment, val: "#comment \\   \nsome more comment\n"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name: "a recursively expanded variable", input: "MCU = atmega32u4",
			expected: []item{
				{typ: itemAssignment, val: "MCU = atmega32u4"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name:  "a variable and a comment",
			input: "MCU = atmega32u4 # comment",
			expected: []item{
				{typ: itemAssignment, val: "MCU = atmega32u4 "},
				{typ: itemComment, val: "# comment"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name:  "2 vars and a comment",
			input: "MCU = atmega32u4 # comment\nMOUSE_ENABLE=yes",
			expected: []item{
				{typ: itemAssignment, val: "MCU = atmega32u4 "},
				{typ: itemComment, val: "# comment\n"},
				{typ: itemAssignment, val: "MOUSE_ENABLE=yes"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a optional variable",
			input: "MCU ?= atmega32u4 # comment it",
			expected: []item{
				{typ: itemAssignment, val: "MCU ?= atmega32u4 "},
				{typ: itemComment, val: "# comment it"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a addition variable",
			input: "MCU += atmega32u4 # comment it 2",
			expected: []item{
				{typ: itemAssignment, val: "MCU += atmega32u4 "},
				{typ: itemComment, val: "# comment it 2"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a simply expanded variable",
			input: "MCU := atmega32u4",
			expected: []item{
				{typ: itemAssignment, val: "MCU := atmega32u4"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a simply expanded variable 2",
			input: "MCU ::= atmega32u4",
			expected: []item{
				{typ: itemAssignment, val: "MCU ::= atmega32u4"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a shell expanded variable",
			input: "MCU != atmega32u4",
			expected: []item{
				{typ: itemAssignment, val: "MCU != atmega32u4"},
				{typ: itemEOF},
			},
		},
	}

	var passed uint64
	var failed uint64
	group := func(t *testing.T) {
		// group them so we can wait until all tests are finished running
		for _, test := range tests {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				_lexer := lex(test.name, test.input)

				//fmt.Printf("test %s\n", test.name)
				tokens := []item{}
				for {
					token := _lexer.nextItem()

					//fmt.Printf("%s\n", token)
					tokens = append(tokens, token)
					if token.typ == itemEOF {
						break
					}
				}
				if len(test.expected) != len(tokens) {
					t.Errorf("ERROR %s: expected %d tokens, got %d\n", test.name, len(test.expected), len(tokens))
					atomic.AddUint64(&failed, 1)
					return
				}
				for i, _item := range test.expected {
					//			t.Logf("item %v %v", _item, tokens[i])
					if _item.String() != tokens[i].String() {
						t.Errorf("ERROR %s: expected token %s, got %s\n", test.name, _item, tokens[i])
						atomic.AddUint64(&failed, 1)
						return
					}
				}
				atomic.AddUint64(&passed, 1)
			})
		}
	}
	t.Run("lexer group", group)
	t.Logf("%d tests %d passed %d failed", len(tests), passed, failed)
}
