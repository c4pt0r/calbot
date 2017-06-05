package main

import (
	"fmt"
	"unicode"

	"github.com/juju/errors"
)

type TokenType int

const (
	Rune TokenType = iota
	Month
	Day
	Hour
	Min
	Week
	WeekDay // Monday to Sunday
	Next
	Prev
	Num // 1~60
	Morning
	Afternoon
)

func (v TokenType) String() string {
	return tokStr[v]
}

var tokStr = []string{
	"rune",
	"mon",
	"day",
	"hour",
	"min",
	"week",
	"weekday",
	"next",
	"prev",
	"num",
	"morning",
	"afternoon",
}

type Span struct {
	start, end int
}

type Token struct {
	typ   TokenType
	val   string
	where Span
}

func (t *Token) String() string {
	return fmt.Sprintf("[%s:%s]", t.val, t.typ)
}

type lexer struct {
	input          []rune
	start, end     int
	curOffset      int
	tokStartOffset int
	tokens         []*Token
}

func isEOF(r rune) bool {
	return r == 0
}

func isSpace(r rune) bool {
	return r == '\t' || unicode.IsSpace(r)
}

func (l *lexer) skipSpace() {
	for {
		for isSpace(l.peek(0)) {
			l.consume(1)
		}
		l.discard()
		break
	}
}

func (l *lexer) peek(n int) rune {
	if l.end+n >= len(l.input) {
		return 0
	}
	return l.input[l.end+n]
}

func (l *lexer) consume(n int) {
	for i := 0; i < n; i++ {
		l.end++
	}
}

func (l *lexer) pushToken(t TokenType) {
	ret := &Token{
		typ:   t,
		val:   string(l.input[l.start:l.end]),
		where: Span{l.start, l.end},
	}
	l.tokens = append(l.tokens, ret)
	l.discard()
}

func (l *lexer) discard() {
	l.start = l.end
}

func (l *lexer) consumeNext() {
	l.consume(1)
	l.pushToken(Next)
}

func (l *lexer) consumePrev() {
	l.consume(1)
	l.pushToken(Prev)
}

func (l *lexer) consumeWeek(n int) {
	l.consume(n)
	l.pushToken(Week)
}

func (l *lexer) consumeMonth() {
	l.consume(1)
	l.pushToken(Month)
}

func (l *lexer) consumeWeekday() {
	l.consume(1)
	l.pushToken(WeekDay)
}

func (l *lexer) consumeMorning() {
	l.consume(2)
	l.pushToken(Morning)
}

func (l *lexer) consumeAfternoon() {
	l.consume(2)
	l.pushToken(Afternoon)
}

func (l *lexer) consumeHour() {
	l.consume(1)
	l.pushToken(Hour)
}

func (l *lexer) consumeMin() {
	l.consume(1)
	l.pushToken(Min)
}

func (l *lexer) consumeClocknum() {
	l.consume(1)
	l.pushToken(Num)
}

func (l *lexer) run() error {
	for {
		l.skipSpace()
		if isEOF(l.peek(0)) {
			return nil
		} else if l.peek(0) == '周' {
			l.consumeWeek(1)
		} else if l.peek(0) == '月' {
			l.consumeMonth()
		} else if l.peek(0) == '一' ||
			l.peek(0) == '二' ||
			l.peek(0) == '三' ||
			l.peek(0) == '四' ||
			l.peek(0) == '五' ||
			l.peek(0) == '六' ||
			l.peek(0) == '天' ||
			l.peek(0) == '日' {
			l.consumeWeekday()
		} else if l.peek(0) == '下' {
			if l.peek(1) == '午' {
				l.consumeAfternoon()
			} else {
				l.consumeNext()
			}
		} else if l.peek(0) == '上' {
			if l.peek(1) == '午' {
				l.consumeMorning()
			} else {
				l.consumePrev()
			}
		} else if l.peek(0) == '星' && l.peek(1) == '期' {
			l.consumeWeek(2)
		} else if l.peek(0) >= '0' && l.peek(0) <= '9' {
			l.consumeClocknum()
		} else if l.peek(0) == '时' || l.peek(0) == '点' {
			l.consumeHour()
		} else if l.peek(0) == '分' {
			l.consumeMin()
		} else {
			msg := string(l.input[0:l.start]) + "[" + string(l.input[l.start]) + "]" + string(l.input[l.start+1:len(l.input)])
			return errors.Errorf("lexer: unknown token: %s", msg)
		}
	}
	return nil
}
