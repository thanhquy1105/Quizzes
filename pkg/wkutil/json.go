package wkutil

import (
	"bytes"
	"encoding/json"
)

func ReadJsonByByte(body []byte, obj interface{}) error {
	mdz := json.NewDecoder(bytes.NewBuffer(body))

	mdz.UseNumber()
	err := mdz.Decode(obj)

	if err != nil {
		return err
	}
	return nil
}
