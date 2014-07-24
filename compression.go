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
