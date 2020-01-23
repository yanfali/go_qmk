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
		{
			name:  "a comment",
			input: "#comment",
			expected: []item{
				{line: 1, typ: itemComment, val: "#comment"},
				{line: 2, typ: itemEOF, val: ""},
			},
		},
		{
			name:  "a multline",
			input: "#comment \\\nsome more comment",
			expected: []item{
				{line: 1, typ: itemComment, val: "#comment \\\nsome more comment"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name:  "a multline with one space before next line",
			input: "#comment \\ \nsome more comment\n#a new comment",
			expected: []item{
				{line: 1, typ: itemComment, val: "#comment \\ \nsome more comment"},
				{line: 3, typ: itemComment, val: "#a new comment"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name:  "a multline with 3 spaces before next line",
			input: "#comment \\   \nsome more comment\n",
			expected: []item{
				{line: 1, typ: itemComment, val: "#comment \\   \nsome more comment"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name: "a recursively expanded variable", input: "MCU = atmega32u4",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "MCU = atmega32u4"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name:  "a variable and a comment",
			input: "MCU = atmega32u4 # comment",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "MCU = atmega32u4 "},
				{line: 1, typ: itemComment, val: "# comment"},
				{typ: itemEOF, val: ""},
			},
		},
		{
			name:  "2 vars and a comment",
			input: "MCU = atmega32u4 # comment\nMOUSE_ENABLE=yes",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "MCU = atmega32u4 "},
				{line: 1, typ: itemComment, val: "# comment"},
				{line: 2, typ: itemAssignment, val: "MOUSE_ENABLE=yes"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a optional variable",
			input: "MCU ?= atmega32u4 # comment it",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "MCU ?= atmega32u4 "},
				{line: 1, typ: itemComment, val: "# comment it"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a addition variable",
			input: "MCU += atmega32u4 # comment it 2",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "MCU += atmega32u4 "},
				{line: 1, typ: itemComment, val: "# comment it 2"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a simply expanded variable",
			input: "MCU := atmega32u4",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "MCU := atmega32u4"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a simply expanded variable 2",
			input: "MCU ::= atmega32u4",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "MCU ::= atmega32u4"},
				{typ: itemEOF},
			},
		},
		{
			name:  "a shell expanded variable",
			input: "MCU != atmega32u4",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "MCU != atmega32u4"},
				{typ: itemEOF},
			},
		},
		{
			name:  "empty",
			input: "",
			expected: []item{
				{typ: itemEOF},
			},
		},
		{
			name:  "empty spaces",
			input: "    \r \t \n",
			expected: []item{
				{typ: itemEOF},
			},
		},
		{
			name:  "windows eol",
			input: "#comment\r\nMCU = stm32",
			expected: []item{
				{line: 1, typ: itemComment, val: "#comment"},
				{line: 2, typ: itemAssignment, val: "MCU = stm32"},
				{typ: itemEOF},
			},
		},
		{
			name:  "multi-line assignment",
			input: "SRC =\tkeyboards/wilba_tech/wt_main.c \\\n\t\tkeyboards/wilba_tech/wt_rgb_backlight.c \\\n\t\tdrivers/issi/is31fl3733.c \\\n\t\tquantum/color.c \\\n\t\tdrivers/arm/i2c_master.c",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "SRC =\tkeyboards/wilba_tech/wt_main.c \\\n\t\tkeyboards/wilba_tech/wt_rgb_backlight.c \\\n\t\tdrivers/issi/is31fl3733.c \\\n\t\tquantum/color.c \\\n\t\tdrivers/arm/i2c_master.c"},
				{typ: itemEOF},
			},
		},
		{
			name:  "multi-line assignment and comment",
			input: "SRC =\tkeyboards/wilba_tech/wt_main.c \\\n\t\tkeyboards/wilba_tech/wt_rgb_backlight.c \\\n\t\tdrivers/issi/is31fl3733.c \\\n\t\tquantum/color.c \\\n\t\tdrivers/arm/i2c_master.c\n# a comment",
			expected: []item{
				{line: 1, typ: itemAssignment, val: "SRC =\tkeyboards/wilba_tech/wt_main.c \\\n\t\tkeyboards/wilba_tech/wt_rgb_backlight.c \\\n\t\tdrivers/issi/is31fl3733.c \\\n\t\tquantum/color.c \\\n\t\tdrivers/arm/i2c_master.c"},
				{line: 6, typ: itemComment, val: "# a comment"},
				{typ: itemEOF},
			},
		},
		{
			name:  "conditional",
			input: "ifeq endif",
			expected: []item{
				{line: 1, typ: itemConditional, val: "ifeq endif"},
				{typ: itemEOF},
			},
		},
		{
			name:  "conditional",
			input: "ifneq endif",
			expected: []item{
				{line: 1, typ: itemConditional, val: "ifneq endif"},
				{typ: itemEOF},
			},
		},
		{
			name:  "conditional",
			input: "ifdef endif",
			expected: []item{
				{line: 1, typ: itemConditional, val: "ifdef endif"},
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
					t.Errorf("ERROR %s: expected %d tokens, got %d (%+v)\n", test.name, len(test.expected), len(tokens), tokens)
					atomic.AddUint64(&failed, 1)
					return
				}
				for i, _item := range test.expected {
					//			t.Logf("item %v %v", _item, tokens[i])
					if _item.String() != tokens[i].String() {
						t.Errorf("ERROR %s on line %d: expected token %s, got %s (%+v)\n", test.name, tokens[i].line, _item, tokens[i], tokens)
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
