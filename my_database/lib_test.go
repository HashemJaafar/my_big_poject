package database

import (
	"testing"
	"tools"
)

func Test_getValue(t *testing.T) {
	var testDb DB
	Open(&testDb, "./test")
	defer testDb.Close()

	key := []byte{200}
	value := []byte{55}
	Update(testDb, key, value)

	got, err := Get(testDb, key)
	tools.Test(err, nil)
	tools.Test(got, value)

	key = []byte{8}
	got, err = Get(testDb, key)
	tools.TestE(err, "my_database", 1)
	tools.Test(got, []byte(nil))
}

func Test2(t *testing.T) {
	var values [][]byte
	var keys [][]byte

	var testDb DB
	Open(&testDb, "./test")
	Read(testDb, func(key, value []byte) {
		values = append(values, value)
		keys = append(keys, key)
	})
	tools.Println(len(keys))
	tools.Println(len(values))
}
