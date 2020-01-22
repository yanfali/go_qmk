package main

// liberally copied from https://golang.org/src/text/template/parse/lex.go

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type itemType int

const (
	itemError itemType = iota
	itemComment
	itemAssignment
	/*
		itemFunction
		itemRule
		itemConditional
		itemDefine
		itemInclude
		itemExport
	*/
	itemEOF
)

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemComment:
		return fmt.Sprintf("COMMENT %q", i.val)
	case itemAssignment:
		return fmt.Sprintf("ASSIGNMENT %q", i.val)
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type stateFn func(*lexer) stateFn

func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		state: lexText,
		input: input,
		items: make(chan item, 2),
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
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

const eof = -1

const startComment = "#"

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], startComment) {
			return lexComment(l)
		}

		r := l.next()
		if r == eof {
			break
		}
		if r == ' ' || r == '\n' {
			l.ignore()
		} else {
			l.backup()
			return lexNextToken(l)
		}

	}
	l.emit(itemEOF)
	return nil
}

func lexNextToken(l *lexer) stateFn {
	l.acceptRun("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	// found a simple assignment
	for {
		r := l.next()
		// recursively expanded
		// simply expanded variable
		if r == ':' && l.peek() == '=' {
			l.pos = l.start
			return lexAssignment
		}

		// simply expanded variable
		if r == ':' && l.accept(":") && l.accept(":") {
			l.pos = l.start
			return lexAssignment
		}

		// optional setting and append setting
		if (r == '?' || r == '+') && l.peek() == '=' {
			l.pos = l.start
			return lexAssignment
		}
		if r == '=' {
			l.pos = l.start
			return lexAssignment
		}

		if r == eof || r == '\n' {
			goto exit
		}
	}
exit:
	return lexText
}

func lexComment(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		// continuation
		// scan until new line or eof
		case r == '\\':
			for {
				r1 := l.next()
				if r1 == eof || r1 == '\n' {
					break
				}
			}
			break
		// end of comment
		case r == eof || r == '\n':
			l.emit(itemComment)
			return lexText
		}
	}
}

func lexAssignment(l *lexer) stateFn {
	for {
		r := l.next()
		if r == eof || r == '\n' {
			goto emitAssignment
		}
		if r == '#' {
			l.backup()
			goto emitAssignment
		}
	}
emitAssignment:
	l.emit(itemAssignment)
	return lexText
}

/*
func lexFunction(l *lexer) stateFn {
	return lexText
}

func lexRule(l *lexer) stateFn {
	return lexText
}

func lexConditional(l *lexer) stateFn {
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
