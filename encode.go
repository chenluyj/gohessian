package gohessian

import (
	"bytes"
	"github.com/chenluyj/gohessian/util"
	"log"
	"runtime"
	"time"
	"unicode/utf8"
)

/*
对 基本数据进行 Hessian 编码
支持:
int8 int16 int32 int64
float64
time.Time
[]byte
[]interface{}
map[interface{}]interface{}
nil
bool
*/

type Encoder struct {
}

const (
	CHUNK_SIZE    = 0x8000
	ENCODER_DEBUG = false
)

func init() {
	_, filename, _, _ := runtime.Caller(1)
	if ENCODER_DEBUG {
		log.SetPrefix(filename + "\n")
	}
}

func Encode(v interface{}) (b []byte, err error) {

	switch v.(type) {

	case []byte:
		b, err = encodeBinary(v.([]byte))

	case bool:
		b, err = encodeBool(v.(bool))

	case time.Time:
		b, err = encodeTime(v.(time.Time))

	case float64:
		b, err = encodeFloat64(v.(float64))

	case int:
		if v.(int) >= -2147483648 && v.(int) <= 2147483647 {
			b, err = encodeInt32(int32(v.(int)))
		} else {
			b, err = encodeInt64(int64(v.(int)))
		}

	case int32:
		b, err = encodeInt32(v.(int32))

	case int64:
		b, err = encodeInt64(v.(int64))

	case string:
		b, err = encodeString(v.(string))

	case nil:
		b, err = encodeNull(v)

	case []Any:
		b, err = encodeList(v.([]Any))

	case map[Any]Any:
		b, err = encodeMap(v.(map[Any]Any))

	default:
		panic("unknow type")
	}
	if ENCODER_DEBUG {
		log.Println(util.SprintHex(b))
	}
	return
}

//=====================================
//对各种数据类型的编码
//=====================================

// binary
func encodeBinary(v []byte) (b []byte, err error) {
	var (
		tag  byte
		lenB []byte
		lenN int
	)

	if len(v) == 0 {
		if lenB, err = util.PackUint16(0); err != nil {
			b = nil
			return
		}
		b = append(b, 'B')
		b = append(b, lenB...)
		return
	}

	rBuf := *bytes.NewBuffer(v)

	for rBuf.Len() > 0 {
		if rBuf.Len() > CHUNK_SIZE {
			tag = 'b'
			if lenB, err = util.PackUint16(uint16(CHUNK_SIZE)); err != nil {
				b = nil
				return
			}
			lenN = CHUNK_SIZE
		} else {
			tag = 'B'
			if lenB, err = util.PackUint16(uint16(rBuf.Len())); err != nil {
				b = nil
				return
			}
			lenN = rBuf.Len()
		}
		b = append(b, tag)
		b = append(b, lenB...)
		b = append(b, rBuf.Next(lenN)...)
	}
	return
}

// boolean
func encodeBool(v bool) (b []byte, err error) {
	if v == true {
		b = append(b, 'T')
	} else {
		b = append(b, 'F')
	}
	return
}

// date
func encodeTime(v time.Time) (b []byte, err error) {
	var tmpV []byte
	b = append(b, 'd')
	if tmpV, err = util.PackInt64(v.UnixNano() / 1000000); err != nil {
		b = nil
		return
	}
	b = append(b, tmpV...)
	return
}

// double
func encodeFloat64(v float64) (b []byte, err error) {
	var tmpV []byte
	if tmpV, err = util.PackFloat64(v); err != nil {
		b = nil
		return
	}
	b = append(b, 'D')
	b = append(b, tmpV...)
	return
}

// int
func encodeInt32(v int32) (b []byte, err error) {
	var tmpV []byte
	if tmpV, err = util.PackInt32(v); err != nil {
		b = nil
		return
	}
	b = append(b, 'I')
	b = append(b, tmpV...)
	return
}

// long
func encodeInt64(v int64) (b []byte, err error) {
	var tmpV []byte
	if tmpV, err = util.PackInt64(v); err != nil {
		b = nil
		return
	}
	b = append(b, 'L')
	b = append(b, tmpV...)
	return

}

// null
func encodeNull(v interface{}) (b []byte, err error) {
	b = append(b, 'N')
	return
}

// string
func encodeString(v string) (b []byte, err error) {
	var (
		lenB []byte
		sBuf = *bytes.NewBufferString(v)
		rLen = utf8.RuneCountInString(v)

		sChunk = func(_len int) {
			for i := 0; i < _len; i++ {
				if r, s, err := sBuf.ReadRune(); s > 0 && err == nil {
					b = append(b, []byte(string(r))...)
				}
			}
		}
	)

	if v == "" {
		if lenB, err = util.PackUint16(uint16(rLen)); err != nil {
			b = nil
			return
		}
		b = append(b, 'S')
		b = append(b, lenB...)
		b = append(b, []byte{}...)
		return
	}

	for {
		rLen = utf8.RuneCount(sBuf.Bytes())
		if rLen == 0 {
			break
		}
		if rLen > CHUNK_SIZE {
			if lenB, err = util.PackUint16(uint16(CHUNK_SIZE)); err != nil {
				b = nil
				return
			}
			b = append(b, 's')
			b = append(b, lenB...)
			sChunk(CHUNK_SIZE)
		} else {
			if lenB, err = util.PackUint16(uint16(rLen)); err != nil {
				b = nil
				return
			}
			b = append(b, 'S')
			b = append(b, lenB...)
			sChunk(rLen)
		}
	}
	return
}

// list
func encodeList(v []Any) (b []byte, err error) {
	list_len := len(v)
	var (
		lenB []byte
		tmpV []byte
	)

	b = append(b, 'V')

	if lenB, err = util.PackInt32(int32(list_len)); err != nil {
		b = nil
		return
	}
	b = append(b, 'l')
	b = append(b, lenB...)

	for _, a := range v {
		if tmpV, err = Encode(a); err != nil {
			b = nil
			return
		}
		b = append(b, tmpV...)
	}
	b = append(b, 'z')
	return
}

// map
func encodeMap(v map[Any]Any) (b []byte, err error) {
	var (
		tmpK []byte
		tmpV []byte
	)
	b = append(b, 'M')
	for k, v := range v {
		if tmpK, err = Encode(k); err != nil {
			b = nil
			return
		}
		if tmpV, err = Encode(v); err != nil {
			b = nil
			return
		}
		b = append(b, tmpK...)
		b = append(b, tmpV...)
	}
	b = append(b, 'z')
	return
}
