package websocket

import (
	"compress/flate"
	"io"
)

//// if we've negotiated the permessage-deflate extension,
//// then the whole message is compressed
//for _, ext := range c.extensions {
//    if ext.Token == "permessage-deflate" {
//    if _, no_client_takeover := ext.Params["client_no_context_takeover"]; no_client_takeover {

// we need to take over context
//        compressed_reader := c.clientFlateReader
//        if compressed_reader == nil {
//            compressed_reader = flate.NewReader(msg_reader)
//            c.clientFlateReader = compressed_reader
//        } else {
//            compressed_reader.Reader = msg_reader
//        }
//    }
//}

func ReadPerMessageDeflateNoTakeover(c *Conn, inputMessageType int, r io.Reader) (messageType int, modifiedReader io.Reader, err error) {
	if c.readCompressed {
		compressed_reader := flate.NewReader(r)
		return inputMessageType, compressed_reader, nil
	} else {
		return inputMessageType, r, nil
	}
}

type redirectableReader struct {
	io.Reader
}

// This doesn't work because of current limitations in the go deflate library
type redirectableDecompressor struct {
	flate_reader        io.ReadCloser
	redirectable_reader *redirectableReader
}

func (rdec redirectableDecompressor) Read(p []byte) (n int, err error) {
	return rdec.flate_reader.Read(p)
}

func (rdec redirectableDecompressor) Redirect(new_reader io.Reader) {
	rdec.redirectable_reader.Reader = new_reader
}

func newRedirectableDecompressor(r io.Reader) redirectableDecompressor {
	redirectable_reader := redirectableReader{r}
	return redirectableDecompressor{
		flate_reader:        flate.NewReader(redirectable_reader),
		redirectable_reader: &redirectable_reader,
	}
}

func ReadPerMessageDeflateWithTakeover(c *Conn, inputMessageType int, r io.Reader) (messageType int, modifiedReader io.Reader, err error) {
	if c.decompressor == nil {
		c.decompressor = newRedirectableDecompressor(r)
	} else {
		decompressor, is_redirectable := c.decompressor.(redirectableDecompressor)
		if is_redirectable {
			decompressor.Redirect(r)
		} else {
			c.decompressor = newRedirectableDecompressor(r)
		}
	}

	if c.readCompressed {
		compressed_reader := flate.NewReader(r)
		return inputMessageType, compressed_reader, nil
	} else {
		return inputMessageType, r, nil
	}
}

type closableFlateWriter struct {
	original_writer   *io.WriteCloser
	compressed_writer *flate.Writer
}

func (cfw closableFlateWriter) Write(p []byte) (n int, err error) {
	n, err = cfw.compressed_writer.Write(p)
	if err != nil {
		return n, err
	}
	// We need to a zero-byte at the end of the DEFLATE block
	// why? I don't know, but otherwise it doesn't work
	// http://stackoverflow.com/questions/22169036/websocket-permessage-deflate-in-chrome-with-no-context-takeover
	n0, err := (*(cfw.original_writer)).Write([]byte{0})
	return n + n0, err
}

func (cfw closableFlateWriter) Close() (err error) {
	return (*(cfw.original_writer)).Close()
}

func WritePerMessageDeflateNoTakeover(c *Conn, w io.WriteCloser) (modifiedWriter io.WriteCloser, err error) {
	flate_writer, err := flate.NewWriter(w, 8)
	if err != nil {
		return nil, err
	}
	closable_flate_writer := closableFlateWriter{&w, flate_writer}
	return closable_flate_writer, nil
}

func WriterFromExtensions(c *Conn, w io.WriteCloser) (modifiedWriter io.WriteCloser, err error) {
	for _, ext := range c.extensions {
		if ext.Token == "permessage-deflate" {
			if _, no_takeover := ext.Params["server-no-context-takeover"]; no_takeover {
				return WritePerMessageDeflateNoTakeover(c, w)
			}
		}
	}
	return w, nil
}
