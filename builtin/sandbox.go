package builtin

import (
	"reflect"
)

type Sandbox struct{}

func newSandbox() *Sandbox {
	return &Sandbox{}
}

func (s *Sandbox) Symbols() map[string]map[string]reflect.Value {
	upper := reflect.ValueOf(func(text string) string {
		return toUpper(text)
	})

	lower := reflect.ValueOf(func(text string) string {
		return toLower(text)
	})

	return map[string]map[string]reflect.Value{
		"strings": {
			"Upper": upper,
			"Lower": lower,
		},
	}
}

func toUpper(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			r = r - 'a' + 'A'
		}
		out = append(out, r)
	}
	return string(out)
}

func toLower(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		out = append(out, r)
	}
	return string(out)
}
