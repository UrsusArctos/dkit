package stringent

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
