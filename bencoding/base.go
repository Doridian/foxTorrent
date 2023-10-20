package bencoding

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func encodeSB(item interface{}, builder *strings.Builder) error {
	switch typedItem := item.(type) {
	case string:
		builder.WriteString(strconv.Itoa(len(typedItem)))
		builder.WriteByte(':')
		builder.WriteString(typedItem)
	case int64:
		builder.WriteByte('i')
		builder.WriteString(strconv.FormatInt(typedItem, 10))
		builder.WriteByte('e')
		return nil
	case int:
		return encodeSB(int64(typedItem), builder)
	case int32:
		return encodeSB(int64(typedItem), builder)
	case int16:
		return encodeSB(int64(typedItem), builder)
	case int8:
		return encodeSB(int64(typedItem), builder)
	case uint64:
		builder.WriteByte('i')
		builder.WriteString(strconv.FormatUint(typedItem, 10))
		builder.WriteByte('e')
		return nil
	case uint:
		return encodeSB(uint64(typedItem), builder)
	case uint32:
		return encodeSB(uint64(typedItem), builder)
	case uint16:
		return encodeSB(uint64(typedItem), builder)
	case uint8:
		return encodeSB(uint64(typedItem), builder)
	case []interface{}:
		builder.WriteByte('l')
		for _, item := range typedItem {
			err := encodeSB(item, builder)
			if err != nil {
				return err
			}
		}
		builder.WriteByte('e')
		return nil
	case map[string]interface{}:
		builder.WriteByte('d')
		for key, value := range typedItem {
			err := encodeSB(key, builder)
			if err != nil {
				return err
			}
			err = encodeSB(value, builder)
			if err != nil {
				return err
			}
		}
		builder.WriteByte('e')
		return nil
	}

	return errors.New("unknown type")
}

func Encode(item interface{}) (string, error) {
	sb := strings.Builder{}
	err := encodeSB(item, &sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func readNumeric(reader *strings.Reader, terminator byte) (int64, error) {
	var res int64 = 0
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		if b < '0' || b > '9' {
			if b != terminator {
				return 0, errors.New("invalid numeric")
			}
			return res, nil
		}

		res = res*10 + int64(b-'0')
	}
}

func newDecoderError(reader *strings.Reader, err string) error {
	pos, _ := reader.Seek(0, io.SeekCurrent)
	return fmt.Errorf("%s at %d", err, pos)
}

var ErrEndOfItem = errors.New("end-of-item")

func decodeReader(reader *strings.Reader) (interface{}, error) {
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, nil
			}
			return nil, err
		}

		if b == 'i' { // integer
			return readNumeric(reader, 'e')
		} else if b == 'l' { // list
			res := make([]interface{}, 0)
			for {
				item, err := decodeReader(reader)
				if err != nil {
					return nil, err
				}
				if item == nil {
					return nil, newDecoderError(reader, "encountered EOF trying to decode item")
				}
				if item == ErrEndOfItem {
					return res, nil
				}
				res = append(res, item)
			}
		} else if b == 'd' { // dict string => any
			res := make(map[string]interface{})
			for {
				key, err := decodeReader(reader)
				if err != nil {
					return nil, err
				}
				if key == nil {
					return nil, newDecoderError(reader, "encountered EOF trying to decode key")
				}
				if key == ErrEndOfItem {
					return res, nil
				}
				keyStr, ok := key.(string)
				if !ok {
					return nil, newDecoderError(reader, fmt.Sprintf("dict key \"%v\" is not a string", key))
				}
				value, err := decodeReader(reader)
				if err != nil {
					return nil, err
				}
				if value == nil {
					return nil, newDecoderError(reader, "encountered EOF trying to decode value")
				}
				if value == ErrEndOfItem {
					return nil, newDecoderError(reader, "encountered end-of-item trying to decode value")
				}
				res[keyStr] = value
			}
		} else if b >= '0' && b <= '9' { // string
			err = reader.UnreadByte()
			if err != nil {
				return nil, err
			}
			length, err := readNumeric(reader, ':')
			if err != nil {
				return nil, err
			}
			buf := make([]byte, length)
			readLength, err := reader.Read(buf)
			if err != nil {
				return nil, err
			}
			if readLength != int(length) {
				return nil, newDecoderError(reader, fmt.Sprintf("encountered short read while reading string (got %d, expected %d)", readLength, length))
			}
			return string(buf), nil
		} else if b == 'e' { // end-of-item
			return ErrEndOfItem, nil
		} else {
			return nil, newDecoderError(reader, fmt.Sprintf("encountered unexpected \"%c\" while trying to read next element", b))
		}
	}
}

func Decode(input string) (interface{}, error) {
	reader := strings.NewReader(input)
	obj, err := decodeReader(reader)
	if obj == ErrEndOfItem {
		return nil, errors.New("unexpected end-of-item")
	}
	return obj, err
}
