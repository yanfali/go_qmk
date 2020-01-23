package main

// liberally copied from https://golang.org/src/text/template/parse/lex.go

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type itemType int

const (
	itemError itemType = iota
	itemComment
	itemAssignment
	itemConditional
	/*
		itemFunction
		itemRule
		itemDefine
		itemInclude
		itemExport
	*/
	itemEOF
)

type item struct {
	typ  itemType
	val  string
	line int
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemComment:
		return fmt.Sprintf("%d COMMENT %q", i.line, i.val)
	case itemAssignment:
		return fmt.Sprintf("%d ASSIGNMENT %q", i.line, i.val)
	case itemConditional:
		return fmt.Sprintf("%d CONDITIONAL %q", i.line, i.val)
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%d %.10q...", i.line, i.val)
	}
	return fmt.Sprintf("%d %q", i.line, i.val)
}

type stateFn func(*lexer) stateFn

func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		state: lexText,
		input: input,
		items: make(chan item, 2),
		line:  1,
	}
	return l
}

type lexer struct {
	name  string    // used only for error reports.
	input string    // the string being scanned.
	state stateFn   // current state of lexer
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
	line  int       // current line number
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos], l.line}
	l.start = l.pos
}

const eof = -1

const startComment = "#"

func lexText(l *lexer) stateFn {
	for {
		curStr := l.input[l.pos:]
		if strings.HasPrefix(curStr, startComment) {
			return lexComment
		}

		if strings.HasPrefix(curStr, "ifeq") ||
			strings.HasPrefix(curStr, "ifneq") ||
			strings.HasPrefix(curStr, "ifdef") {
			return lexConditional
		}

		r := l.next()
		if r == eof {
			break
		}
		if unicode.IsSpace(r) {
			switch r {
			case '\r':
				l.ignore()
				break
			case '\n':
				// count lines for debugging
				l.line++
				l.ignore()
				break
			}
		}

		wholeStr := l.input[l.start:]
		if l.start != l.pos && strings.Contains(wholeStr, "=") {
			return lexAssignment(l)
		}

	}
	l.emit(itemEOF)
	return nil
}

func (l *lexer) errorf(format string, values ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, values...),
		l.line,
	}
	return nil
}

func lexComment(l *lexer) stateFn {
	skipped := 0
	for {
		switch r := l.next(); r {
		// continuation
		// scan until new line or eof
		case '\\':
			for {
				if isEOL(l.next(), l) {
					skipped++
					break
				}
			}
			break
		// end of comment
		default:
			if isEOL(r, l) {
				l.backup()
				l.emit(itemComment)
				l.line += skipped
				return lexText
			}
		}
	}
}

func isEOL(r rune, l *lexer) bool {
	return r == eof || r == '\n' || (r == '\r' && l.peek() == '\n')
}

func lexAssignment(l *lexer) stateFn {
	skipped := 0
	for {
		r := l.next()
		// handle multiline assignment
		if r == '\\' {
			for {
				if isEOL(l.next(), l) {
					skipped++
					break
				}
			}
			continue
		}
		if isEOL(r, l) {
			goto emitAssignment
		}
		if r == '#' {
			goto emitAssignment
		}
	}
emitAssignment:
	l.backup()
	l.emit(itemAssignment)
	l.line += skipped
	return lexText
}

func lexConditional(l *lexer) stateFn {
	for {
		if strings.HasSuffix(l.input[l.start:l.pos], "endif") {
			l.emit(itemConditional)
			return lexText
		}
		l.next()
	}
}

/*
func lexFunction(l *lexer) stateFn {
	return lexText
}

func lexRule(l *lexer) stateFn {
	return lexText
}

func lexDefine(l *lexer) stateFn {
	return lexText
}

func lexInclude(l *lexer) stateFn {
	return lexText
}

func lexExport(l *lexer) stateFn {
	return lexText
}
*/

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w :=
		utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// This insight into concurrency by Pike is just amazing.
// Select against the channel if there's nothing to do,
// keep running the state machine otherwise return the token
func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
}
