package toolspkg

func CutString(s string, from int, to int) string {
	r := []rune(s)

	if from < to && len(r) > to {
		r = r[from:to]
	}

	return string(r)
}
