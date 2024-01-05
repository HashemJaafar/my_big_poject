package database

import (
	"crypto/rand"
	"tools"

	badger "github.com/dgraph-io/badger/v3"
)

type DB = *badger.DB

func Open(db *DB, path string) {
	var err error
	*db, err = badger.Open(badger.DefaultOptions(path))
	tools.PanicIfNotNil(err)
}

func Delete(db DB, key []byte) {
	db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func Add(db DB, key []byte, value []byte) error {
	_, err := Get(db, key)
	if err == nil {
		return tools.Errorf("my_database", 2, "key %v is used", key)
	}
	Update(db, key, value)
	return nil
}

func Update(db DB, key []byte, value []byte) {
	txn := db.NewTransaction(true)
	defer txn.Commit()
	txn.Set(key, value)
}

func Get(db DB, key []byte) ([]byte, error) {
	var valCopy []byte

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			valCopy = val
			return nil
		})
		return err
	})
	if err != nil && err.Error() == "Key not found" {
		return valCopy, tools.Errorf("my_database", 1, "key %v not found", key)
	}
	return valCopy, err
}

func Read(db DB, function func(key, value []byte)) {
	db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			item.Value(func(value []byte) error {
				function(item.Key(), value)
				return nil
			})
		}
		return nil
	})
}

func NewKey(db DB, keyLen uint) []byte {
	for {
		key := make([]byte, keyLen)
		rand.Read(key)
		_, err := Get(db, key)
		if err != nil {
			return key
		}
	}
}
