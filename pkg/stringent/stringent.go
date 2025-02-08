package stringent

import "strings"

func BreakString(s string, partlen int) []string {
	r := []rune(s)
	subs := ((len(r) - 1) / partlen) + 1
	bres := make([]string, subs)
	for parti := 0; parti < subs; parti++ {
		if parti < subs-1 {
			bres[parti] = string(r[parti*partlen : (parti+1)*partlen])
		} else {
			bres[parti] = string(r[parti*partlen:])
		}
	}
	return bres
}

func BreakPath(s string) []string {
	return strings.Split(s, "/")
}

func StrPtr(s string) *string {
	return &s
}
