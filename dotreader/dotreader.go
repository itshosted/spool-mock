// Insipired on the dotReader from textproto but way simpler
// to improve performance.
package dotreader

import (
	"bytes"
	"io"
	"fmt"
)

// End Of Stream
var END = []byte("\r\n.\r\n")
var END_SHORT = []byte(".\r\n")

// Ignore EOF
var testNoEOF bool

type DotReader struct {
	r     io.Reader
	begin bool
	done  bool

	buf   []byte
	pos   int
}

func New(r io.Reader) *DotReader {
	return &DotReader{r: r, begin: true}
}

// Just keep on reading until we found that END.
func (d *DotReader) Read(b []byte) (n int, err error) {
	if d.done {
		return 0, io.EOF
	}
	n, e := d.r.Read(b)
	if testNoEOF && e == io.EOF {
		// Ignore EOF in datasource
		e = nil
	}

	// Remember last 5 bytes so we can find EOF
	// if we receive the EOF in parts..
	if n >= 5 {
		d.buf = b[n-5:n]
		d.pos = 5
	} else if n > 0 {
		// TODO: User can OOM when slowly adding data?
		// few bytes
		d.buf = append(d.buf, b[:n]...)
		d.pos += n
	} else {
		// Zero bytes, maybe done?
		fmt.Println("WARN: Read 0 bytes, buggy DotReader?")
	}

	if d.begin && bytes.Index(b[0:len(END_SHORT)], END_SHORT) == 0 {
		fmt.Println("WARN: Short end")
		e = io.EOF
	}
	d.begin = false
	if idx := bytes.Index(d.buf, END); idx != -1 {
		d.done = true
		e = io.EOF
	}
	return n, e
}
