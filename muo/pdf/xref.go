package pdf

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// A small PDF reader: enough of ISO 32000 to rasterize pages for a reader.
// It handles what the format's own structures need — xref tables and streams,
// object streams, the Flate and binary ASCII filters, and the encryption flag
// (an encrypted file is refused, not misread) — and leaves the image codecs to
// the rasterizer.
//
// It is not a general PDF library: content it does not understand is skipped
// rather than guessed at, so a page it cannot draw comes out blank instead of
// wrong.

var errPDF = errors.New("muo: unsupported PDF")

// pdfValue is a parsed PDF object of any kind.
type pdfValue struct {
	kind  pdfKind
	num   float64
	name  string
	str   []byte
	array []pdfValue
	dict  map[string]pdfValue
}

type pdfKind uint8

const (
	pdfNull pdfKind = iota
	pdfBool
	pdfNumber
	pdfName
	pdfString
	pdfArray
	pdfDict
	pdfRef
)

func (v pdfValue) is(key string) bool { return v.kind == pdfName && v.name == key }

func (v pdfValue) get(key string) (pdfValue, bool) {
	if v.kind != pdfDict {
		return pdfValue{}, false
	}
	x, ok := v.dict[key]
	return x, ok
}

func (v pdfValue) int() int {
	if v.kind != pdfNumber && v.kind != pdfRef {
		return 0
	}
	return int(v.num)
}

// numList reads an array of numbers, as a rectangle or a matrix.
func (v pdfValue) numList(n int) ([]float64, bool) {
	if v.kind != pdfArray || len(v.array) < n {
		return nil, false
	}
	out := make([]float64, n)
	for i := range out {
		if v.array[i].kind != pdfNumber {
			return nil, false
		}
		out[i] = v.array[i].num
	}
	return out, true
}

// pdfFile is a parsed document. Objects stay unparsed until asked for, so
// opening a large file costs the cross-reference and nothing else.
type pdfFile struct {
	data       []byte
	offsets    map[int]int    // object number → byte offset
	members    map[int][]byte // object-stream members, by object number
	pages      []pdfValue     // page dictionaries, in document order
	hasEncrypt bool
	expanded   bool // object streams are decoded into members
}

// openPDF reads a file's cross-reference, either table or stream.
func openPDF(data []byte) (*pdfFile, error) {
	if len(data) < 8 || !bytes.HasPrefix(data, []byte("%PDF-")) {
		return nil, fmt.Errorf("%w: no header", errPDF)
	}
	f := &pdfFile{
		data:    data,
		offsets: map[int]int{},
		members: map[int][]byte{},
	}
	if err := f.readXref(); err != nil {
		return nil, err
	}
	if err := f.collectPages(); err != nil {
		return nil, err
	}
	return f, nil
}

// readXref finds startxref at the end and follows the chain of sections, an
// older section filling the gaps a newer one left.
func (f *pdfFile) readXref() error {
	tail := f.data
	if len(tail) > 2048 {
		tail = tail[len(tail)-2048:]
	}
	i := bytes.LastIndex(tail, []byte("startxref"))
	if i < 0 {
		return fmt.Errorf("%w: no startxref", errPDF)
	}
	rest := tail[i+len("startxref"):]
	// Only the offset follows; an end-of-file marker or a comment may come
	// after it, so the number ends at the first non-digit.
	start := skipSpace(rest, 0)
	end := start
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	off, err := strconv.Atoi(string(rest[start:end]))
	if err != nil {
		return fmt.Errorf("%w: bad startxref", errPDF)
	}
	seen := map[int]bool{}
	for off > 0 && off < len(f.data) && !seen[off] {
		seen[off] = true
		next, err := f.readXrefSection(off)
		if err != nil {
			return err
		}
		off = next
	}
	if len(f.offsets) == 0 {
		return fmt.Errorf("%w: empty xref", errPDF)
	}
	return nil
}

// readXrefSection reads one section and returns the offset of the previous
// one, which 0 ends.
func (f *pdfFile) readXrefSection(off int) (int, error) {
	at := skipSpace(f.data, off)
	if bytes.HasPrefix(f.data[at:], []byte("xref")) {
		return f.readXrefTable(at + len("xref"))
	}
	// The offset points at "n g obj", the header of the stream's own object.
	d, err := f.readIndirectAt(at)
	if err != nil {
		return 0, fmt.Errorf("%w: xref object", errPDF)
	}
	if t, _ := d.get("Type"); !t.is("XRef") {
		return 0, fmt.Errorf("%w: no xref section at %d", errPDF, off)
	}
	raw, err := f.decodeStream(d)
	if err != nil {
		return 0, err
	}
	prev := mustGet(d, "Prev").int()
	if err := f.readXrefStream(d, raw); err != nil {
		return 0, err
	}
	return prev, nil
}

// readXrefTable reads the classic "n m" subsections and returns the trailer's
// /Prev, if any.
func (f *pdfFile) readXrefTable(at int) (int, error) {
	for {
		at = skipSpace(f.data, at)
		if at >= len(f.data) {
			return 0, fmt.Errorf("%w: truncated xref", errPDF)
		}
		if bytes.HasPrefix(f.data[at:], []byte("trailer")) {
			trailer := f.parseValue(at + len("trailer"))
			if _, ok := trailer.get("Encrypt"); ok {
				f.hasEncrypt = true
			}
			return mustGet(trailer, "Prev").int(), nil
		}
		start, next, err := readInt(f.data, at)
		if err != nil {
			return 0, err
		}
		count, next, err := readInt(f.data, next)
		if err != nil {
			return 0, err
		}
		at = next
		for i := 0; i < count; i++ {
			var entry int
			if entry, at, err = readInt(f.data, at); err != nil {
				return 0, err
			}
			// The generation number follows; it only matters to writers.
			if _, at, err = readInt(f.data, at); err != nil {
				return 0, err
			}
			at = skipSpace(f.data, at)
			if at >= len(f.data) {
				return 0, fmt.Errorf("%w: truncated xref", errPDF)
			}
			kind := f.data[at]
			if n := bytes.IndexByte(f.data[at:], '\n'); n >= 0 {
				at += n + 1
			} else {
				at = len(f.data)
			}
			if kind != 'n' || entry <= 0 {
				continue
			}
			if _, ok := f.offsets[start+i]; !ok {
				f.offsets[start+i] = entry
			}
		}
	}
}

// readXrefStream fills the table from a /Type /XRef stream.
func (f *pdfFile) readXrefStream(d pdfValue, raw []byte) error {
	w, ok := mustGet(d, "W").numList(3)
	if !ok {
		return fmt.Errorf("%w: xref stream widths", errPDF)
	}
	size := mustInt(d, "Size")
	index := []float64{0, float64(size)}
	if v, ok := d.get("Index"); ok {
		if l, ok := v.numList(len(v.array)); ok {
			index = l
		}
	}
	w0, w1, w2 := int(w[0]), int(w[1]), int(w[2])
	row := w0 + w1 + w2
	if row <= 0 {
		return fmt.Errorf("%w: xref stream rows", errPDF)
	}
	field := func(at, n int) int {
		v := 0
		for k := 0; k < n; k++ {
			v = v<<8 | int(raw[at+k])
		}
		return v
	}
	at := 0
	for i := 0; i+1 < len(index); i += 2 {
		start, count := int(index[i]), int(index[i+1])
		for j := 0; j < count; j++ {
			if at+row > len(raw) {
				return nil
			}
			// Entry type 1 is a plain object; types 0 and 2 are free and
			// compressed, which need no offset.
			kind := 1
			if w0 > 0 {
				kind = field(at, w0)
			}
			off := field(at+w0, w1)
			_ = w2
			if num := start + j; kind == 1 && off > 0 {
				if _, seen := f.offsets[num]; !seen {
					f.offsets[num] = off
				}
			}
			at += row
		}
	}
	return nil
}

// encrypted reports whether the trailer names an encryption dictionary. The
// flag alone is enough: the reader refuses the file rather than showing
// ciphertext as text.
func (f *pdfFile) encrypted() bool { return f.hasEncrypt }

// object returns an object by number, expanding an object stream on demand.
func (f *pdfFile) object(num int) (pdfValue, error) {
	if raw, ok := f.members[num]; ok {
		return f.parseMember(raw), nil
	}
	off, ok := f.offsets[num]
	if !ok {
		return pdfValue{}, fmt.Errorf("%w: no object %d", errPDF, num)
	}
	v, err := f.readIndirectAt(off)
	if err != nil {
		return pdfValue{}, err
	}
	if t, _ := v.get("Type"); t.is("ObjStm") {
		// The member is found by scanning every object stream once.
		f.expandObjectStreams()
		if raw, ok := f.members[num]; ok {
			return f.parseMember(raw), nil
		}
	}
	return v, nil
}

// expandObjectStreams decodes the object streams in the file once, so that any
// member can be found by number afterwards.
func (f *pdfFile) expandObjectStreams() {
	if f.expanded {
		return
	}
	f.expanded = true
	for _, off := range f.offsets {
		d, err := f.readIndirectAt(off)
		if err != nil {
			continue
		}
		if t, _ := d.get("Type"); !t.is("ObjStm") {
			continue
		}
		raw, err := f.decodeStream(d)
		if err != nil {
			continue
		}
		f.expandObjectStream(d, raw)
	}
}

// readIndirectAt parses "n g obj … endobj" at a byte offset.
func (f *pdfFile) readIndirectAt(at int) (pdfValue, error) {
	at = skipSpace(f.data, at)
	_, next, err := readInt(f.data, at)
	if err != nil {
		return pdfValue{}, err
	}
	if _, next, err = readInt(f.data, next); err != nil {
		return pdfValue{}, err
	}
	at = skipSpace(f.data, next)
	if !bytes.HasPrefix(f.data[at:], []byte("obj")) {
		return pdfValue{}, fmt.Errorf("%w: object header", errPDF)
	}
	return f.parseValue(at + len("obj")), nil
}

// expandObjectStream caches the members of an object stream, so that repeated
// lookups do not decode it again.
func (f *pdfFile) expandObjectStream(d pdfValue, raw []byte) {
	n, first := mustInt(d, "N"), mustInt(d, "First")
	type span struct{ start, end int }
	spans := make([]span, 0, n)
	nums := make([]int, 0, n)
	at := 0
	for i := 0; i < n; i++ {
		num, next, err := readInt(raw, at)
		if err != nil {
			break
		}
		rel, next, err := readInt(raw, next)
		if err != nil {
			break
		}
		at = next
		nums = append(nums, num)
		spans = append(spans, span{start: first + rel, end: len(raw)})
	}
	for i := range spans {
		if i+1 < len(spans) && spans[i+1].start > spans[i].start {
			spans[i].end = spans[i+1].start
		}
		if spans[i].start >= len(raw) || spans[i].end > len(raw) || spans[i].end < spans[i].start {
			continue
		}
		if _, ok := f.members[nums[i]]; !ok {
			f.members[nums[i]] = raw[spans[i].start:spans[i].end]
		}
	}
}

// resolve follows a reference; a value that is not one comes back as is.
func (f *pdfFile) resolve(v pdfValue) pdfValue {
	for i := 0; i < 8 && v.kind == pdfRef; i++ {
		next, err := f.object(v.int())
		if err != nil {
			return pdfValue{}
		}
		v = next
	}
	return v
}

// decodeStream returns a stream's bytes after its filters.
func (f *pdfFile) decodeStream(d pdfValue) ([]byte, error) {
	length := f.resolve(mustGet(d, "Length")).int()
	raw := f.streamBytes(d, length)
	if raw == nil {
		return nil, fmt.Errorf("%w: no stream", errPDF)
	}
	out := raw
	for _, name := range f.filterList(d) {
		var err error
		if out, err = decodeFilter(name, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// streamBytes finds the payload after "stream" and the EOL that follows it.
// The dictionary leaves the offset it ended at in num, and the keyword is the
// first thing after it.
func (f *pdfFile) streamBytes(d pdfValue, length int) []byte {
	start := int(d.num)
	if start <= 0 || start >= len(f.data) {
		return nil
	}
	i := bytes.Index(f.data[start:], []byte("stream"))
	if i < 0 || i > 8 {
		return nil
	}
	at := start + i + len("stream")
	if at < len(f.data) && f.data[at] == '\r' {
		at++
	}
	if at < len(f.data) && f.data[at] == '\n' {
		at++
	}
	if length > 0 && at+length <= len(f.data) {
		return f.data[at : at+length]
	}
	// A stream whose length is wrong or indirect ends at endstream.
	end := bytes.Index(f.data[at:], []byte("endstream"))
	if end < 0 {
		return nil
	}
	return f.data[at : at+end]
}

// filterList flattens /Filter into a list of names.
func (f *pdfFile) filterList(d pdfValue) []string {
	v := f.resolve(mustGet(d, "Filter"))
	switch v.kind {
	case pdfName:
		return []string{v.name}
	case pdfArray:
		out := make([]string, 0, len(v.array))
		for _, e := range v.array {
			if e = f.resolve(e); e.kind == pdfName {
				out = append(out, e.name)
			}
		}
		return out
	}
	return nil
}

func decodeFilter(name string, in []byte) ([]byte, error) {
	switch name {
	case "FlateDecode", "Fl":
		if r, err := zlib.NewReader(bytes.NewReader(in)); err == nil {
			defer r.Close()
			out, err := io.ReadAll(r)
			if err == nil {
				return out, nil
			}
		}
		// Some writers emit raw deflate; flate reads it directly.
		if out, err := io.ReadAll(flate.NewReader(bytes.NewReader(in))); err == nil {
			return out, nil
		}
		return nil, fmt.Errorf("%w: corrupted Flate stream", errPDF)
	case "RunLengthDecode", "RL":
		return runLengthDecode(in), nil
	case "ASCIIHexDecode", "AHx":
		return asciiHexDecode(in), nil
	case "ASCII85Decode", "A85":
		return ascii85Decode(in), nil
	case "LZWDecode", "LZW", "JBIG2Decode", "CCITTFaxDecode", "DCTDecode", "JPXDecode":
		// Image codecs are the rasterizer's business; JS is never run.
		return in, nil
	}
	return in, nil
}

func runLengthDecode(in []byte) []byte {
	out := make([]byte, 0, len(in)*2)
	for i := 0; i < len(in); {
		n := int(in[i])
		i++
		switch {
		case n == 128:
			return out
		case n < 128:
			end := i + n + 1
			if end > len(in) {
				end = len(in)
			}
			out = append(out, in[i:end]...)
			i = end
		default:
			if i >= len(in) {
				return out
			}
			for k := 0; k < 257-n; k++ {
				out = append(out, in[i])
			}
			i++
		}
	}
	return out
}

func asciiHexDecode(in []byte) []byte {
	out := make([]byte, 0, len(in)/2)
	var half byte
	odd := false
	for _, c := range in {
		if c == '>' {
			break
		}
		v := hexVal(c)
		if v < 0 {
			continue
		}
		if !odd {
			half = byte(v << 4)
			odd = true
			continue
		}
		out = append(out, half|byte(v))
		odd = false
	}
	if odd {
		out = append(out, half)
	}
	return out
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func ascii85Decode(in []byte) []byte {
	out := make([]byte, 0, len(in))
	var acc uint32
	n := 0
	for i := 0; i < len(in); i++ {
		c := in[i]
		if c == '~' {
			break
		}
		if c == 'z' && n == 0 {
			out = append(out, 0, 0, 0, 0)
			continue
		}
		if c <= ' ' || c < '!' || c > 'u' {
			continue
		}
		acc = acc*85 + uint32(c-'!')
		n++
		if n == 5 {
			out = append(out, byte(acc>>24), byte(acc>>16), byte(acc>>8), byte(acc))
			acc, n = 0, 0
		}
	}
	if n > 1 {
		for i := n; i < 5; i++ {
			acc = acc*85 + 84
		}
		out = append(out, byte(acc>>24), byte(acc>>16), byte(acc>>8), byte(acc))
	}
	return out
}

// helpers -------------------------------------------------------------------

func mustGet(d pdfValue, key string) pdfValue {
	v, _ := d.get(key)
	return v
}

func mustInt(d pdfValue, key string) int { return mustGet(d, key).int() }

func skipSpace(b []byte, at int) int {
	for at < len(b) && isSpace(b[at]) {
		at++
	}
	return at
}

func readInt(b []byte, at int) (int, int, error) {
	at = skipSpace(b, at)
	start := at
	if at < len(b) && (b[at] == '-' || b[at] == '+') {
		at++
	}
	for at < len(b) && b[at] >= '0' && b[at] <= '9' {
		at++
	}
	if start == at {
		return 0, at, fmt.Errorf("%w: integer", errPDF)
	}
	v, err := strconv.Atoi(string(b[start:at]))
	return v, at, err
}

// parseMember reads a decoded object-stream member. Object streams never hold
// streams, so there is no keyword to look for and no dictionary offset.
func (f *pdfFile) parseMember(raw []byte) pdfValue {
	v, _ := f.parse(raw, 0, 0)
	return v
}









// float returns a number's value, or zero for anything that is not a number.
// The interpreter asks this of dictionary entries that may be missing or of
// another kind, so it answers rather than failing.
func (v pdfValue) float() float64 {
	if v.kind != pdfNumber {
		return 0
	}
	return v.num
}