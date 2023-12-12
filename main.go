package jsonmatch

import (
	"fmt"
	"reflect"
	"strconv"
)

type oJSONObject map[string]interface{}

type oJSONArray []interface{}

type cardinalityMap map[interface{}]int

type keySet map[string]struct{}

type CompareOptions struct {
	IgnoreArrayOrder bool
	SkipDepthGreater int
}

type JsonComparator struct {
	opts CompareOptions
}

func (c JsonComparator) Compare(left interface{}, right interface{}) (bool, error) {
	return c.jsonCompare(left, right, "", 1)
}

func (c JsonComparator) jsonCompare(left interface{}, right interface{}, path string, depth int) (bool, error) {
	if c.opts.SkipDepthGreater != 0 && depth > c.opts.SkipDepthGreater {
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

func (c JsonComparator) compareObject(lo oJSONObject, ro oJSONObject, path string, depth int) (bool, error) {
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

func (c JsonComparator) compareArray(la oJSONArray, ra oJSONArray, path string, depth int) (bool, error) {
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
		if c.opts.IgnoreArrayOrder {
			return comparePrimitiveArrayIgnoreOrder(la, ra, path, depth)
		}
		return comparePrimitiveArrayWithOrder(la, ra, path, depth)
	}

	if c.opts.IgnoreArrayOrder {
		return c.recursiveCompareArrayIgnoreOrder(la, ra, path, depth)
	}
	return c.recursiveCompareArrayWithOrder(la, ra, path, depth)
}

func (c JsonComparator) recursiveCompareArrayIgnoreOrder(la oJSONArray, ra oJSONArray, path string, depth int) (bool, error) {
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

func (c JsonComparator) recursiveCompareArrayWithOrder(la oJSONArray, ra oJSONArray, path string, depth int) (bool, error) {
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

func comparePrimitiveArrayIgnoreOrder(la oJSONArray, ra oJSONArray, path string, depth int) (bool, error) {
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

func comparePrimitiveArrayWithOrder(la oJSONArray, ra oJSONArray, path string, depth int) (bool, error) {
	for i := range la {
		if la[i] != ra[i] {
			return false, getMatchError(path)
		}
	}

	return true, nil
}

func isOnlyPrimitiveItems(a oJSONArray) bool {
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

func getCardinalityMap(a oJSONArray) cardinalityMap {
	m := make(cardinalityMap)

	for _, v := range a {
		m[v] += 1
	}

	return m
}

func getKeySet(o oJSONObject) keySet {
	m := make(keySet)

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

func NewComparator(opts CompareOptions) JsonComparator {
	return JsonComparator{opts}
}
