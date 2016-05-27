package mail

import (
	"io"
	"mime/multipart"
	"net/textproto"
)

// EncodingMultipartWriter contains headers and information needed to write a MIME Multipart message
type EncodingMultipartWriter struct {
	out             io.Writer
	headers         map[string]string
	headersWritten  bool
	multipartWriter *multipart.Writer
}

const newline = "\r\n"

// NewEncodingMultipartWriter returns a new encoding multipart writer of the given type writing to w.
// The multitype can be for example "mixed", "alternative", or "encrypted".
// Optionally, a protocol for the Content-Type header can be given, for example "application/pgp-encrypted"
// The writer defines the headers "Mime-Version", "Content-Type" and "Content-Transfer-Encoding".
// Additional headers can be passed via the extraHeaders field. The extraHeaders field takes precedence.
func NewEncodingMultipartWriter(w io.Writer, multitype, protocol string, extraHeaders map[string]string) *EncodingMultipartWriter {
	multipartWriter := multipart.NewWriter(w)
	contentType := "multipart/" + multitype + ";" + newline +
		" boundary=\"" + multipartWriter.Boundary() + "\""
	if len(protocol) > 0 {
		contentType += ";" + newline +
			" protocol=\"" + protocol + "\";"
	}
	headers := map[string]string{
		"Mime-Version":              "1.0",
		"Content-Type":              contentType,
		"Content-Transfer-Encoding": "7bit",
	}
	if extraHeaders != nil {
		for key, value := range extraHeaders {
			headers[key] = value
		}
	}
	return &EncodingMultipartWriter{w, headers, false, multipartWriter}
}

func (w *EncodingMultipartWriter) checkWriteHeaders() error {
	if w.headersWritten {
		return nil
	}
	w.headersWritten = true
	for key, value := range w.headers {
		// TODO: Ensure that headers are 7bit
		_, err := w.out.Write([]byte(key + ": " + value + newline))
		if err != nil {
			return err
		}
	}
	_, err := w.out.Write([]byte(newline))
	return err
}

// WritePGPMIMEVersion writes a multipart containing PGP Version information
func (w *EncodingMultipartWriter) WritePGPMIMEVersion() error {
	err := w.checkWriteHeaders()
	if err != nil {
		return err
	}
	partWriter, err := w.multipartWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"application/pgp-encrypted"},
		"Content-Description": {"PGP/MIME version identification"},
	})
	if err != nil {
		return err
	}
	_, err = partWriter.Write([]byte("Version: 1" + newline))
	return err
}

func (w *EncodingMultipartWriter) writeFile(name, mimeType, description, disposition string) (io.Writer, error) {
	err := w.checkWriteHeaders()
	if err != nil {
		return nil, err
	}
	return w.multipartWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {mimeType + "; name=\"" + name + "\""},
		"Content-Disposition": {disposition + "; filename=\"" + name + "\""},
		"Content-Description": {description},
	})
}

// WriteInlineFile writes a multipart header for an inline file and provides a writer for the file contents
func (w *EncodingMultipartWriter) WriteInlineFile(name, mimeType, description string) (io.Writer, error) {
	return w.writeFile(name, mimeType, description, "inline")
}

// WriteAttachedFile writes a multipart header for an attached file and provides a writer for the file contents
func (w *EncodingMultipartWriter) WriteAttachedFile(name, mimeType, description string) (io.Writer, error) {
	return w.writeFile(name, mimeType, description, "attachment")
}

// WritePlainText writes the given text as a text/plain segment
func (w *EncodingMultipartWriter) WritePlainText(text string) error {
	err := w.checkWriteHeaders()
	if err != nil {
		return err
	}
	partWriter, err := w.multipartWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"text/plain"},
	})
	if err != nil {
		return err
	}
	_, err = partWriter.Write([]byte(text))
	return err
}

// Close writes the trailing information for an multipartWriter
func (w *EncodingMultipartWriter) Close() error {
	err := w.checkWriteHeaders()
	if err != nil {
		return err
	}
	return w.multipartWriter.Close()
}
