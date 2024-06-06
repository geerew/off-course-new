package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/geerew/off-course/utils/appFs"
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

// TimeTrack is a function that receives a time value and a name string,
// calculates the time elapsed since the provided time, and then prints
// out the name and elapsed time. It is used for simple performance profiling
// of code sections.
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("name: %s, elapsed: %s\n", name, elapsed)
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

// DiffStructsByKey takes in two slices of type T (left and right) and a key (string) as arguments.
// The key defines the which key to use when comparing.
//
// It returns two slices:
//   - Items in the left slice that are not in the right slice based on the provided key.
//   - Items in the right slice that are not in the left slice based on the provided key.
//
// If both left and right slices are empty, nil and nil is returned.
// If left is empty and right is not, nil and the entire right slice are returned.
// If right is empty and left is not, the entire left slice and nil are returned.
//
// When either left or right is not a struct or the key is not a valid key for the struct,
// nil and nil are returned.
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

// CompareStructs compares two structs and returns true if they are equal, ignoring the specified keys. It
// also supports comparing nested structs
//
// Parameters:
// - a: The first struct to compare. Can be a struct or a pointer to a struct
// - b: The second struct to compare. Can be a struct or a pointer to a struct
// - ignoreKeys: A slice of strings representing the field names to ignore during the comparison
//
// Returns:
// - true if the structs are equal (ignoring the specified keys), false otherwise
//
// The function works as follows:
//  1. It checks if the types of a and b are the same. If not, it returns false
//  2. It dereferences pointers to structs, if necessary, so that it can work with the actual structs
//  3. It ensures that a and b are structs. If they are not, it returns false
//  4. It creates a map from the ignoreKeys slice for quick lookup of keys to ignore
//  5. It iterates over the fields of struct a, skipping any fields that are in the ignoreKeys map or are
//     unexported
//  6. For each field, it checks if the corresponding field exists in struct b. If not, it returns false
//  7. If a field is itself a struct, it recursively calls CompareStructs to compare the nested structs
//  8. For non-struct fields, it uses reflect.DeepEqual to check if the field values are equal. If any field
//     values are not equal, it returns false
//  9. If all fields (except the ignored ones) are equal, it returns true
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
// maps T to type V. It returns a new slice of type V with the mapped values.
func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// partialHash is a function that receives a file path and a chunk size as arguments and
// returns a partial hash of the file, by reading the first, middle, and last
// chunks of the file, as well as two random chunks, and hashes them together.
//
// It uses the SHA-256 hashing algorithm from the standard library to calculate the hash
func PartialHash(appFs *appFs.AppFs, filePath string, chunkSize int64) (string, error) {
	file, err := appFs.Fs.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	// Append file size to the hash
	fileSize := fileInfo.Size()
	binary.Write(hash, binary.LittleEndian, fileSize)

	// Function to read and hash a chunk at a given position
	readAndHashChunk := func(position int64) error {
		_, err := file.Seek(position, 0)
		if err != nil {
			return err
		}
		chunk := make([]byte, chunkSize)
		n, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		hash.Write(chunk[:n])
		return nil
	}

	// Read and hash the first chunk
	if err = readAndHashChunk(0); err != nil {
		return "", err
	}

	// Read and hash the middle chunk
	middlePosition := fileSize / 2
	if middlePosition < fileSize {
		if err = readAndHashChunk(middlePosition); err != nil {
			return "", err
		}
	}

	// Read and hash the last chunk
	lastPosition := fileSize - chunkSize
	if lastPosition < 0 {
		lastPosition = 0
	}
	if lastPosition < fileSize {
		if err = readAndHashChunk(lastPosition); err != nil {
			return "", err
		}
	}

	// Random chunks
	additionalPositions := []int64{fileSize / 4, 3 * fileSize / 4}
	for _, position := range additionalPositions {
		if position < fileSize {
			if err = readAndHashChunk(position); err != nil {
				return "", err
			}
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
