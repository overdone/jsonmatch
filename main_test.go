package jsonmatch

import (
	"encoding/json"
	"os"
	"testing"
)

var comparatorIgnoreOrder сomparator
var comparatorWithOrder сomparator
var comparatorSkipDepth сomparator

type JsonTestSet []struct {
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

func TestJsonCompareBaseCases(t *testing.T) {
	tests := JsonTestSet{
		{true, `""`, `""`},
		{true, `{}`, `{}`},
		{true, `[]`, `[]`},
		{true, `1`, `1`},
		{true, `"1"`, `"1"`},
		{true, `null`, `null`},
		{true, `true`, `true`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorIgnoreOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonCompareNestedObjects(t *testing.T) {
	tests := JsonTestSet{
		{true, `{"a":[3,2,1],"b":{"b":["d","a","c"]}}`, `{"a":[3,1,2],"b":{"b":["d","c","a"]}}`},
		{true, `{"a":[3,2,1],"b":2}`, `{"b":2,"a":[3,1,2]}`},
		{false, `{"o":{"a":{"g":1,"c":1}}}`, `{"o":{"a":{"c":1,"b":1}}}`},
		{false, `{"o":{"a":{"g":1,"c":1}}}`, `{"o":{"a":{"g":1,"c":1,"f":1}}}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorIgnoreOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonCompareDifferentFiledTypes(t *testing.T) {
	tests := JsonTestSet{
		{false, `{"a":[3,2,1],"b":"2"}`, `{"b":2,"a":[3,1,2]}`},
		{false, `{"a":[3,2,1],"b":[]}`, `{"b":{},"a":[3,1,2]}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorIgnoreOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonComparePrimitiveArrayIgnoreOrder(t *testing.T) {
	tests := JsonTestSet{
		{true, `[1, 2, 3]`, `[3, 2, 1]`},
		{true, `["a", "b", "c"]`, `["c", "b", "a"]`},
		{true, `[1.1, 1.2, 1.3]`, `[1.3, 1.2, 1.1]`},
		{true, `[true, true, false]`, `[false, true, true]`},
		{false, `{"a":[3,2,1,1]}`, `{"a":[3,2,1]}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorIgnoreOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonComparePrimitiveArrayWithOrder(t *testing.T) {
	tests := JsonTestSet{
		{false, `[1, 2, 3]`, `[3, 2, 1]`},
		{true, `["a", "b", "c"]`, `["a", "b", "c"]`},
		{true, `[1.1, 1.2, 1.3]`, `[1.1, 1.2, 1.3]`},
		{false, `[true, true, false]`, `[false, true, true]`},
		{false, `{"a":[3,2,1,1]}`, `{"a":[3,2,1]}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorWithOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonMultiTypeArrayIgnoreOrder(t *testing.T) {
	tests := JsonTestSet{
		{true, `{"a":[1,2,3,"a"]}`, `{"a":["a",1,2,3]}`},
		{true, `{"a":[1,{"o":1},true]}`, `{"a":[{"o":1},true,1]}`},
		{false, `{"a":[1,2,3]}`, `{"a":[{},{},{}]}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorIgnoreOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonMultiTypeArrayWithOrder(t *testing.T) {
	tests := JsonTestSet{
		{true, `{"a":[1,2,3,"a"]}`, `{"a":[1,2,3,"a"]}`},
		{true, `{"a":[1,{"o":1},true]}`, `{"a":[1,{"o":1},true]}`},
		{false, `{"a":[1,{"o":1},true]}`, `{"a":[{"o":1},true,1]}`},
		{false, `{"a":[1,2,3,"a"]}`, `{"a":["a",1,2,3]}`},
		{false, `{"a":[1,2,3]}`, `{"a":[{},{},{}]}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorWithOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonCompareObjectArrayIgnoreOrder(t *testing.T) {
	tests := JsonTestSet{
		{true, `{"o":[{"b":[5,6,7,8,9]},{"b":[0,1,2,3,4]}]}`, `{"o":[{"b":[0,1,2,3,4]},{"b":[5,6,7,8,9]}]}`},
		{false, `{"a":[{"o":1},{"o":1},{"o":2}]}`, `{"a":[{"o":1},{"o":2},{"o":2}]}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorIgnoreOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonCompareObjectArrayWithOrder(t *testing.T) {
	tests := JsonTestSet{
		{false, `{"o":[{"b":[5,6,7,8,9]},{"b":[0,1,2,3,4]}]}`, `{"o":[{"b":[0,1,2,3,4]},{"b":[5,6,7,8,9]}]}`},
		{true, `{"a":[{"o":1},{"o":1},{"o":2}]}`, `{"a":[{"o":1},{"o":1},{"o":2}]}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorWithOrder.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
	}
}

func TestJsonCompareSkipDepth(t *testing.T) {
	tests := JsonTestSet{
		{true, `{"l2":{"l3":{"l4":"hello"}}}`, `{"l2":{"l3":{"l4":"bye"}}}`},
		{false, `{"l2":{"l3":"hello"}}`, `{"l2":{"l3":"bye"}}`},
		{true, `{"l2":[{"l4":1},{"l4":6}]}`, `{"l2":[{"l4":1},{"l4":6}]}`},
		{false, `{"l2":[{"l4":1},{"l4":6},{}]}`, `{"l2":[{"l4":1},{"l4":6}]}`},
	}

	for _, tst := range tests {
		var j1, j2 interface{}

		stringToJson(tst.In1, &j1)
		stringToJson(tst.In2, &j2)

		res, _ := comparatorSkipDepth.Compare(j1, j2)
		if res != tst.Expected {
			t.Errorf("got %v, wanted %v", res, tst.Expected)
		}
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

func setup() {
	comparatorIgnoreOrder = NewComparator(CompareOptions{ignoreArrayOrder: true})
	comparatorWithOrder = NewComparator(CompareOptions{ignoreArrayOrder: false})
	comparatorSkipDepth = NewComparator(CompareOptions{skipDepthGreater: 3})
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}
