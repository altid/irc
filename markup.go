package main

import (
	"bytes"
	"fmt"

	"github.com/altid/libs/markup"
)

func input(l *markup.Lexer) (*msg, error) {
	var m bytes.Buffer
	for {
		i := l.Next()
		switch i.ItemType {
		case markup.EOF:
			d := m.String()
			m := &msg{
				data: d,
				fn:   fself,
			}
			return m, nil
		case markup.ErrorText:
			return nil, fmt.Errorf("error parsing input: %v", i.Data)
		case markup.UrlLink, markup.UrlText, markup.ImagePath, markup.ImageLink, markup.ImageText:
			continue
		case markup.ColorText, markup.ColorTextBold:
			m.WriteString(getColors(i.Data, l))
		case markup.BoldText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case markup.EmphasisText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case markup.UnderlineText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		default:
			m.Write(i.Data)
		}
	}
}

func getColors(current []byte, l *markup.Lexer) string {
	var text bytes.Buffer
	var color bytes.Buffer
	text.Write(current)
	for {
		i := l.Next()
		switch i.ItemType {
		case markup.EOF:
			return color.String()
		case markup.ColorCode:
			code := getColorCode(i.Data)
			if n := bytes.IndexByte(i.Data, ','); n >= 0 {
				code = getColorCode(i.Data[:n])
				code += ","
				code += getColorCode(i.Data[n+1:])
			}
			color.WriteString("")
			color.WriteString(code)
			color.WriteString(text.String())
			color.WriteString("")
			return color.String()
		case markup.ColorTextBold:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case markup.ColorTextEmphasis:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case markup.ColorTextUnderline:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case markup.ColorText:
			text.Write(i.Data)
		}
	}
}

func getColorCode(d []byte) string {
	switch string(d) {
	case markup.White:
		return "0"
	case markup.Black:
		return "1"
	case markup.Blue:
		return "2"
	case markup.Green:
		return "3"
	case markup.Red:
		return "4"
	case markup.Brown:
		return "5"
	case markup.Purple:
		return "6"
	case markup.Orange:
		return "7"
	case markup.Yellow:
		return "8"
	case markup.LightGreen:
		return "9"
	case markup.Cyan:
		return "10"
	case markup.LightCyan:
		return "11"
	case markup.LightBlue:
		return "12"
	case markup.Pink:
		return "13"
	case markup.Grey:
		return "14"
	case markup.LightGrey:
		return "15"
	}
	return ""
}
