package tools

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	fuzz "github.com/google/gofuzz"
)

const (
	ColorGreen  = "\033[32m"
	ColorRed    = "\033[31m"
	ColorBlue   = "\033[34m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

func DeleteAtIndex[t any](slice []t, index int) []t {
	return append(slice[:index], slice[index+1:]...)
}

func removeElement(slice *[]any, valuesToRemove ...any) {
	slice1 := *slice
loop:
	for i := 0; i < len(slice1); i++ {
		url := slice1[i]
		for _, rem := range valuesToRemove {
			if url == rem {
				slice1 = append(slice1[:i], slice1[i+1:]...)
				i-- // Important: decrease index
				continue loop
			}
		}
	}
	*slice = slice1
}

func IsSameSign(number1, number2 float64) bool {
	if number1 == 0 || number2 == 0 {
		return true
	}
	return (number1 > 0) == (number2 > 0)
}

func ChangeSign(number1 float64, number2 float64) float64 {
	switch {
	case number1 > 0:
		return math.Abs(number2)
	case number1 < 0:
		return -math.Abs(number2)
	}
	return number2
}

func Find[t comparable](element t, elements []t) (int, bool) {
	for k1, v1 := range elements {
		if v1 == element {
			return k1, true
		}
	}
	return 0, false
}

func Hash64Bit(b []byte) [8]byte {
	s := sha256.Sum256(b)
	s1 := [8]byte(s[:9])
	return s1
}

func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

var PanicIfError bool = true

// func GetFunctionName(i interface{}) string {
// 	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
// }

func errorPrefix(packageName string, errorIndex uint) string {
	return fmt.Sprintf("package (%v) error index (%v) ", packageName, errorIndex)
}

func Errorf(packageName string, errorIndex uint, format string, a ...any) error {
	if packageName == "" {
		panic("dont pass zero value to packageName")
	}
	if errorIndex == 0 {
		panic("dont pass zero value to errorIndex")
	}
	return fmt.Errorf(errorPrefix(packageName, errorIndex)+format, a...)
}

func ErrorHandler(packageName string, errorIndex uint, err error) bool {
	if err == nil {
		return false
	}
	errString := err.Error()
	prefix := errorPrefix(packageName, errorIndex)
	if len(errString) < len(prefix) {
		return false
	}
	for i, v := range prefix {
		if v != rune(errString[i]) {
			return false
		}
	}
	return true
}

func Encode(decoded any) ([]byte, error) {
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	err := enc.Encode(decoded)
	return network.Bytes(), err
}

func Decode[t any](encoded []byte) (t, error) {
	var decoded t
	network := bytes.NewReader(encoded)
	dec := gob.NewDecoder(network)
	err := dec.Decode(&decoded)
	return decoded, err
}

func Println(a ...any) {
	fmt.Println(ColorBlue, Stack(8), ColorReset)
	for i, v := range a {
		fmt.Printf("%v%v:%v%v\n", ColorYellow, i, ColorReset, v)
	}
}

func Stack(line uint) string {
	if line < 8 {
		line = 8
	}
	if line%2 != 0 {
		line++
	}
	return strings.Split(string(debug.Stack()), "\n")[line]
}

func Test[t any](actual, expected t) {
	isEqual := reflect.DeepEqual(actual, expected)
	printTest(isEqual, actual, expected, Stack(8))
}

func TestE(err error, packageName string, errorIndex uint) {
	isEqual := ErrorHandler(packageName, errorIndex, err)
	printTest(isEqual, err, errorPrefix(packageName, errorIndex), Stack(8))
}

func printTest(isEqual bool, actual, expected any, stack string) {
	if !isEqual {
		fmt.Println(ColorRed, stack, ColorReset)
		fmt.Printf("%v%#v\n", ColorYellow, actual)
		fmt.Printf("%v%#v%v\n", ColorBlue, expected, ColorReset)

		if PanicIfError {
			os.Exit(1)
		}
		return
	}
	fmt.Println(ColorGreen, stack, ColorReset)
}

func Rand[t any]() t {
	var result t
	fuzz.New().Fuzz(&result)
	time.Sleep(1 * time.Microsecond)
	return result
}

type t time.Time

func Time() t {
	return t(time.Now())
}
func (s t) Print() {
	fmt.Println(ColorBlue, Stack(8), ColorReset)
	fmt.Printf("it takes %v\n", time.Since(time.Time(s)))
}

func Append[t any](l ...[]t) []t {
	var l2 []t
	for _, v := range l {
		l2 = append(l2, v...)
	}
	return l2
}
