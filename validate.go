package goisbn

import (
	"strconv"
)

func validate10(isbn10 string) bool {
	if len(isbn10) == 10 {
		s := sum10(isbn10)
		return s%11 == 0
	}
	return false
}

func validate13(isbn13 string) bool {
	if len(isbn13) == 13 {
		s := sum13(isbn13)
		return s%10 == 0
	}
	return false
}

func sum10(isbn string) int {
	s := 0
	w := 10
	for k, v := range isbn {
		if k == 9 && v == 88 {
			s += 10
		} else {
			n, err := strconv.Atoi(string(v))
			if err != nil {
				return -1
			}
			s += n * w
		}
		w--
	}
	return s
}

func sum13(isbn string) int {
	s := 0
	w := 1
	for _, v := range isbn {
		n, err := strconv.Atoi(string(v))
		if err != nil {
			return -1
		}
		s += n * w
		if w == 1 {
			w = 3
		} else {
			w = 1
		}
	}
	return s
}
