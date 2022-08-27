package format

import (
	"bytes"
	"fmt"

	mp "github.com/altid/libs/markup"
)

var colorCode = map[string]string{
	mp.White:      "0",
	mp.Black:      "1",
	mp.Blue:       "2",
	mp.Green:      "3",
	mp.Red:        "4",
	mp.Brown:      "5",
	mp.Purple:     "6",
	mp.Orange:     "7",
	mp.Yellow:     "8",
	mp.LightGreen: "9",
	mp.Cyan:       "10",
	mp.LightCyan:  "11",
	mp.LightBlue:  "12",
	mp.Pink:       "13",
	mp.Grey:       "14",
	mp.LightGrey:  "15",
}

func Input(l *mp.Lexer) (string, error) {
	var m bytes.Buffer
	for {
		i := l.Next()
		switch i.ItemType {
		case mp.EOF:
			return m.String(), nil
		case mp.ErrorText:
			return "", fmt.Errorf("error parsing input: %v", i.Data)
		case mp.URLLink, mp.URLText, mp.ImagePath, mp.ImageLink, mp.ImageText:
			continue
		case mp.ColorText, mp.ColorTextBold:
			data, err := getColors(i.Data, l)
			if err != nil {
				return "", err
			}

			m.WriteString(data)
		case mp.BoldText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case mp.EmphasisText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case mp.StrongText:
			m.WriteString("")
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
			m.WriteString("")
		default:
			m.Write(i.Data)
		}
	}
}

func getColors(current []byte, l *mp.Lexer) (string, error) {
	var text bytes.Buffer
	var color bytes.Buffer

	text.Write(current)

	for {
		i := l.Next()
		switch i.ItemType {
		case mp.ErrorText:
			return "", fmt.Errorf("%s", i.Data)
		case mp.EOF:
			return color.String(), nil
		case mp.ColorCode:
			code := colorCode[string(i.Data)]
			if n := bytes.IndexByte(i.Data, ','); n >= 0 {
				code = colorCode[string(i.Data[:n])]
				code += ","
				code += colorCode[string(i.Data[n+1:])]
			}
			color.WriteString("")
			color.WriteString(code)
			color.WriteString(text.String())
			color.WriteString("")
			return color.String(), nil
		case mp.ColorTextBold:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case mp.ColorTextEmphasis:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case mp.ColorTextStrong:
			text.WriteString("")
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
			text.WriteString("")
		case mp.ColorText:
			text.Write(i.Data)
		}
	}
}
