package determinants

import "fmt"

const (
	Len64Bit  = 8
	Len256Bit = 32
)

type Sha256 [32]byte
type Sha512 [64]byte

func DBPath(name string) string {
	return fmt.Sprintf("C:/Users/hashem/Desktop/database_test/%v", name)
}
