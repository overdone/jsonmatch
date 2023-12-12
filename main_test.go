package jsonmatch

import (
	"encoding/json"
	"os"
	"testing"
)

var comparatorIgnoreOrder JsonComparator
var comparatorWithOrder JsonComparator
var comparatorSkipDepth JsonComparator

type JsonTestSet []struct {
	Name     string
	Expected bool
	In1      string
	In2      string
}

type ArrayTestSet []struct {
	Expected bool
	Input    []interface{}
}

type Array2TestSet []struct {
	Expected bool
	Input1   []interface{}
	Input2   []interface{}
}

func stringToJson(input string, output *interface{}) {
	if err := json.Unmarshal([]byte(input), output); err != nil {
		panic("Error while unmarshaling string to json")
	}
}

func TestIsOnlyPrimitiveItems(t *testing.T) {
	tests := ArrayTestSet{
		{true, []interface{}{1, "2", 3.0, true}},
		{false, []interface{}{1, "2", []int{1, 2}}},
		{false, []interface{}{1, "2", struct{}{}}},
	}

	for _, tst := range tests {
		res := isOnlyPrimitiveItems(tst.Input)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestGetCardinalityMap(t *testing.T) {
	a := []interface{}{1, 1, "a", "b", "b", "b", true}
	expected := map[interface{}]int{
		1:    2,
		"a":  1,
		"b":  3,
		true: 1,
	}

	m := getCardinalityMap(a)

	if len(m) != len(expected) {
		t.Errorf("got len %v, wanted len %v", len(m), len(expected))
	}

	for k, v := range m {
		if e, ok := expected[k]; !ok || e != v {
			t.Errorf("got %v for %v, wanted %v", v, k, e)
		}
	}
}

func TestAppendStep(t *testing.T) {
	r := appendStep("", "a")
	if r != "a" {
		t.Errorf("got %v, wanted \"a\"", r)
	}

	r = appendStep("b", "a")
	if r != "b.a" {
		t.Errorf("got %v, wanted \"b.a\"", r)
	}
}

func TestIgnoreOrderComparator(t *testing.T) {
	var tests = JsonTestSet{
		{"Base empty string", true, `""`, `""`},
		{"Base empty object", true, `{}`, `{}`},
		{"Base empty array", true, `[]`, `[]`},
		{"Base number", true, `1`, `1`},
		{"Base string", true, `"1"`, `"1"`},
		{"Base null", true, `null`, `null`},
		{"Base bool", true, `true`, `true`},
		{"Nested object 1", true, `{"a":[3,2,1],"b":{"b":["d","a","c"]}}`, `{"a":[3,1,2],"b":{"b":["d","c","a"]}}`},
		{"Nested object 2", true, `{"a":[3,2,1],"b":2}`, `{"b":2,"a":[3,1,2]}`},
		{"Nested object 3", false, `{"o":{"a":{"g":1,"c":1}}}`, `{"o":{"a":{"c":1,"b":1}}}`},
		{"Nested object 4", false, `{"o":{"a":{"g":1,"c":1}}}`, `{"o":{"a":{"g":1,"c":1,"f":1}}}`},
		{"Diff field type 1", false, `{"a":[3,2,1],"b":"2"}`, `{"b":2,"a":[3,1,2]}`},
		{"Diff field type 2", false, `{"a":[3,2,1],"b":[]}`, `{"b":{},"a":[3,1,2]}`},
		{"Primitive array 1", true, `[1, 2, 3]`, `[3, 2, 1]`},
		{"Primitive array 2", true, `["a", "b", "c"]`, `["c", "b", "a"]`},
		{"Primitive array 3", true, `[1.1, 1.2, 1.3]`, `[1.3, 1.2, 1.1]`},
		{"Primitive array 4", true, `[true, true, false]`, `[false, true, true]`},
		{"Primitive array 5", false, `{"a":[3,2,1,1]}`, `{"a":[3,2,1]}`},
		{"Multitype array 1", true, `{"a":[1,2,3,"a"]}`, `{"a":["a",1,2,3]}`},
		{"Multitype array 2", true, `{"a":[1,{"o":1},true]}`, `{"a":[{"o":1},true,1]}`},
		{"Multitype array 3", false, `{"a":[1,2,3]}`, `{"a":[{},{},{}]}`},
		{"Object array 1", true, `{"o":[{"b":[5,6,7,8,9]},{"b":[0,1,2,3,4]}]}`, `{"o":[{"b":[0,1,2,3,4]},{"b":[5,6,7,8,9]}]}`},
		{"Object array 2", false, `{"a":[{"o":1},{"o":1},{"o":2}]}`, `{"a":[{"o":1},{"o":2},{"o":2}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var j1, j2 interface{}

			stringToJson(tt.In1, &j1)
			stringToJson(tt.In2, &j2)

			res, _ := comparatorIgnoreOrder.Compare(j1, j2)

			if res != tt.Expected {
				t.Errorf("got %v, want %v", res, tt.Expected)
			}
		})
	}
}

func TestWithOrderComparator(t *testing.T) {
	var tests = JsonTestSet{
		{"Primitive array 1", false, `[1, 2, 3]`, `[3, 2, 1]`},
		{"Primitive array 2", true, `["a", "b", "c"]`, `["a", "b", "c"]`},
		{"Primitive array 3", true, `[1.1, 1.2, 1.3]`, `[1.1, 1.2, 1.3]`},
		{"Primitive array 4", false, `[true, true, false]`, `[false, true, true]`},
		{"Primitive array 5", false, `{"a":[3,2,1,1]}`, `{"a":[3,2,1]}`},
		{"Multitype array 1", true, `{"a":[1,2,3,"a"]}`, `{"a":[1,2,3,"a"]}`},
		{"Multitype array 2", true, `{"a":[1,{"o":1},true]}`, `{"a":[1,{"o":1},true]}`},
		{"Multitype array 3", false, `{"a":[1,{"o":1},true]}`, `{"a":[{"o":1},true,1]}`},
		{"Multitype array 4", false, `{"a":[1,2,3,"a"]}`, `{"a":["a",1,2,3]}`},
		{"Multitype array 5", false, `{"a":[1,2,3]}`, `{"a":[{},{},{}]}`},
		{"Object array 1", false, `{"o":[{"b":[5,6,7,8,9]},{"b":[0,1,2,3,4]}]}`, `{"o":[{"b":[0,1,2,3,4]},{"b":[5,6,7,8,9]}]}`},
		{"Object array 2", true, `{"a":[{"o":1},{"o":1},{"o":2}]}`, `{"a":[{"o":1},{"o":1},{"o":2}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var j1, j2 interface{}

			stringToJson(tt.In1, &j1)
			stringToJson(tt.In2, &j2)

			res, _ := comparatorWithOrder.Compare(j1, j2)

			if res != tt.Expected {
				t.Errorf("got %v, want %v", res, tt.Expected)
			}
		})
	}
}

func TestSkipDepthComparator(t *testing.T) {
	var tests = JsonTestSet{
		{"Skip depth 1", true, `{"l2":{"l3":{"l4":"hello"}}}`, `{"l2":{"l3":{"l4":"bye"}}}`},
		{"Skip depth 2", false, `{"l2":{"l3":"hello"}}`, `{"l2":{"l3":"bye"}}`},
		{"Skip depth 3", true, `{"l2":[{"l4":1},{"l4":6}]}`, `{"l2":[{"l4":1},{"l4":6}]}`},
		{"Skip depth 4", false, `{"l2":[{"l4":1},{"l4":6},{}]}`, `{"l2":[{"l4":1},{"l4":6}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var j1, j2 interface{}

			stringToJson(tt.In1, &j1)
			stringToJson(tt.In2, &j2)

			res, _ := comparatorSkipDepth.Compare(j1, j2)

			if res != tt.Expected {
				t.Errorf("got %v, want %v", res, tt.Expected)
			}
		})
	}
}

func setup() {
	comparatorIgnoreOrder = NewComparator(CompareOptions{IgnoreArrayOrder: true})
	comparatorWithOrder = NewComparator(CompareOptions{IgnoreArrayOrder: false})
	comparatorSkipDepth = NewComparator(CompareOptions{SkipDepthGreater: 3})
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}
