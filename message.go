package eventpool

import (
	"bytes"
	"encoding/json"
)

type messageFunc func() ([]byte, error)

func Send(message []byte) messageFunc {
	return func() ([]byte, error) {
		return message, nil
	}
}

func SendString(message string) messageFunc {
	return func() ([]byte, error) {
		return []byte(message), nil
	}
}

func SendJson(message interface{}) messageFunc {
	return func() ([]byte, error) {
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(&message)
		if err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}
}
