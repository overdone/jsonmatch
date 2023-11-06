package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type JSONObject map[string]interface{}

type JSONArray []interface{}

type CardinalityMap map[interface{}]int

type KeySet map[string]struct{}

type CompareOptions struct {
	ignoreArrayOrder bool
	skipDepthGreater int
}

type Comparator struct {
	opts CompareOptions
}

func (c Comparator) Compare(left interface{}, right interface{}) (bool, error) {
	return c.jsonCompare(left, right, "", 1)
}

func (c Comparator) jsonCompare(left interface{}, right interface{}, path string, depth int) (bool, error) {
	if c.opts.skipDepthGreater != 0 && depth > c.opts.skipDepthGreater {
		return true, nil
	}

	leftT := reflect.TypeOf(left)
	rightT := reflect.TypeOf(right)

	if leftT != rightT {
		return false, getMatchError(path)
	}

	if leftT == nil {
		return true, nil
	}

	switch leftT.Kind() {
	case reflect.Bool, reflect.Float64, reflect.String:
		return left == right, getMatchError(path)
	case reflect.Slice:
		return c.compareArray(left.([]interface{}), right.([]interface{}), path, depth)
	case reflect.Map:
		return c.compareObject(left.(map[string]interface{}), right.(map[string]interface{}), path, depth)
	default:
		return true, nil
	}
}

func (c Comparator) compareObject(lo JSONObject, ro JSONObject, path string, depth int) (bool, error) {
	if len(lo) != len(ro) {
		return false, getMatchError(path)
	}

	keys1 := getKeySet(lo)
	keys2 := getKeySet(ro)

	for k := range keys1 {
		if _, ok := keys2[k]; !ok {
			return false, getMatchError(appendStep(path, k))
		}
	}

	for k, v1 := range lo {
		res, err := c.jsonCompare(v1, ro[k], appendStep(path, k), depth+1)
		if !res {
			return false, err
		}
	}

	return true, nil
}

func (c Comparator) compareArray(la JSONArray, ra JSONArray, path string, depth int) (bool, error) {
	if len(la) != len(ra) {
		return false, getMatchError(path)
	}

	if len(la) == 0 {
		return true, nil
	}

	isPrimitive1 := isOnlyPrimitiveItems(la)
	isPrimitive2 := isOnlyPrimitiveItems(ra)

	if isPrimitive1 != isPrimitive2 {
		return false, getMatchError(path)
	} else if isPrimitive1 {
		if c.opts.ignoreArrayOrder {
			return comparePrimitiveArrayIgnoreOrder(la, ra, path, depth)
		}
		return comparePrimitiveArrayWithOrder(la, ra, path, depth)
	}

	if c.opts.ignoreArrayOrder {
		return c.recursiveCompareArrayIgnoreOrder(la, ra, path, depth)
	}
	return c.recursiveCompareArrayWithOrder(la, ra, path, depth)
}

func (c Comparator) recursiveCompareArrayIgnoreOrder(la JSONArray, ra JSONArray, path string, depth int) (bool, error) {
	matched := make(map[int]struct{})

	for j, v1 := range la {
		found := false
		t1 := reflect.TypeOf(v1)

		for i, v2 := range ra {
			if _, ok := matched[i]; ok {
				continue
			}

			t2 := reflect.TypeOf(v2)
			if t1 != t2 {
				continue
			}
			equal, _ := c.jsonCompare(v1, v2, appendStep(path, "["+strconv.Itoa(j)+"]"), depth+1)
			if equal {
				matched[i] = struct{}{}
				found = true
				break
			}
		}

		if !found {
			return false, getMatchError(appendStep(path, "["+strconv.Itoa(j)+"]"))
		}
	}

	return true, nil
}

func (c Comparator) recursiveCompareArrayWithOrder(la JSONArray, ra JSONArray, path string, depth int) (bool, error) {
	for i, v1 := range la {
		v2 := ra[i]
		t1 := reflect.TypeOf(v1)
		t2 := reflect.TypeOf(v2)

		p := appendStep(path, "["+strconv.Itoa(i)+"]")

		if t1 != t2 {
			return false, getMatchError(p)
		}

		equal, err := c.jsonCompare(v1, v2, p, depth+1)
		if !equal {
			return false, err
		}
	}

	return true, nil
}

func comparePrimitiveArrayIgnoreOrder(la JSONArray, ra JSONArray, path string, depth int) (bool, error) {
	a1map := getCardinalityMap(la)
	a2map := getCardinalityMap(ra)

	if len(a1map) != len(a2map) {
		return false, getMatchError(path)
	}

	for k, v1 := range a1map {
		if v2, ok := a2map[k]; !ok || v1 != v2 {
			return false, getMatchError(path)
		}
	}

	return true, nil
}

func comparePrimitiveArrayWithOrder(la JSONArray, ra JSONArray, path string, depth int) (bool, error) {
	for i := range la {
		if la[i] != ra[i] {
			return false, getMatchError(path)
		}
	}

	return true, nil
}

func isOnlyPrimitiveItems(a JSONArray) bool {
	for _, v := range a {
		k := reflect.TypeOf(v).Kind()
		switch k {
		case reflect.Bool, reflect.Float64, reflect.Int, reflect.String:
			continue
		default:
			return false
		}
	}

	return true
}

func getCardinalityMap(a JSONArray) CardinalityMap {
	m := make(CardinalityMap)

	for _, v := range a {
		m[v] += 1
	}

	return m
}

func getKeySet(o JSONObject) KeySet {
	m := make(KeySet)

	for k := range o {
		m[k] = struct{}{}
	}

	return m
}

func appendStep(path string, step string) string {
	if path == "" {
		return step
	}
	return path + "." + step
}

func getMatchError(path string) error {
	return fmt.Errorf("LJSON.%v not match RJSON.%v", path, path)
}

func NewComparator(opts CompareOptions) Comparator {
	return Comparator{opts}
}

func main() {
	var j1, j2, j3, j4 interface{}

	s1 := `{"l2":{"l3":{"l4":"hello"}}}`
	s2 := `{"l2":{"l3":{"l4":"bye"}}}`

	// Keep in mind array adds one more level by itself
	s3 := `{"l2":[{"l4":1},{"l4":1}]}`
	s4 := `{"l2":[{"l4":1},{"l4":0}]}`

	json.Unmarshal([]byte(s1), &j1)
	json.Unmarshal([]byte(s2), &j2)
	json.Unmarshal([]byte(s3), &j3)
	json.Unmarshal([]byte(s4), &j4)

	comparator := NewComparator(CompareOptions{skipDepthGreater: 3})
	res1, _ := comparator.Compare(j1, j2)
	res2, _ := comparator.Compare(j3, j4)
	fmt.Println(res1)
	fmt.Println(res2)
}
