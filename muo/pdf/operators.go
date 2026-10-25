package pdf

// Content-stream operators. A token that is one of these ends the operands
// collected before it; every other name is a value, as /F1 is. The set is the
// operators this interpreter implements plus the ones it must consume cleanly
// so that their operands do not leak into the next operator.
var operatorNames = map[string]bool{
	// graphics state
	"q": true, "Q": true, "cm": true, "w": true, "J": true, "j": true,
	"M": true, "d": true, "ri": true, "i": true, "gs": true,
	// color
	"g": true, "G": true, "rg": true, "RG": true, "k": true, "K": true,
	"cs": true, "CS": true, "sc": true, "SC": true, "scn": true, "SCN": true,
	"sh": true,
	// paths
	"m": true, "l": true, "c": true, "v": true, "y": true, "h": true,
	"re": true, "S": true, "s": true, "f": true, "F": true, "f*": true,
	"B": true, "B*": true, "b": true, "b*": true, "n": true, "W": true, "W*": true,
	// text
	"BT": true, "ET": true, "Tf": true, "Tc": true, "Tw": true, "Tz": true,
	"TL": true, "Ts": true, "Tr": true, "Td": true, "TD": true, "Tm": true,
	"T*": true, "Tj": true, "'": true, "\"": true, "TJ": true,
	// xobjects and marked content
	"Do": true, "BMC": true, "BDC": true, "EMC": true, "MP": true, "DP": true,
	// other
	"BX": true, "EX": true, "d0": true, "d1": true,
}

// isOperator reports whether a parsed token is an operator rather than an
// operand. Operators are bare words: names without the leading slash, which
// the parser reads as pdfName too.
func isOperator(v pdfValue) bool {
	return v.kind == pdfName && operatorNames[v.name]
}
