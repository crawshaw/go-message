package message

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"

	"github.com/emersion/go-textwrapper"
)

func encodingReader(enc string, r io.Reader) (io.Reader, error) {
	var dec io.Reader
	switch strings.ToLower(enc) {
	case "quoted-printable":
		dec = quotedprintable.NewReader(r)
	case "base64":
		dec = base64.NewDecoder(base64.StdEncoding, r)
	case "7bit", "8bit", "binary", "":
		dec = r
	default:
		return nil, fmt.Errorf("unhandled encoding %q", enc)
	}
	return dec, nil
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func encodingWriter(enc string, w io.Writer) io.WriteCloser {
	var wc io.WriteCloser
	switch strings.ToLower(enc) {
	case "quoted-printable":
		wc = quotedprintable.NewWriter(w)
	case "base64":
		wc = base64.NewEncoder(base64.StdEncoding, textwrapper.NewRFC822(w))
	case "7bit", "8bit":
		wc = nopCloser{textwrapper.New(w, "\r\n", 1000)}
	default: // "binary"
		wc = nopCloser{w}
	}
	return wc
}

// EncodedSize returns the transfer-encoded size of the contents of r.
//
// The exact transfer-encoding size is not fixed by any standard.
// For example, the choice of column number to wrap lines in a
// base64 encoded part can change the total number of bytes.
//
// This function measures by performing the encoding, which is not
// particularly efficient but is efficient-enough and easy to verify.
func EncodedSize(transferEncoding string, r io.Reader) (int64, error) {
	lw := &lenWriter{}
	w := encodingWriter(transferEncoding, lw)
	if _, err := io.Copy(w, r); err != nil {
		return lw.n, err
	}
	if err := w.Close(); err != nil {
		return lw.n, err
	}
	return lw.n, nil
}

type lenWriter struct{ n int64 }

func (w *lenWriter) Write(p []byte) (n int, err error) {
	w.n += int64(len(p))
	return len(p), nil
}
