package bencoding_test

import (
	"testing"

	"github.com/Doridian/foxTorrent/bencoding"
	"github.com/stretchr/testify/assert"
)

func TestEncodeInteger(t *testing.T) {
	res, err := bencoding.EncodeString(int64(123))
	assert.NoError(t, err)
	assert.Equal(t, []byte("i123e"), res)

	res, err = bencoding.EncodeString(int64(0))
	assert.NoError(t, err)
	assert.Equal(t, []byte("i0e"), res)

	res, err = bencoding.EncodeString(int64(-1234))
	assert.NoError(t, err)
	assert.Equal(t, []byte("i-1234e"), res)

	res, err = bencoding.EncodeString(uint64(1111))
	assert.NoError(t, err)
	assert.Equal(t, []byte("i1111e"), res)

	res, err = bencoding.EncodeString(uint(1111))
	assert.NoError(t, err)
	assert.Equal(t, []byte("i1111e"), res)

	res, err = bencoding.EncodeString(int(-1111))
	assert.NoError(t, err)
	assert.Equal(t, []byte("i-1111e"), res)
}

func TestEncodeString(t *testing.T) {
	res, err := bencoding.EncodeString([]byte("spam"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("4:spam"), res)

	res, err = bencoding.EncodeString([]byte(""))
	assert.NoError(t, err)
	assert.Equal(t, []byte("0:"), res)

	res, err = bencoding.EncodeString("")
	assert.NoError(t, err)
	assert.Equal(t, []byte("0:"), res)

	res, err = bencoding.EncodeString("acme")
	assert.NoError(t, err)
	assert.Equal(t, []byte("4:acme"), res)
}

func TestEncodeList(t *testing.T) {
	res, err := bencoding.EncodeString([]interface{}{[]byte("spam"), int64(123)})
	assert.NoError(t, err)
	assert.Equal(t, []byte("l4:spami123ee"), res)

	res, err = bencoding.EncodeString([]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("le"), res)
}

func TestEncodeDict(t *testing.T) {
	res, err := bencoding.EncodeString(map[string]interface{}{
		"cow":  []byte("moo"),
		"spam": []byte("eggs"),
	})
	assert.NoError(t, err)
	assert.Equal(t, []byte("d3:cow3:moo4:spam4:eggse"), res)

	res, err = bencoding.EncodeString(map[string]interface{}{
		"spam": []interface{}{
			[]byte("a"),
			[]byte("b"),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, []byte("d4:spaml1:a1:bee"), res)

	res, err = bencoding.EncodeString(map[string]interface{}{
		"publisher":          []byte("bob"),
		"publisher-webpage":  []byte("www.example.com"),
		"publisher.location": []byte("home"),
	})
	assert.NoError(t, err)
	assert.Equal(t, []byte("d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee"), res)

	res, err = bencoding.EncodeString(map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("de"), res)

	_, err = bencoding.EncodeString(map[int]interface{}{})
	assert.Error(t, err)

	_, err = bencoding.EncodeString(map[interface{}]interface{}{})
	assert.Error(t, err)
}
