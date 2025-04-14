package parser

import (
	"fmt"
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
	itemExport
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

func (i itemType) String() string {
	switch i {
	case itemEOF:
		return "EOF"
	case itemError:
		return "error"
	case itemIdentifier:
		return "identifier"
	case itemString, itemNumber, itemBool:
		return "literal"
	case itemComment:
		return "comment"
	case itemNewline:
		return "newline"
	default:
		for k, v := range commands {
			if v == i {
				return k
			}
		}
		for k, v := range runeKeywords {
			if v == i {
				return string(k)
			}
		}
		for k, v := range strOperators {
			if v == i {
				return k
			}
		}
		return "unknown"
	}
}

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

var mathSymbols = map[string]itemType{
	"*": itemMult,
	"+": itemPlus,
	"-": itemMinus,
	"/": itemDiv,
	"(": itemLeftParen,
	")": itemRightParen,
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
	"export":     itemExport,
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
