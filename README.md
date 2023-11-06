# JSONMatch

[![Go Report Card](https://goreportcard.com/report/github.com/overdone/jsonmatch)](https://goreportcard.com/report/github.com/overdone/jsonmatch)

A GO library for deep comparing of two JSONs.

## Key features
- Ignoring array order within comparing
- Ignoring comparing after some depth

## Examples
### Basic
```go
func main() {
	var j1, j2 interface{}

	s1 := `{"a":1, "b":2}`
	s2 := `{"b":2, "a":1}`

	json.Unmarshal([]byte(s1), &j1)
	json.Unmarshal([]byte(s2), &j2)

	comparator := NewComparator(CompareOptions{})
	res, _ := comparator.Compare(j1, j2)
	fmt.Println(res)
}
```
```
true
```

### Ignore array order

```go
func main() {
	var j1, j2 interface{}

	s1 := `{"a":[1,2,3]}`
	s2 := `{"a":[3,2,1]}`

	json.Unmarshal([]byte(s1), &j1)
	json.Unmarshal([]byte(s2), &j2)

	comparator := NewComparator(CompareOptions{ignoreArrayOrder: true})
	res, _ := comparator.Compare(j1, j2)
	fmt.Println(res)
}
```
```
true
```

### Ignore after depth
```go
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
```
```
true
true
```
