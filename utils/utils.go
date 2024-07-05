package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PrettyFormat is a function that receives any data type, converts it to a
// JSON string, and formats it with indents for readability. It uses the
// standard library's json.MarshalIndent function to create a string
// representation of the input data with two-space indents.
func PrettyFormat(x any) string {
	b, _ := json.MarshalIndent(x, "", "  ")
	return string(b)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TrimQuotes is a function that receives a string and removes double quote characters
// from the start and end of the string, if they exist. It returns the string with
// removed quotes. If the input string doesn't start and end with a quote, or if it's
// less than 2 characters long, it returns the original string.
func TrimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NormalizeWindowsDrive normalizes a given path to ensure that drive letters on Windows
// are correctly interpreted. If the path starts with a drive letter, it appends a
// backslash (\) to paths like "C:" to make them "C:\", and inserts a backslash in paths
// like "C:folder" to make them "C:\folder"
func NormalizeWindowsDrive(path string) string {
	if runtime.GOOS == "windows" {
		if len(path) >= 2 && path[1] == ':' {
			if len(path) == 2 {
				path += `\`
			} else if path[2] != '\\' {
				path = path[:2] + `\` + path[2:]
			}
		}
	}

	return path
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DecodeString is a function that receives a Base64-encoded string and first decodes
// it from Base64 and then URL-decodes it. The function returns the decoded string, or
// an error if either of the decoding operations fails. It uses standard library
// functions for both decoding operations.
func DecodeString(p string) (string, error) {
	bytePath, err := base64.StdEncoding.DecodeString(p)
	if err != nil {
		return "", fmt.Errorf("failed to decode path")
	}

	decodedPath, err := url.QueryUnescape(string(bytePath))
	if err != nil {
		return "", fmt.Errorf("failed to unescape path")

	}

	return decodedPath, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// EncodeString is a function that receives a string, URL-encodes it, and then encodes
// the result in Base64. The function returns the Base64-encoded string. It uses
// standard library functions for both encoding operations.
func EncodeString(p string) string {
	encodedPath := url.QueryEscape(p)

	res := base64.StdEncoding.EncodeToString([]byte(encodedPath))

	return res
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DiffSliceOfStructsByKey takes in two slices of type T (left and right) and a key (string) as
// arguments. The key defines the which key to use when comparing.
//
// It returns two slices:
//   - Items in the left slice that are not in the right slice based on the provided key
//   - Items in the right slice that are not in the left slice based on the provided key
//
// If both left and right slices are empty, nil and nil is returned
// If left is empty and right is not, nil and the entire right slice are returned
// If right is empty and left is not, the entire left slice and nil are returned
//
// When either left or right is not a struct or the key is not a valid key for the struct,
// nil and nil are returned
//
// The function uses maps to optimize lookup operations and determine differences between the
// slices
func DiffSliceOfStructsByKey[T any](left, right []T, key string) ([]T, []T, error) {
	leftDiff := []T{}
	rightDiff := []T{}

	// Check if the slices contain structs or pointers to structs
	if len(left) > 0 && !IsStructWithKey(left[0], key) || len(right) > 0 && !IsStructWithKey(right[0], key) {
		return nil, nil, fmt.Errorf("invalid struct or key")
	}

	// Both are empty
	if len(left) == 0 && len(right) == 0 {
		return nil, nil, nil
	}

	// Left is empty, return right
	if len(left) == 0 {
		return nil, right, nil
	}

	// Right is empty, return left
	if len(right) == 0 {
		return left, nil, nil
	}

	leftMap := make(map[string]T)
	rightMap := make(map[string]T)

	for _, v := range left {
		meta := reflect.ValueOf(v).Elem()
		field := meta.FieldByName((key))
		if field != (reflect.Value{}) {
			leftMap[ValueToString(field)] = v
		}
	}

	for _, v := range right {
		meta := reflect.ValueOf(v).Elem()
		field := meta.FieldByName((key))
		if field != (reflect.Value{}) {
			rightMap[ValueToString(field)] = v
		}
	}

	for k, v := range leftMap {
		if _, ok := rightMap[k]; !ok {
			leftDiff = append(leftDiff, v)
		}
	}

	for k, v := range rightMap {
		if _, ok := leftMap[k]; !ok {
			rightDiff = append(rightDiff, v)
		}
	}

	return leftDiff, rightDiff, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CompareStructs compares deeply nested structs, returning true if they are equal. It can also
// ignore specified keys during the comparison
//
// Parameters:
// - a: The first struct to compare (struct or a pointer to a struct)
// - b: The second struct to compare (struct or a pointer to a struct)
// - ignoreKeys: A slice of strings representing the field names to ignore during the comparison
//
// Returns:
// - true if the structs are equal, false otherwise
func CompareStructs(a, b interface{}, ignoreKeys []string) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	// Ensure we are dealing with structs or pointers to structs
	if valA.Kind() == reflect.Ptr {
		valA = valA.Elem()
	}
	if valB.Kind() == reflect.Ptr {
		valB = valB.Elem()
	}

	if valA.Kind() != reflect.Struct || valB.Kind() != reflect.Struct {
		return false
	}

	ignoreMap := make(map[string]bool)
	for _, key := range ignoreKeys {
		ignoreMap[key] = true
	}

	for i := 0; i < valA.NumField(); i++ {
		fieldA := valA.Type().Field(i)
		if ignoreMap[fieldA.Name] {
			continue
		}

		// Skip unexported fields
		if !fieldA.IsExported() {
			continue
		}

		fieldB := valB.FieldByName(fieldA.Name)
		if !fieldB.IsValid() {
			return false
		}

		if fieldA.Type.Kind() == reflect.Struct {
			if !CompareStructs(valA.Field(i).Interface(), fieldB.Interface(), ignoreKeys) {
				return false
			}
		} else if !reflect.DeepEqual(valA.Field(i).Interface(), fieldB.Interface()) {
			return false
		}
	}

	return true
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsStructWithKey checks if the provided value is a struct or a pointer to a struct
// and if it contains the provided key.
func IsStructWithKey(value any, key string) bool {
	t := reflect.TypeOf(value)

	if t.Kind() == reflect.Struct {
		_, ok := t.FieldByName(key)
		return ok
	} else if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		_, ok := t.Elem().FieldByName(key)
		return ok
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ValueToString converts a reflect.Value to a string. It supports basic types like
// int, uint, float, string, and bool. For other types, it returns an empty string
func ValueToString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		// Return an empty string for unsupported types
		return ""
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Map is a generic function that takes a slice of type T and a function that
// maps T to type V. It returns a new slice of type V with the mapped values
func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}
