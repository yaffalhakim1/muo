package muo

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Fixtures for the app's tests. The PDF writer lives here rather than in the pdf
// package so that the app's tests can build a document without the pdf package
// exporting a writer it does not otherwise need.

// writeTemp writes data to a file of the given name in a test's own directory.
func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// simplePDF builds a one-page file whose content is given.
func simplePDF(content string) []byte {
	w := newPDFWriter()
	w.obj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	w.obj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	w.obj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>")
	w.stream(4, "", []byte(content))
	w.obj(5, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /FirstChar 32 /LastChar 126 /Widths ["+helveticaWidths()+"] >>")
	return w.finish(1)
}

// helveticaWidths is a plausible advance table for the standard font: the
// interpreter only needs numbers, but real ones make a wrong sum visible.
func helveticaWidths() string {
	var out bytes.Buffer
	for i := 32; i <= 126; i++ {
		if i > 32 {
			out.WriteString(" ")
		}
		out.WriteString("556")
	}
	return out.String()
}

// newPDFWriter writes a minimal PDF: numbered objects, then a cross-reference
// table and a trailer.
type pdfWriter struct {
	buf     bytes.Buffer
	offsets map[int]int
	next    int
}

func newPDFWriter() *pdfWriter {
	w := &pdfWriter{offsets: map[int]int{}}
	w.buf.WriteString("%PDF-1.7\n")
	return w
}

func (w *pdfWriter) obj(n int, body string) {
	w.offsets[n] = w.buf.Len()
	if n >= w.next {
		w.next = n + 1
	}
	w.buf.WriteString(itoa(n))
	w.buf.WriteString(" 0 obj\n")
	w.buf.WriteString(body)
	w.buf.WriteString("\nendobj\n")
}

func (w *pdfWriter) stream(n int, dict string, data []byte) {
	w.offsets[n] = w.buf.Len()
	if n >= w.next {
		w.next = n + 1
	}
	w.buf.WriteString(itoa(n))
	w.buf.WriteString(" 0 obj\n<< /Length ")
	w.buf.WriteString(itoa(len(data)))
	if dict != "" {
		w.buf.WriteString(" ")
		w.buf.WriteString(dict)
	}
	w.buf.WriteString(" >>\nstream\n")
	w.buf.Write(data)
	w.buf.WriteString("\nendstream\nendobj\n")
}

func (w *pdfWriter) finish(root int) []byte {
	xref := w.buf.Len()
	w.buf.WriteString("xref\n0 ")
	w.buf.WriteString(itoa(w.next))
	w.buf.WriteString("\n0000000000 65535 f \n")
	for i := 1; i < w.next; i++ {
		off, ok := w.offsets[i]
		if !ok {
			w.buf.WriteString("0000000000 65535 f \n")
			continue
		}
		s := itoa(off)
		for len(s) < 10 {
			s = "0" + s
		}
		w.buf.WriteString(s)
		w.buf.WriteString(" 00000 n \n")
	}
	w.buf.WriteString("trailer\n<< /Size ")
	w.buf.WriteString(itoa(w.next))
	w.buf.WriteString(" /Root ")
	w.buf.WriteString(itoa(root))
	w.buf.WriteString(" 0 R >>\nstartxref\n")
	w.buf.WriteString(itoa(xref))
	w.buf.WriteString("\n%%EOF\n")
	return w.buf.Bytes()
}

// itoa is strconv.Itoa without the import, kept here so the fixture reads as one
// piece.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
