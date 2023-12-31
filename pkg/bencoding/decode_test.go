package bencoding_test

import (
	"testing"

	"github.com/Doridian/foxTorrent/pkg/bencoding"
	"github.com/stretchr/testify/assert"
)

func TestDecodeInteger(t *testing.T) {
	res, err := bencoding.DecodeString([]byte("i100003e"))
	assert.NoError(t, err)
	assert.Equal(t, int64(100003), res)

	res, err = bencoding.DecodeString([]byte("i0e"))
	assert.NoError(t, err)
	assert.Equal(t, int64(0), res)

	res, err = bencoding.DecodeString([]byte("i-1004e"))
	assert.NoError(t, err)
	assert.Equal(t, int64(-1004), res)

	_, err = bencoding.DecodeString([]byte("ie"))
	assert.Error(t, err)

	_, err = bencoding.DecodeString([]byte("i-e"))
	assert.Error(t, err)

	_, err = bencoding.DecodeString([]byte("i1-2e"))
	assert.Error(t, err)

	_, err = bencoding.DecodeString([]byte("i-0e"))
	assert.Error(t, err)

	_, err = bencoding.DecodeString([]byte("i02e"))
	assert.Error(t, err)
}

func TestDecodeString(t *testing.T) {
	res, err := bencoding.DecodeString([]byte("4:spam"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("spam"), res)

	res, err = bencoding.DecodeString([]byte("0:"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), res)

	_, err = bencoding.DecodeString([]byte("4spam"))
	assert.Error(t, err)

	_, err = bencoding.DecodeString([]byte("99:tooshort"))
	assert.Error(t, err)

	_, err = bencoding.DecodeString([]byte("-1:aaaaaaaaaaa"))
	assert.Error(t, err)

	_, err = bencoding.DecodeString([]byte("-0:aaaaaaaaaaa"))
	assert.Error(t, err)
}

func TestDecodeList(t *testing.T) {
	res, err := bencoding.DecodeString([]byte("l4:spami123ee"))
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{[]byte("spam"), int64(123)}, res)

	res, err = bencoding.DecodeString([]byte("le"))
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{}, res)

	_, err = bencoding.DecodeString([]byte("l"))
	assert.Error(t, err)
}

func TestDecodeDict(t *testing.T) {
	res, err := bencoding.DecodeString([]byte("d3:cow3:moo4:spam4:eggse"))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"cow":      []byte("moo"),
		"spam":     []byte("eggs"),
		"$$meta$$": bencoding.DictMeta{Begin: 0, End: 24},
	}, res)

	res, err = bencoding.DecodeString([]byte("d4:spaml1:a1:bee"))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"spam": []interface{}{
			[]byte("a"),
			[]byte("b"),
		},
		"$$meta$$": bencoding.DictMeta{Begin: 0, End: 16},
	}, res)

	res, err = bencoding.DecodeString([]byte("d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee"))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"publisher":          []byte("bob"),
		"publisher-webpage":  []byte("www.example.com"),
		"publisher.location": []byte("home"),
		"$$meta$$":           bencoding.DictMeta{Begin: 0, End: 83},
	}, res)

	res, err = bencoding.DecodeString([]byte("de"))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"$$meta$$": bencoding.DictMeta{Begin: 0, End: 2},
	}, res)

	_, err = bencoding.DecodeString([]byte("d"))
	assert.Error(t, err)
}
