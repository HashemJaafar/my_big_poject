package determinants

import "fmt"

const (
	Len64Bit  = 8
	Len256Bit = 32
)

func DBPath(name string) string {
	return fmt.Sprintf("C:/Users/hashem/Desktop/database_test/%v", name)
}
