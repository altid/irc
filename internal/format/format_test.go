package format

import (
	"testing"

	"github.com/altid/libs/markup"
)

func TestInput(t *testing.T) {
	tests := map[string]string{
		"**bold test**": "bold test",
		"inline **bold** test": "inline bold test",
		"*emphasis test*": "emphasis test",
		"inline *emphasis* test": "inline emphasis test",
		"%[coloured text test](blue)": "2coloured text test",
		//"%[coloured text with inline **bold** ](blue)": "2coloured text with inline bold",
	}

	for key, value := range tests {
		l := markup.NewStringLexer(key)
		out, _ := Input(l)

		if out != value {
			t.Error("mismatched values")
		}
	}
}
