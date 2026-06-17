package config

// stripJSONComments removes // and /* */ comments outside JSON string literals.
func stripJSONComments(data []byte) []byte {
	var out []byte
	inString := false
	escape := false
	for i := 0; i < len(data); i++ {
		if inString {
			out, escape, inString = appendStringByte(out, data[i], escape, inString)
			continue
		}
		if data[i] == '"' {
			inString = true
			out = append(out, data[i])
			continue
		}
		if lineComment, skip := skipLineComment(data, i); skip {
			i += lineComment
			continue
		}
		if blockComment, skip := skipBlockComment(data, i); skip {
			i += blockComment
			continue
		}
		out = append(out, data[i])
	}
	return out
}

func appendStringByte(out []byte, b byte, escape, inString bool) ([]byte, bool, bool) {
	out = append(out, b)
	if escape {
		return out, false, inString
	}
	switch b {
	case '\\':
		return out, true, inString
	case '"':
		return out, false, false
	default:
		return out, false, inString
	}
}

func skipLineComment(data []byte, i int) (advance int, skip bool) {
	if i+1 >= len(data) || data[i] != '/' || data[i+1] != '/' {
		return 0, false
	}
	j := i + 2
	for j < len(data) && data[j] != '\n' {
		j++
	}
	return j - i - 1, true
}

func skipBlockComment(data []byte, i int) (advance int, skip bool) {
	if i+1 >= len(data) || data[i] != '/' || data[i+1] != '*' {
		return 0, false
	}
	j := i + 2
	for j+1 < len(data) && !(data[j] == '*' && data[j+1] == '/') {
		j++
	}
	if j+1 < len(data) {
		j++
	}
	return j - i, true
}
