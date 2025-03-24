package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type item struct {
	typ itemType
	val string

	pos  int
	line int
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemCommand:
		return fmt.Sprintf("cmd: %s", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type itemType int

const (
	itemError itemType = iota
	itemEOF

	itemIdentifier
	itemVariable

	// predefined identifiers
	itemSelfStar
	itemStream
	itemUnderscore

	// number operators
	itemDiv
	itemMinus
	itemMult
	itemPlus

	// stream operators
	itemAssign
	itemDeclare
	itemPipe

	// list operators
	itemConcatOp

	// literals
	itemNumber
	itemString
	itemBool // TODO

	// delimiters
	itemComma
	itemLeftBrace
	itemLeftParen
	itemNewline
	itemRightBrace
	itemRightParen

	// comment
	itemComment

	// commands
	itemCommand // to delimit commands
	itemBrigtness
	itemConcat
	itemContrast
	itemCrossfade
	itemCut
	itemFade
	itemHue
	itemMap
	itemOpen
	itemPitch
	itemSaturation
	itemSpeed
	itemTrackLine
	itemVolume
)

const globalStream = "stream"
const selfStar = "*"

var runeKeywords = map[rune]itemType{
	'(':  itemLeftParen,
	')':  itemRightParen,
	',':  itemComma,
	'[':  itemLeftBrace,
	']':  itemRightBrace,
	'\n': itemNewline,

	'_': itemUnderscore,

	'*': itemMult,
	'+': itemPlus,
	'-': itemMinus,
	'/': itemDiv,
	'=': itemAssign,
}

var strOperators = map[string]itemType{
	":=": itemDeclare,
	"|>": itemPipe,
	"..": itemConcatOp,
}

func isStrOperator(s string) bool {
	_, ok := strOperators[s]
	return ok
}

var commands = map[string]itemType{
	"brightness": itemBrigtness,
	"concat":     itemConcat,
	"contrast":   itemContrast,
	"crossfade":  itemCrossfade,
	"cut":        itemCut,
	"fade":       itemFade,
	"hue":        itemHue,
	"map":        itemMap,
	"open":       itemOpen,
	"pitch":      itemPitch,
	"saturation": itemSaturation,
	"speed":      itemSpeed,
	"trackline":  itemTrackLine,
	"volume":     itemVolume,
}

func isCommand(s string) bool {
	_, ok := commands[s]
	return ok
}

const eof = -1

type lexer struct {
	input         string    // string scanned
	start         int       // start position of this item
	pos           int       // current input position
	startLine     int       // start line
	line          int       // current line
	width         int       // width of last rune read from input
	items         chan item // channel of scanned items
	allowSelfStar bool      // whether `*` is a selfstar or not
	reachedEOF    bool      // whether EOF has been reached
}

type stateFn func(*lexer) stateFn

func lex(input string) *lexer {
	l := &lexer{
		input:         input,
		items:         make(chan item),
		allowSelfStar: false,
	}
	go run(l)
	return l
}

func run(l *lexer) {
	for state := lexScript; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos], l.start, l.startLine}
	l.start = l.pos
	l.startLine = l.line
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...any) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...), l.start, l.startLine}
	l.start = 0
	l.pos = 0
	l.input = l.input[:0]
	return nil
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

// ignore skips over the pending input before this point.
// It tracks newlines in the ignored text, so use it only
// for text that is skipped without calling l.next.
func (l *lexer) ignore() {
	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.reachedEOF = true
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	if r == '\n' {
		l.line++
	}
	return r
}

// backup steps back one rune.
func (l *lexer) backup() {
	if l.pos > 0 {
		r, w := utf8.DecodeLastRuneInString(l.input[:l.pos])
		l.pos -= w
		// Correct newline count.
		if r == '\n' {
			l.line--
		}
	}
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func lexScript(l *lexer) stateFn {
	if l.reachedEOF {
		l.emit(itemEOF)
		return nil
	}

	for isSpace(l.peek()) {
		l.next()
		l.ignore()
	}

	r := l.next()
	switch {
	case r == eof:
		l.emit(itemEOF)
		return nil
	case r == '#':
		return lexComment
	case r == '"':
		return lexString
	case unicode.IsDigit(r):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	}

	if op, ok := runeKeywords[r]; ok {
		if op == itemMult && l.allowSelfStar {
			op = itemSelfStar
		}
		if op == itemAssign || op == itemLeftBrace {
			l.allowSelfStar = true
		}
		if op == itemRightParen {
			l.allowSelfStar = false
		}
		l.emit(op)
		return lexScript
	}

	w := string(r)
	p := l.next()
	if p == eof {
		l.emit(itemEOF)
		return nil
	}
	w += string(p)
	if op, ok := strOperators[w]; ok {
		l.emit(op)
		if op == itemDeclare {
			l.allowSelfStar = true
		}
		if op == itemPipe {
			l.allowSelfStar = false
		}
		return lexScript
	}

	return l.errorf("unexpected character %#U", r)
}

func lexComment(l *lexer) stateFn {
	for l.peek() == '#' {
		l.next()
		l.ignore()
	}
	for {
		c := l.next()
		switch c {
		case '\n':
			l.emit(itemComment)
			return lexScript
		case eof:
			l.emit(itemComment)
			l.emit(itemEOF)
			return nil
		}
	}
}

func lexIdentifier(l *lexer) stateFn {
	var r rune
	for r = l.next(); isAlphaNumeric(r); {
		r = l.next()
	}
	l.backup()

	word := l.input[l.start:l.pos]

	if item, ok := commands[word]; ok {
		l.emit(item)
		return lexScript
	}

	switch word {
	case globalStream:
		l.emit(itemStream)
	default:
		l.emit(itemIdentifier)
	}
	if r == eof {
		l.emit(itemEOF)
		return nil
	}
	return lexScript
}

// lexNumber scans a number: decimal, float
func lexNumber(l *lexer) stateFn {
	// Optional leading sign
	l.accept("+-")
	// Is it a number?
	digits := "0123456789"
	l.acceptRun(digits)

	// Decimal point?
	if l.accept(".") {
		l.acceptRun(digits)
	}

	l.emit(itemNumber)
	return lexScript
}

// lexString scans a string literal. The opening quote has already been consumed
func lexString(l *lexer) stateFn {
	for {
		r := l.next()
		if r == eof {
			return l.errorf("unterminated string")
		}
		if r == '\\' {
			// Handle escape sequence
			r = l.next()
			if r == eof {
				return l.errorf("unterminated string escape")
			}
			continue
		}
		if r == '"' {
			break
		}
	}
	l.emit(itemString)
	return lexScript
}

func isAlphaNumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}
