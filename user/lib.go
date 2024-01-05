package user

import (
	"crypto/sha256"
	"determinants"
	ht "http"
	db "my_database"
	"net/http"
	"tools"

	"github.com/samber/lo"
)

const (
	packageName = "user"
	host        = "localhost"
	port        = 8000
)

var userIdAndHash db.DB

type (
	Id       [determinants.Len64Bit]byte
	Password string
	Hash     [determinants.Len256Bit]byte
)

func store(id Id, password Password) {
	passwordHash := createHash(id, password)
	db.Update(userIdAndHash, id[:], passwordHash[:])
}

func isHashCorrect(id Id, password Password, passwordHash Hash) bool {
	return createHash(id, password) == passwordHash
}

func createHash(id Id, password Password) Hash {
	return sha256.Sum256(append(id[:], password[:]...))
}

func Server() {
	db.Open(&userIdAndHash, determinants.DBPath("userIdAndHash"))
	defer userIdAndHash.Close()

	mux := http.NewServeMux()

	ht.HandleFunc(mux, Create.Pattern, Create.Handle)
	ht.HandleFunc(mux, Check.Pattern, Check.Handle)
	ht.HandleFunc(mux, ChangePassword.Pattern, ChangePassword.Handle)
	ht.HandleFunc(mux, GetHash.Pattern, GetHash.Handle)

	ht.ListenAndServe(mux, host, port)
}

var Create = ht.Create[Password, Id](host, port, "/Create", func(req Password) (Id, error) {
	password := req

	id := Id(db.NewKey(userIdAndHash, determinants.Len64Bit))
	store(id, password)

	return id, nil
})

type ReqTCheck struct {
	Id       Id
	Password Password
}

var Check = ht.Create[ReqTCheck, any](host, port, "/Check", func(req ReqTCheck) (any, error) {
	id := req.Id
	password := req.Password

	passwordHash, err := GetHash.Process(id)
	if err != nil {
		return nil, err
	}
	if !isHashCorrect(id, password, passwordHash) {
		return nil, tools.Errorf(packageName, 1, "the password is uncorrect for user %v", id)
	}

	return nil, nil
})

type ReqTChangePassword struct {
	Id          Id
	Password    Password
	NewPassword Password
}

var ChangePassword = ht.Create[ReqTChangePassword, any](host, port, "/ChangePassword", func(req ReqTChangePassword) (any, error) {
	id := req.Id
	password := req.Password
	newPassword := req.NewPassword

	_, err := Check.Process(ReqTCheck{id, password})
	if err == nil {
		store(id, newPassword)
	}

	return nil, err
})

var GetHash = ht.Create[Id, Hash](host, port, "/GetHash", func(req Id) (Hash, error) {
	id := req

	value, err := db.Get(userIdAndHash, id[:])
	if err != nil {
		return Hash{}, tools.Errorf(packageName, 2, "the user %v is not exist", id)
	}

	return Hash(value), nil
})

func GuessPassword(id Id, passwordProbabilities []string) (Password, error) {
	passwordHash, err := GetHash.Request(id)
	if err != nil {
		return "", err
	}

	passwordProbabilities = lo.Compact(passwordProbabilities)

	password := make([]rune, len(passwordProbabilities))
	index := 0
	var correctPassword Password

	var findCorrectPassword func(passwordProbabilities []string, password []rune)
	findCorrectPassword = func(passwordProbabilities []string, password []rune) {
		for _, v := range passwordProbabilities[index] {
			if correctPassword != "" {
				return
			}
			password[index] = v

			if index == len(password)-1 {
				if isHashCorrect(id, Password(password), passwordHash) {
					correctPassword = Password(password)
				}
				continue
			}

			index++
			findCorrectPassword(passwordProbabilities, password)
		}
		index--
	}

	findCorrectPassword(passwordProbabilities, password)

	if correctPassword == "" {
		return "", tools.Errorf(packageName, 3, "the password probabilities is uncorrect or not enough")
	}

	return correctPassword, nil
}
