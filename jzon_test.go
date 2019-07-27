package jzon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const deepJSON = `
{
	"key-object": {
		"key-o-o": {
			"number": 1234,
			"string": "a string"
		}
	},
	"key-array": [
		{
			"number": 4567,
			"string": "another string 1",
			"null": null,
			"bool": false,
			"empty-object": {

			}
		},
		{
			"number": 4567,
			"string": "another string 2",
			"null": null,
			"bool": false,
			"empty-object": {

			}
		},
		{
			"number": 4567,
			"string": "another string 3",
			"null": null,
			"bool": false,
			"empty-object": {

			}
		}
	],
	"key-汉字": "值也是汉字",
	"key-escaped-.[];-key": "escape success"
}`

// parser.go

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

	parseFile := func(path string) error {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = Parse(content)
		if err != nil {
			return err
		}
		return nil
	}

	filepath.Walk("data/roundtrip", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		fmt.Print(path)
		err = parseFile(path)
		if err != nil {
			t.Errorf("fail on file: %v", path)
		}
		fmt.Print("\tpassed\n")
		return nil
	})

	filepath.Walk("data/jsonchecker", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		base := filepath.Base(path)

		if strings.HasPrefix(base, "pass") {
			fmt.Print(base)
			err = parseFile(path)
			if err != nil {
				fmt.Print("\tfailed\n")
			} else {
				fmt.Print("\tpassed\n")
			}
		}

		if strings.HasPrefix(base, "fail") {
			fmt.Print(base)
			err = parseFile(path)
			if err == nil {
				fmt.Print("\tfailed\n")
			} else {
				fmt.Print("\tpassed\n")
			}
		}
		return nil
	})
}

func TestParseObj(t *testing.T) {
	const text = `{
		"key1" : "value1" ,
		"key2" : ["string",true,null,false] ,
		"key3" : 1234,
		"key4" : "\u5f20\u91d1\u708e\u8001\u5e08\u7684\u76f4\u64ad\u8bb2\u5ea7",
		"key5" : [],
		"key6" : {},
		"key7" : "汉字",
		"key8" : "の"
	}`

	o, rem, err := parseObj([]byte(text))
	if err != nil {
		t.Error(err)
	}

	obj, _ := o.Object()

	if len(obj) != 8 {
		t.Errorf("expect len(obj) = 8, but len(obj) is %v", len(obj))
	}

	if len(rem) != 0 {
		t.Errorf("expect rem = empty []byte, but rem is %v", rem)
	}
}

func TestParseArr(t *testing.T) {
	const text = `[
		"key1", "key2", "key3", "key4",
		"key5", "key6", "key7", "key8"
	]`

	a, rem, err := parseArr([]byte(text))
	if err != nil {
		t.Error(err)
	}

	arr, _ := a.Array()

	if len(arr) != 8 {
		t.Errorf("expect len(arr) = 8, but len(arr) is %v", len(arr))
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

	num, _ := n.Integer()

	if num != 1234 {
		t.Errorf("expect num = 1234, but num is %d", num)
	}

	if len(rem) != 0 {
		t.Errorf("expect rem = empty []byte, but rem is %v", rem)
	}
}

func TestParseTru(t *testing.T) {
	boolean := []byte(`true`)
	b, rem, err := parseTru(boolean)
	if err != nil {
		t.Error(err)
	}

	bol, _ := b.Bool()

	if bol != true {
		t.Errorf("expect bol = true, but bol is %v", bol)
	}

	if len(rem) != 0 {
		t.Errorf("expect rem = empty []byte, but rem is %v", rem)
	}
}

func TestParseFls(t *testing.T) {
	boolean := []byte(`false`)
	b, rem, err := parseFls(boolean)
	if err != nil {
		t.Error(err)
	}

	bol, _ := b.Bool()

	if bol != false {
		t.Errorf("expect bol = false, but bol is %v", bol)
	}

	if len(rem) != 0 {
		t.Errorf("expect rem = empty []byte, but rem is %v", rem)
	}
}

func TestParseNul(t *testing.T) {
	null := []byte(`null`)
	nul, rem, err := parseNul(null)
	if err != nil {
		t.Error(err)
	}

	if !nul.IsNull() {
		t.Errorf("expect nul of type JzTypeNul, but nul of type: %s", typeStrings[nul.Type])
	}

	if len(rem) != 0 {
		t.Errorf("expect rem = empty []byte, but rem is %v", rem)
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

	kanji := make([]byte, 0)
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

func TestParseNumeric(t *testing.T) {
	var integer = []byte("1234")
	var float = []byte("12.34")
	var frac = []byte("1.2E+04")
	var zero = []byte("0")
	var neg = []byte("-12")
	var more = []byte("123.4f")
	var big = []byte("10000000000000000000000000")

	var n int64
	var f float64
	var err error
	var isInt bool
	var rem []byte

	n, f, isInt, rem, err = parseNumeric(integer)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("n = %5d| f = %5.2f\t| isInt = %v\t| rem = \"%8s\"| err = %v\n", n, f, isInt, string(rem), err)

	if n == 0 {
		t.Errorf("expect n = 1234, but n = %d", n)
	}

	n, f, isInt, rem, err = parseNumeric(float)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("n = %5d| f = %5.2f\t| isInt = %v\t| rem = \"%8s\"| err = %v\n", n, f, isInt, string(rem), err)
	if f-12.34 >= 0.00001 {
		t.Errorf("expect f = 12.34, but f = %f", f)
	}

	n, f, isInt, rem, err = parseNumeric(frac)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("n = %5d| f = %5.2f\t| isInt = %v\t| rem = \"%8s\"| err = %v\n", n, f, isInt, string(rem), err)
	if f-12000.0 >= 0.01 {
		t.Errorf("expect f = 12000 but f = %f", f)
	}

	n, f, isInt, rem, err = parseNumeric(zero)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("n = %5d| f = %5.2f\t| isInt = %v\t| rem = \"%8s\"| err = %v\n", n, f, isInt, string(rem), err)
	if n != 0 {
		t.Errorf("expect n = 0 but n = %d", n)
	}

	n, f, isInt, rem, err = parseNumeric(neg)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("n = %5d| f = %5.2f\t| isInt = %v\t| rem = \"%8s\"| err = %v\n", n, f, isInt, string(rem), err)
	if n != -12 {
		t.Errorf("expect n = -12 but n = %d", n)
	}

	n, f, isInt, rem, err = parseNumeric(more)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("n = %5d| f = %5.2f\t| isInt = %v\t| rem = \"%8s\"| err = %v\n", n, f, isInt, string(rem), err)
	if f-12000.0 >= 0.01 {
		t.Errorf("expect f = 12000 but f = %f", f)
	}

	if isInt {
		t.Errorf("expect isInt = false, but isInt is %v", isInt)
	}

	if rem[0] != 'f' {
		t.Errorf("expect rem[0] = 'f', but rem[0] is %v", rem[0])
	}

	n, f, isInt, rem, err = parseNumeric(big)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("n = %5d| f = %5.2f\t| isInt = %v\t| rem = \"%8s\"| err = %v\n", n, f, isInt, string(rem), err)
}

// query.go

func TestQuery(t *testing.T) {
	var res *Jzon
	var num int64
	var str string

	jz, err := Parse([]byte(deepJSON))
	if err != nil {
		t.Error(err)
	}

	res, err = jz.Query("$.key-object.key-o-o.number")
	if err != nil {
		t.Error(err)
	}

	if num, err = res.Integer(); err != nil {
		t.Error(err)
	}

	if num != 1234 {
		t.Errorf("expect num = 1234, but num is %v", num)
	}

	res, err = jz.Query(`$.key-array[1].string`)
	if err != nil {
		t.Error(err)
	}

	if str, err = res.String(); err != nil {
		t.Error(err)
	}

	if str != "another string 2" {
		t.Errorf("expect str = another string 2, but str is %v", str)
	}

	res, err = jz.Query("$.key-汉字")
	if err != nil {
		t.Error(err)
	}

	if str, err = res.String(); err != nil {
		t.Error(err)
	}

	if str != "值也是汉字" {
		t.Errorf("expect str = 值也是汉字, but str is %v", str)
	}

	res, err = jz.Query(`$.key-escaped-\.\[\]\;-key`)
	if err != nil {
		t.Error(err)
	}

	if str, err = res.String(); err != nil {
		t.Error(err)
	}

	if str != "escape success" {
		t.Errorf("expect str = escape success, but str is %v", str)
	}
}

func TestSearch(t *testing.T) {
	jz, err := Parse([]byte(deepJSON))
	if err != nil {
		t.Error(err)
	}

	ok := jz.Search(`$.key-escaped-\.\[\]\;-key`)

	if !ok {
		t.Errorf("expect ok = true")
	}
}

// utilities.go
func TestCompact(t *testing.T) {
	jz, err := Parse([]byte(deepJSON))
	if err != nil {
		t.Error(err)
	}

	fmt.Print(jz.Compact())
}

func TestFormat(t *testing.T) {
	jz, err := Parse([]byte(deepJSON))
	if err != nil {
		t.Error(err)
	}

	fmt.Println()
	fmt.Println()
	fmt.Print(jz.Format(0, 2))
	fmt.Println()
	fmt.Println()
}

func TestColoring(t *testing.T) {
	jz, err := Parse([]byte(deepJSON))
	if err != nil {
		t.Error(err)
	}

	jz.Coloring(os.Stdout)
	fmt.Println()
}

// reflect.go
type User struct {
	Name     string         `jzon:"name"`
	Pwd      string         `jzon:"pwd"`
	Model    int            `jzon:"model"`
	EmptyTag int            `jzon:","`
	Nest     Nested         `jzon:"nested"`
	Array    []int          `jzon:"array"`
	Map      map[string]int `jzon:"map"`
	NilSlice []byte         `jzon:","`
	Ptr      *int           `jzon:"ptr"`
	Bool     bool           `jzon:"bool"`
	NoTag    int
}

type ImpInterface int

func (i ImpInterface) SerializeJzon() string {
	return "INT" + strconv.Itoa(int(i))
}

type Nested struct {
	Inner string `jzon:"inner"`
}

func TestSerialize(t *testing.T) {
	user := User{
		Name:  "ZuoXinyu",
		Pwd:   "password",
		Model: 1,
		Nest:  Nested{Inner: "innerString"},
		Array: []int{1, 2, 3, 4, 5},
		Map: map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
			"four":  4,
		},
	}
	jz, err := Serialize(user)
	if err != nil {
		t.Error(err)
	}

	jz.Print()
	fmt.Printf("\n")

	m := map[ImpInterface]int{
		4: 1,
		3: 2,
		2: 1,
	}

	jz1, err := Serialize(m)
	if err != nil {
		t.Error(err)
	}

	jz1.Print()
	fmt.Printf("\n")
}

func TestDeserialize(t *testing.T) {
	js := `{"ptr":null,"pwd":"password","EmptyTag":0,"array":[1,2,3,4,5],"map":{"one":1,"two":2,"three":3,"four":4},"name":"ZuoXinyu","model":1,"nested":{"inner":"innerString"},"NilSlice":null, "bool": true}`
	user := User{}

	err := Deserialize([]byte(js), &user)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%#v\n", user)
}

// Benchmarks

func BenchmarkJzonParseTwitter(b *testing.B) {
	content, err := ioutil.ReadFile("data/twitter.json")
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	b.SetBytes(0)
	for i := 0; i < b.N; i++ {
		Parse(content)
	}
	b.ReportAllocs()
}

func BenchmarkJzonParseCanada(b *testing.B) {
	content, err := ioutil.ReadFile("data/canada.json")
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	b.SetBytes(0)
	for i := 0; i < b.N; i++ {
		Parse(content)
	}
	b.ReportAllocs()
}

func BenchmarkJzonParseCatalog(b *testing.B) {
	content, err := ioutil.ReadFile("data/citm_catalog.json")
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	b.SetBytes(0)
	for i := 0; i < b.N; i++ {
		Parse(content)
	}
	b.ReportAllocs()
}

func BenchmarkJsonParseTwitter(b *testing.B) {
	content, err := ioutil.ReadFile("data/twitter.json")
	if err != nil {
		b.Error(err)
	}
	var m map[string]interface{}
	b.ResetTimer()
	b.SetBytes(0)
	for i := 0; i < b.N; i++ {
		json.Unmarshal(content, &m)
	}
	b.ReportAllocs()
}

func BenchmarkJsonParseCanada(b *testing.B) {
	content, err := ioutil.ReadFile("data/canada.json")
	if err != nil {
		b.Error(err)
	}
	var m map[string]interface{}
	b.ResetTimer()
	b.SetBytes(0)
	for i := 0; i < b.N; i++ {
		json.Unmarshal(content, &m)
	}
	b.ReportAllocs()
}

func BenchmarkJsonParseCatalog(b *testing.B) {
	content, err := ioutil.ReadFile("data/citm_catalog.json")
	if err != nil {
		b.Error(err)
	}
	var m map[string]interface{}
	b.ResetTimer()
	b.SetBytes(0)
	for i := 0; i < b.N; i++ {
		json.Unmarshal(content, &m)
	}
	b.ReportAllocs()
}
