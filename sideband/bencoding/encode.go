package bencoding

import (
	"bytes"
	"errors"
	"strconv"
)

func encodeWriter(item interface{}, buffer *bytes.Buffer) error {
	switch typedItem := item.(type) {
	case string:
		return encodeWriter([]byte(typedItem), buffer)
	case []byte:
		buffer.WriteString(strconv.Itoa(len(typedItem)))
		buffer.WriteByte(':')
		buffer.Write(typedItem)
		return nil
	case int64:
		buffer.WriteByte('i')
		buffer.WriteString(strconv.FormatInt(typedItem, 10))
		buffer.WriteByte('e')
		return nil
	case int:
		return encodeWriter(int64(typedItem), buffer)
	case int32:
		return encodeWriter(int64(typedItem), buffer)
	case int16:
		return encodeWriter(int64(typedItem), buffer)
	case int8:
		return encodeWriter(int64(typedItem), buffer)
	case uint64:
		buffer.WriteByte('i')
		buffer.WriteString(strconv.FormatUint(typedItem, 10))
		buffer.WriteByte('e')
		return nil
	case uint:
		return encodeWriter(uint64(typedItem), buffer)
	case uint32:
		return encodeWriter(uint64(typedItem), buffer)
	case uint16:
		return encodeWriter(uint64(typedItem), buffer)
	case uint8:
		return encodeWriter(uint64(typedItem), buffer)
	case []interface{}:
		buffer.WriteByte('l')
		for _, item := range typedItem {
			err := encodeWriter(item, buffer)
			if err != nil {
				return err
			}
		}
		buffer.WriteByte('e')
		return nil
	case map[string]interface{}:
		buffer.WriteByte('d')
		for key, value := range typedItem {
			err := encodeWriter(key, buffer)
			if err != nil {
				return err
			}
			err = encodeWriter(value, buffer)
			if err != nil {
				return err
			}
		}
		buffer.WriteByte('e')
		return nil
	}

	return errors.New("unknown type")
}

func EncodeString(item interface{}) ([]byte, error) {
	buffer := bytes.Buffer{}
	err := encodeWriter(item, &buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
