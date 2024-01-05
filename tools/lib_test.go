package tools

import (
	"errors"
	"testing"

	"github.com/samber/lo"
)

func Test_hash64Bit(t *testing.T) {
	got := Hash64Bit([]byte{88})
	Println(got)
}

func Test_isSameSign(t *testing.T) {
	type args struct {
		number1 float64
		number2 float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{1, 1}, true},
		{"", args{1, -1}, false},
		{"", args{-1, 1}, false},
		{"", args{-1, -1}, true},
		{"", args{1, 0}, true},
		{"", args{0, 1}, true},
		{"", args{0, 0}, true},
		{"", args{-1, 0}, true},
		{"", args{0, -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSameSign(tt.args.number1, tt.args.number2); got != tt.want {
				t.Errorf("IsSameSign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Test(t *testing.T) {
	Test(1, 1)
	Test(1, 2)
}

func Test_standardError(t *testing.T) {
	a := errorPrefix("at", 0)
	Test(a, "package (at) error index (0) ")
	a = errorPrefix("a", 1)
	Test(a, "package (a) error index (1) ")

	err := Errorf("at", 0, "the password is uncorrect for user %v", 0)
	Test(err.Error(), "package (at) error index (0) the password is uncorrect for user 0")

	a1 := ErrorHandler("at", 0, err)
	Test(a1, true)
	a1 = ErrorHandler("at", 1, err)
	Test(a1, false)
}

func TestEncode(t *testing.T) {
	{
		s := []int{1, 1, 1, 1}
		a, err := Decode[[]int](Encode(s))
		Test(a, s)
		Test(err, nil)
	}
	{
		s := []string{"1, 1, 1, 1", "ksoso"}
		a, err := Decode[[]string](Encode(s))
		Test(a, s)
		Test(err, nil)
	}
	{
		s := "1, 1, 1, 1"
		a, err := Decode[string](Encode(s))
		Test(a, s)
		Test(err, nil)
	}
	{
		type t struct {
			A int
			B string
		}
		s := t{A: 1, B: "lol"}
		a, err := Decode[t](Encode(s))
		Test(a, s)
		Test(err, nil)
	}
	{
		var s error
		a, err := Decode[error](Encode(s))
		Test(a, s)
		Test(err, nil)
	}
	{
		s := errors.New("lol")
		a, err := Decode[error](Encode(s))
		Test(a, s)
		Test(err, nil)
	}
}

func Test1(t *testing.T) {
	names := lo.Uniq[string]([]string{"Samuel", "John", "Samuel"})
	Println(names)
}
