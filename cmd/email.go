package cmd

import (
	"bytes"
	"fmt"
	"mime/quotedprintable"
	"strings"
	"time"
)

func formatEmailMessage(messageID, subject, to, from string, body []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString(fmt.Sprintf("X-Filterctl-Request-ID: <%s>\r\n", strings.Trim(messageID, "<>")))
	buf.WriteString("Content-Type: text/plain; charset=\"us-ascii\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buf.WriteString("\r\n")
	writer := quotedprintable.NewWriter(&buf)
	_, err := writer.Write(body)
	if err != nil {
		return nil, err
	}
	writer.Close()
	return buf.Bytes(), nil
}
