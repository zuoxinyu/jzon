package jzon

import (
	"testing"
)

func TestParse(t *testing.T) {
	const testCorrect = `{
		"key1" : "value1" ,
		"key2" : ["string",true,null,false] ,
		"key3" : 1234,
		"key4" : "\u5f20\u91d1\u708e\u8001\u5e08\u7684\u76f4\u64ad\u8bb2\u5ea7",
		"key5" : [],
		"key6" : {},
		"key7" : "汉字",
		"key8" : "の"
	}`

	_, err := Parse([]byte(testCorrect))
	if err != nil {
		t.Error(err)
	}
}

func TestParseHex4(t *testing.T) {
	pos.col = 0
	pos.row = 0
	s := `0020`
	h, rem, err := parseHex4([]byte(s))
	if err != nil {
		t.Error(err)
	}

	if len(rem) != 0 {
		t.Errorf("expect empty slice, but rem is %v", rem)
	}

	if h != 32 {
		t.Errorf("expect h = 32, but h is %v", h)
	}
}

func TestParseUnicode(t *testing.T) {
	pos.col = 0
	pos.row = 0
	s := []byte("\\u5f20abcd")

	str, rem, err := parseUnicode(s)
	if err != nil {
		t.Error(err)
	}

	if len(rem) != 4 {
		t.Errorf("expect len(rem) = 4, but rem is %s", string(rem))
	}

	kanji := []byte{}
	for _, c := range str {
		kanji = append(kanji, c)
	}

	if string(kanji) != "张" {
		zhang := []byte("张")
		t.Errorf("len(kanji) is %v", len(kanji))
		t.Errorf("kanji in hex is %x %x %x", kanji[0], kanji[1], kanji[2])
		t.Errorf("张    in hex is %x %x %x", zhang[0], zhang[1], zhang[2])
		t.Errorf("expect kanji = 张, but kanji is %s", string(kanji))
	}
}

func TestParseKey(t *testing.T) {
	s := []byte(`"\u5f20\u91d1\u708e\u8001\u5e08\u7684\u76f4\u64ad\u8bb2\u5ea7"`)
	key, rem, err := parseKey(s)
	if err != nil {
		t.Error(err)
	}

	if string(key) != "张金炎老师的直播讲座" {
		t.Errorf("expect key = 张金炎老师的直播讲座, but key is %v", string(key))
	}

	if len(rem) != 0 {
		t.Errorf("expect rem = empty []byte, but rem is %v", rem)
	}
}

func TestParseNum(t *testing.T) {
	numeric := []byte(`1234`)
	n, rem, err := parseNum(numeric)
	if err != nil {
		t.Error(err)
	}

	num, _ := n.Number()

	if num != 1234 {
		t.Errorf("expect num = 1234, but num is %d", num)
	}

	if len(rem) != 0 {
		t.Errorf("expect rem = empty []byte, but rem is %v", rem)
	}
}
