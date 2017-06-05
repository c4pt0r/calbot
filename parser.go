package main

import (
	"strconv"
	"time"

	"github.com/jinzhu/now"
	"github.com/juju/errors"
)

type matchFunc func() error

func chain(calls ...matchFunc) error {
	for _, fn := range calls {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

type parser struct {
	l         *lexer
	cur       int
	prevCnt   int
	nextCnt   int
	weekday   int
	morning   bool
	afternoon bool
	hour      int
	min       int
}

func newParser(l *lexer) *parser {
	return &parser{l: l}
}

func (p *parser) curToken() *Token {
	if p.cur > len(p.l.tokens)-1 {
		return nil
	}
	return p.l.tokens[p.cur]
}

func (p *parser) accept() {
	p.cur++
}

func (p *parser) matchPrefix() error {
	if p.curToken().typ == Next {
		p.nextCnt++
		p.accept()
	} else if p.curToken().typ == Prev {
		p.prevCnt++
		p.accept()
	} else {
		return nil
	}
	p.matchPrefix()
	return nil
}

func (p *parser) matchWeekday() error {
	if p.curToken().typ == WeekDay {
		days := []rune("一二三四五六日")
		for i, day := range days {
			if day == []rune(p.curToken().val)[0] {
				p.weekday = i + 1
				p.accept()
				return nil
			}
		}
		if []rune(p.curToken().val)[0] == '天' {
			p.weekday = 7
			p.accept()
			return nil
		}
	}
	return errors.Errorf("invalid weekday")
}

func (p *parser) matchDay() error {
	if p.curToken().typ == Week {
		p.accept()
		return p.matchWeekday()
	}
	return errors.Errorf("invalid weekday")
}

func (p *parser) matchMorningOrAfternoon() error {
	if p.curToken().typ == Morning {
		p.morning = true
		p.accept()
	} else if p.curToken().typ == Afternoon {
		p.afternoon = true
		p.accept()
	} else {
		return nil
	}
	return nil
}

func (p *parser) matchHour() error {
	hour := ""
	for p.curToken() != nil && p.curToken().typ == Num {
		hour += p.curToken().val
		p.accept()
	}
	p.accept()
	var err error
	p.hour, err = strconv.Atoi(hour)
	if err != nil {
		return errors.Errorf("invalid hour")
	}
	return nil
}

func (p *parser) matchMin() error {
	min := ""
	for p.curToken() != nil && p.curToken().typ == Num {
		min += p.curToken().val
		p.accept()
	}
	p.accept()
	var err error
	p.min, err = strconv.Atoi(min)
	if err != nil {
		return errors.Errorf("invalid num")
	}
	return nil
}

func (p *parser) matchTime() error {
	return chain(
		p.matchMorningOrAfternoon,
		p.matchHour,
		p.matchMin,
	)
}
func (p *parser) Run() error {
	if err := p.l.run(); err != nil {
		return err
	}
	return chain(
		p.matchPrefix,
		p.matchDay,
		p.matchTime,
	)
}

// weekday: 1~7
func weekdayTime(w int) time.Time {
	now.FirstDayMonday = true
	t := now.BeginningOfDay()
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	d := time.Duration(-weekday+w) * 24 * time.Hour
	return t.Truncate(time.Hour).Add(d)
}

func (p *parser) Exec() time.Time {
	if p.afternoon {
		p.hour += 12
	}
	return weekdayTime(p.weekday).
		Add(time.Duration(p.hour)*time.Hour).
		Add(time.Duration(p.min)*time.Minute).
		AddDate(0, 0, p.nextCnt*7)
}
