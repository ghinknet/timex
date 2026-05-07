package timex

import (
	"strings"

	"go.gh.ink/regexp"
)

var patternDetect = regexp.MustCompile(`[yMdHhmsSaEzZ]{2,}`)

func hasPattern(s string) bool {
	return patternDetect.MatchString(s)
}

func convertLayout(layout string) string {
	parts := strings.Split(layout, "'")
	for i := 0; i < len(parts); i += 2 {
		parts[i] = tokenRe.ReplaceAllStringFunc(parts[i], mapToken)
	}
	return strings.Join(parts, "")
}

var tokenRe = regexp.MustCompile(`[y]+|[M]+|[d]+|[H]+|[h]+|[m]+|[s]+|[S]+|[a]+|[E]+|[z]+|[Z]+`)

func mapToken(token string) string {
	switch token[0] {
	case 'y':
		if len(token) >= 4 {
			return "2006" // yyyy
		}
		return "06" // yy
	case 'M':
		switch len(token) {
		case 1:
			return "1"
		case 2:
			return "01"
		case 3:
			return "Jan" // MMM
		default:
			return "January" // MMMM
		}
	case 'd':
		if len(token) == 1 {
			return "2"
		}
		return "02" // dd
	case 'H':
		return "15" // HH or H
	case 'h':
		if len(token) == 1 {
			return "3"
		}
		return "03" // hh
	case 'm':
		if len(token) == 1 {
			return "4"
		}
		return "04" // mm
	case 's':
		if len(token) == 1 {
			return "5"
		}
		return "05" // ss
	case 'S':
		return strings.Repeat("0", len(token))
	case 'a':
		return "PM" // am/pm
	case 'E':
		if len(token) <= 3 {
			return "Mon" // E/EEE
		}
		return "Monday" // EEEE
	case 'z':
		return "MST"
	case 'Z':
		return "-0700"
	}
	return token
}
