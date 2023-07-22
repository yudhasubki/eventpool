package eventpool

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

type messageFunc func() (io.Reader, error)

func Send(message io.Reader) messageFunc {
	return func() (io.Reader, error) {
		return message, nil
	}
}

func SendString(message string) messageFunc {
	return func() (io.Reader, error) {
		return strings.NewReader(message), nil
	}
}

func SendByte(message []byte) messageFunc {
	return func() (io.Reader, error) {
		return bytes.NewReader(message), nil
	}
}

func SendJson(message interface{}) messageFunc {
	return func() (io.Reader, error) {
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(&message)
		if err != nil {
			return nil, err
		}

		return &buf, nil
	}
}
