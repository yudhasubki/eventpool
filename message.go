package eventpool

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

type MessageFunc func() (io.Reader, error)

func Send(message io.Reader) MessageFunc {
	return func() (io.Reader, error) {
		return message, nil
	}
}

func SendString(message string) MessageFunc {
	return func() (io.Reader, error) {
		return strings.NewReader(message), nil
	}
}

func SendByte(message []byte) MessageFunc {
	return func() (io.Reader, error) {
		return bytes.NewReader(message), nil
	}
}

func SendJson(message interface{}) MessageFunc {
	return func() (io.Reader, error) {
		var buf *bytes.Buffer
		err := json.NewEncoder(buf).Encode(message)
		if err != nil {
			return nil, err
		}

		return buf, nil
	}
}
