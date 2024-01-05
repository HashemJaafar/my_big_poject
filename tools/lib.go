package tools

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"
	"runtime/debug"
	"strings"
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

func PanicIfNotNil(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

var (
	failTestNumber uint
	PanicIfError   bool = true
)

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

func Encode(decoded any) []byte {
	//use gob in future for fast encode/decode
	encoded, err := json.Marshal(decoded)
	PanicIfNotNil(err)
	return encoded
}

func Decode[t any](encoded []byte) (t, error) {
	//use gob in future for fast encode/decode
	var decoded t
	err := json.Unmarshal(encoded, &decoded)
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
		fmt.Printf("%v%#v\n", ColorBlue, actual)
		fmt.Printf("%v%#v\n", ColorYellow, expected)

		failTestNumber++
		fmt.Println(ColorRed, stack, "\t", failTestNumber, ColorReset)

		if PanicIfError {
			log.Panic()
		}
		return
	}
	fmt.Println(ColorGreen, stack, ColorReset)
}
