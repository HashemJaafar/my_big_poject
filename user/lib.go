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
	hash := createHash(id, password)
	db.Update(userIdAndHash, id[:], hash[:])
}

func isHashCorrect(id Id, password Password, hash Hash) bool {
	return createHash(id, password) == hash
}

func createHash(id Id, password Password) Hash {
	return sha256.Sum256(append(id[:], password[:]...))
}

func Server() {
	db.Open(&userIdAndHash, determinants.DBPath("userIdAndHash"))
	defer userIdAndHash.Close()

	mux := http.NewServeMux()

	Create.Handle(mux)
	Check.Handle(mux)
	ChangePassword.Handle(mux)
	GetHash.Handle(mux)

	ht.ListenAndServe(mux, host, port)
}

var Create = ht.Create(host, port, "/Create", func(req Password) (Id, error) {
	password := req

	id := Id(db.New64BitKey(userIdAndHash))
	store(id, password)

	return id, nil
})

type ReqTCheck struct {
	Id       Id
	Password Password
}

var Check = ht.Create(host, port, "/Check", func(req ReqTCheck) (ht.Useless, error) {
	id := req.Id
	password := req.Password

	hash, err := GetHash.Process(id)
	if err != nil {
		return ht.Useless{}, err
	}
	if !isHashCorrect(id, password, hash) {
		return ht.Useless{}, tools.Errorf(packageName, 1, "the password is uncorrect for user %v", id)
	}

	return ht.Useless{}, nil
})

type ReqTChangePassword struct {
	Id          Id
	Password    Password
	NewPassword Password
}

var ChangePassword = ht.Create(host, port, "/ChangePassword", func(req ReqTChangePassword) (ht.Useless, error) {
	id := req.Id
	password := req.Password
	newPassword := req.NewPassword

	_, err := Check.Process(ReqTCheck{id, password})
	if err == nil {
		store(id, newPassword)
	}

	return ht.Useless{}, err
})

var GetHash = ht.Create(host, port, "/GetHash", func(req Id) (Hash, error) {
	id := req

	value, err := db.Get(userIdAndHash, id[:])
	if err != nil {
		return Hash{}, tools.Errorf(packageName, 2, "the user %v is not exist", id)
	}

	return Hash(value), nil
})

func GuessPassword(id Id, hash Hash, passwordProbabilities []string) (Password, error) {
	// hash, err := GetHash.Request(id)
	// if err != nil {
	// 	return "", err
	// }

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
				if isHashCorrect(id, Password(password), hash) {
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
		return "", tools.Errorf(packageName, 3, "the password probabilities is uncorrect or not enough or the user id is uncorrect or the user hash is uncorrect")
	}

	return correctPassword, nil
}

// const (
// 	packageName = "user"
// 	dbName      = "keys"
// )

// var keys db.DB

// type (
// 	Signature []byte
// 	Address   determinants.Sha256
// 	Password  string
// )

// func GenerateKey() (*rsa.PrivateKey, *rsa.PublicKey) {
// 	privateKey, err := rsa.GenerateKey(rand.Reader, 12000)
// 	tools.Panic(err)
// 	publicKey := &privateKey.PublicKey
// 	return privateKey, publicKey
// }

// func CreateAddress(publicKey *rsa.PublicKey) Address {
// 	return sha256.Sum256(publicKey.N.Bytes())
// }

// func Encrypt(publicKey *rsa.PublicKey, text []byte) ([]byte, error) {
// 	return rsa.EncryptOAEP(sha512.New(), rand.Reader, publicKey, text, nil)
// }

// func Decrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
// 	return rsa.DecryptOAEP(sha512.New(), rand.Reader, privateKey, ciphertext, nil)
// }

// func CreateSignature(privateKey *rsa.PrivateKey, hash determinants.Sha512) (Signature, error) {
// 	return rsa.SignPSS(rand.Reader, privateKey, crypto.SHA512, hash[:], nil)
// }

// func VerifiSignature(publicKey *rsa.PublicKey, hash determinants.Sha512, signature Signature) error {
// 	return rsa.VerifyPSS(publicKey, crypto.SHA512, hash[:], signature, nil)
// }

// func StoreKey(privateKey *rsa.PrivateKey, password Password) {
// 	db.Open(&keys, determinants.DBPath(dbName))
// 	defer keys.Close()

// 	key2 := x509.MarshalPKCS1PrivateKey(privateKey)
// 	key3 := encrypt(password, key2)
// 	address := CreateAddress(&privateKey.PublicKey)
// 	db.Update(keys, address[:], key3)
// }

// func ChangePassword(address Address, password Password, newPassword Password) error {
// 	key, err := GetKey(address, password)
// 	if err != nil {
// 		return err
// 	}

// 	StoreKey(key, newPassword)
// 	return nil
// }

// func GetKey(address Address, password Password) (*rsa.PrivateKey, error) {
// 	db.Open(&keys, determinants.DBPath(dbName))
// 	defer keys.Close()

// 	key1, err := db.Get(keys, address[:])
// 	if err != nil {
// 		return nil, err
// 	}

// 	key2, err := decrypt(password, key1)
// 	if err != nil {
// 		return nil, err
// 	}

// 	key3, err := x509.ParsePKCS1PrivateKey(key2)
// 	tools.Panic(err)

// 	return key3, nil
// }

// func GuessPassword(address Address, passwordProbabilities []string) (Password, error) {
// 	db.Open(&keys, determinants.DBPath(dbName))
// 	defer keys.Close()

// 	c, err := db.Get(keys, address[:])
// 	if err != nil {
// 		return "", err
// 	}

// 	passwordProbabilities = lo.Compact(passwordProbabilities)

// 	password := make([]rune, len(passwordProbabilities))
// 	index := 0
// 	var correctPassword Password

// 	var findCorrectPassword func(passwordProbabilities []string, password []rune)
// 	findCorrectPassword = func(passwordProbabilities []string, password []rune) {
// 		for _, v := range passwordProbabilities[index] {
// 			if correctPassword != "" {
// 				return
// 			}
// 			password[index] = v

// 			if index == len(password)-1 {
// 				if _, err := decrypt(Password(password), c); err == nil {
// 					correctPassword = Password(password)
// 				}
// 				continue
// 			}

// 			index++
// 			findCorrectPassword(passwordProbabilities, password)
// 		}
// 		index--
// 	}

// 	findCorrectPassword(passwordProbabilities, password)

// 	if correctPassword == "" {
// 		return "", tools.Errorf(packageName, 3, "the password probabilities is uncorrect or not enough or the user id is uncorrect or the user hash is uncorrect")
// 	}

// 	return correctPassword, nil
// }

// func encrypt(password Password, text []byte) []byte {
// 	key := sha256.Sum256([]byte(password))

// 	c, err := aes.NewCipher(key[:])
// 	tools.Panic(err)

// 	gcm, err := cipher.NewGCM(c)
// 	tools.Panic(err)

// 	nonce := make([]byte, gcm.NonceSize())

// 	_, err = io.ReadFull(rand.Reader, nonce)
// 	tools.Panic(err)

// 	result := gcm.Seal(nonce, nonce, text, nil)

// 	return result
// }

// func decrypt(password Password, ciphertext []byte) ([]byte, error) {
// 	key := sha256.Sum256([]byte(password))

// 	c, err := aes.NewCipher(key[:])
// 	tools.Panic(err)

// 	gcm, err := cipher.NewGCM(c)
// 	tools.Panic(err)

// 	nonceSize := gcm.NonceSize()
// 	if len(ciphertext) < nonceSize {
// 		panic("ciphertext size is less than nonceSize")
// 	}

// 	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
// 	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return plaintext, nil
// }
