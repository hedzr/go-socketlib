package main

func skipws(data []byte, from int) (pos int) {
	pos = from
	for pos < len(data) && data[pos] == ' ' {
		pos++
	}
	return
}

func tillcr(data []byte, from int) (str string, pos int) {
	str, pos = tillChars(data, from, '\r', '\n')
	return
}

func tillChars(data []byte, from int, chars ...byte) (str string, pos int) {
	pos = from
	for pos < len(data) && notEquChars(data[pos], chars...) {
		pos++
	}
	str = string(data[from:pos])
	for pos < len(data) && equChars(data[pos], chars...) {
		pos++
	}
	return
}

func equChars(ch byte, chars ...byte) (ok bool) {
	for _, c := range chars {
		if ok = ch == c; ok {
			break
		}
	}
	return
}

func notEquChars(ch byte, chars ...byte) (ne bool) {
	ne = true
	for _, c := range chars {
		if ne = ch != c; !ne {
			break
		}
	}
	return
}
