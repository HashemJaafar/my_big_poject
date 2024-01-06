package database

import (
	"testing"
	"tools"
)

var db DB

func TestMain(m *testing.M) {
	Open(&db, "test")
	defer db.Close()

	m.Run()
}
func Test_getValue(t *testing.T) {
	key := []byte{200, 98}
	value := []byte{55}
	Update(db, key, value)

	got, err := Get(db, key)
	tools.Test(err, nil)
	tools.Test(got, value)

	key = []byte{8}
	got, err = Get(db, key)
	tools.TestE(err, packageName, 1)
	tools.Test(got, nil)
}

func Test2(t *testing.T) {
	var values [][]byte
	var keys [][]byte

	View(db, func(key, value []byte) {
		values = append(values, value)
		keys = append(keys, key)
	})
	tools.Println(len(keys), len(values))
}
