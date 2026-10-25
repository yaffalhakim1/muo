package pdf

import (
	"bytes"
	"strconv"
)

// The object parser: a small recursive-descent reader over the file's bytes.
// Objects are parsed on demand from an offset, never all at once, so a large
// file costs only the objects a page needs.

// parseValue reads one object starting at at.
func (f *pdfFile) parseValue(at int) pdfValue {
	v, _ := f.parse(f.data, at, 0)
	return v
}

// parse reads one object from buf at at. Objects inside an object stream come
// from a decoded member, which is why the buffer is a parameter: a stream's
// payload is never searched for inside another stream's bytes.
func (f *pdfFile) parse(buf []byte, at, depth int) (pdfValue, int) {
	switch {
	case buf == nil:
		return pdfValue{}, at
	case at >= len(buf):
		return pdfValue{}, at
	}
	if depth > 32 {
		return pdfValue{}, at
	}
	at = skipSpace(buf, at)
	if at >= len(buf) {
		return pdfValue{}, at
	}
	switch c := buf[at]; {
	case c == '/':
		return f.parseName(buf, at)
	case c == '(':
		return f.parseString(buf, at)
	case c == '<' && at+1 < len(buf) && buf[at+1] == '<':
		return f.parseDict(buf, at, depth)
	case c == '<':
		return f.parseHexString(buf, at)
	case c == '[':
		return f.parseArray(buf, at, depth)
	case c >= '0' && c <= '9' || c == '-' || c == '+' || c == '.':
		return f.parseNumber(buf, at)
	default:
		return parseKeyword(buf, at)
	}
}

func (f *pdfFile) parseName(buf []byte, at int) (pdfValue, int) {
	at++ // '/'
	start := at
	for at < len(buf) && !isDelim(buf[at]) && !isSpace(buf[at]) {
		at++
	}
	name := string(buf[start:at])
	// A name may encode characters as #xx.
	if bytes.IndexByte(buf[start:at], '#') >= 0 {
		var out []byte
		for i := 0; i < len(name); i++ {
			if name[i] == '#' && i+2 < len(name) {
				if v := hexVal(name[i+1])*16 + hexVal(name[i+2]); v >= 0 {
					out = append(out, byte(v))
					i += 2
					continue
				}
			}
			out = append(out, name[i])
		}
		name = string(out)
	}
	return pdfValue{kind: pdfName, name: name}, at
}

func (f *pdfFile) parseNumber(buf []byte, at int) (pdfValue, int) {
	start := at
	if buf[at] == '-' || buf[at] == '+' {
		at++
	}
	for at < len(buf) {
		c := buf[at]
		if (c >= '0' && c <= '9') || c == '.' {
			at++
			continue
		}
		break
	}
	// A trailing "g R" makes this a reference.
	if save := skipSpace(buf, at); save != at {
		if _, next, err := readInt(buf, save); err == nil {
			if next = skipSpace(buf, next); next < len(buf) && buf[next] == 'R' {
				num, _ := strconv.Atoi(string(buf[start:at]))
				return pdfValue{kind: pdfRef, num: float64(num)}, next + 1
			}
		}
	}
	v, err := strconv.ParseFloat(string(buf[start:at]), 64)
	if err != nil {
		v = 0
	}
	return pdfValue{kind: pdfNumber, num: v}, at
}

func (f *pdfFile) parseString(buf []byte, at int) (pdfValue, int) {
	at++ // '('
	depth := 1
	var out []byte
	for at < len(buf) && depth > 0 {
		c := buf[at]
		switch c {
		case '\\':
			at++
			if at >= len(buf) {
				break
			}
			switch e := buf[at]; e {
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case 'b':
				out = append(out, '\b')
			case 'f':
				out = append(out, '\f')
			case '(', ')', '\\':
				out = append(out, e)
			case '\r':
				if at+1 < len(buf) && buf[at+1] == '\n' {
					at++
				}
			case '\n':
			default:
				if e >= '0' && e <= '7' {
					v := 0
					for k := 0; k < 3 && at < len(buf) && buf[at] >= '0' && buf[at] <= '7'; k++ {
						v = v*8 + int(buf[at]-'0')
						at++
					}
					at--
					out = append(out, byte(v))
				} else {
					out = append(out, e)
				}
			}
			at++
			continue
		case '(':
			depth++
			out = append(out, c)
		case ')':
			depth--
			if depth == 0 {
				return pdfValue{kind: pdfString, str: out}, at + 1
			}
			out = append(out, c)
		default:
			out = append(out, c)
		}
		at++
	}
	return pdfValue{kind: pdfString, str: out}, at
}

func (f *pdfFile) parseHexString(buf []byte, at int) (pdfValue, int) {
	end := bytes.IndexByte(buf[at:], '>')
	if end < 0 {
		return pdfValue{kind: pdfString}, len(buf)
	}
	return pdfValue{kind: pdfString, str: asciiHexDecode(buf[at+1 : at+end])}, at + end + 1
}

func (f *pdfFile) parseArray(buf []byte, at, depth int) (pdfValue, int) {
	at++ // '['
	var arr []pdfValue
	for at < len(buf) {
		at = skipSpace(buf, at)
		if at >= len(buf) {
			break
		}
		if buf[at] == ']' {
			return pdfValue{kind: pdfArray, array: arr}, at + 1
		}
		v, next := f.parse(buf, at, depth+1)
		if next == at {
			break
		}
		arr = append(arr, v)
		at = next
	}
	return pdfValue{kind: pdfArray, array: arr}, at
}

// parseDict reads a dictionary and, when a stream follows it in the file,
// remembers the offset it ended at in num, where streamBytes looks for the
// "stream" keyword.
func (f *pdfFile) parseDict(buf []byte, at, depth int) (pdfValue, int) {
	at += 2 // '<<'
	dict := map[string]pdfValue{}
	for at < len(buf) {
		at = skipSpace(buf, at)
		if at+1 < len(buf) && buf[at] == '>' && buf[at+1] == '>' {
			at += 2
			break
		}
		if at >= len(buf) {
			break
		}
		if buf[at] != '/' {
			// Skip what should not be here rather than failing the object.
			_, next := f.parse(buf, at, depth+1)
			if next == at {
				break
			}
			at = next
			continue
		}
		key, next := f.parseName(buf, at)
		at = next
		val, next := f.parse(buf, at, depth+1)
		if next == at {
			break
		}
		at = next
		dict[key.name] = val
	}
	d := pdfValue{kind: pdfDict, dict: dict}
	if len(buf) > 0 && len(f.data) >= len(buf) && &buf[0] == &f.data[0] {
		if next := skipSpace(f.data, at); bytes.HasPrefix(f.data[next:], []byte("stream")) {
			d.num = float64(at)
		}
	}
	return d, at
}

func parseKeyword(buf []byte, at int) (pdfValue, int) {
	start := at
	for at < len(buf) && !isDelim(buf[at]) && !isSpace(buf[at]) {
		at++
	}
	switch word := string(buf[start:at]); word {
	case "true":
		return pdfValue{kind: pdfBool, num: 1}, at
	case "false":
		return pdfValue{kind: pdfBool}, at
	case "null":
		return pdfValue{kind: pdfNull}, at
	default:
		// Any other bare word is an operator in a content stream, or a keyword
		// such as R or obj in the file's own syntax.
		return pdfValue{kind: pdfName, name: word}, at
	}
}

// parseObjectAt reads one complete object and returns its end offset.
func (f *pdfFile) parseObjectAt(at int) (pdfValue, int) {
	return f.parse(f.data, at, 0)
}

func isSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\r', '\n', '\f', 0:
		return true
	}
	return false
}

func isDelim(c byte) bool {
	switch c {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	}
	return false
}


