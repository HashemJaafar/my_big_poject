package user

import (
	"crypto/rand"
	"determinants"
	"fmt"
	db "my_database"
	"testing"
	"time"
	"tools"
)

func TestMain(m *testing.M) {
	db.Open(&userIdAndHash, determinants.DBPath("userIdAndHash"))
	defer userIdAndHash.Close()

	m.Run()
}
func TestCheckIfUserIdAndPasswordCorrect(t *testing.T) {
	_, a := GetHash.Process(Id{})
	tools.TestE(a, packageName, 2)

	_, a = Check.Process(ReqTCheck{Id{}, ""})
	tools.TestE(a, packageName, 2)

	p := make([]byte, 3)
	rand.Read(p)

	password := Password(p)
	userId, _ := Create.Process(password)

	_, a = Check.Process(ReqTCheck{userId, ""})
	tools.TestE(a, packageName, 1)

	_, a = Check.Process(ReqTCheck{userId, password})
	tools.Test(a, nil)

	_, a = GetHash.Process(userId)
	tools.Test(a, nil)
}

func BenchmarkCheckIfUserIdAndPasswordCorrect(b *testing.B) {
	for i := 0; i < 1000; i++ {
		TestCheckIfUserIdAndPasswordCorrect(&testing.T{})
	}
}

func TestGuessPassword(t *testing.T) {
	go Server()
	time.Sleep(1000 * time.Millisecond)

	{
		e1 := Password("")
		i1 := Id{}
		i2 := []string{"Hh", "Aa", "Ss", "Hh", "Ee", "Mm"}
		hash, _ := GetHash.Request(i1)
		a1, err := GuessPassword(i1, hash, i2)
		tools.Test(a1, e1)
		tools.TestE(err, packageName, 3)
	}
	{
		e1 := Password("hashem")
		i1, _ := Create.Request("hashem")
		i2 := []string{"Hh", "", "Aa", "Ss", "Hh", "Ee", "Mm"}
		hash, _ := GetHash.Request(i1)
		a1, err := GuessPassword(i1, hash, i2)
		tools.Test(a1, e1)
		tools.Test(err, nil)
	}
	{
		e1 := Password("hashe")
		i1, _ := Create.Request("hashe")
		i2 := []string{"Hh", "Aa", "Ss", "Hh", "Ee", ""}
		hash, _ := GetHash.Request(i1)
		a1, err := GuessPassword(i1, hash, i2)
		tools.Test(a1, e1)
		tools.Test(err, nil)
	}
	{
		e1 := Password("")
		i1 := Id{}
		i2 := []string{"Hh", "Aa", "Ss", "Hh", "Ee", ""}
		hash, _ := GetHash.Request(i1)
		a1, err := GuessPassword(i1, hash, i2)
		tools.Test(a1, e1)
		tools.TestE(err, packageName, 3)
	}
	{
		e1 := Password("")
		i1, _ := Create.Request("hashem")
		i2 := []string{"Hh", "A", "Ss", "Hh", "Ee", "Mm"}
		hash, _ := GetHash.Request(i1)
		a1, err := GuessPassword(i1, hash, i2)
		tools.Test(a1, e1)
		tools.TestE(err, packageName, 3)
	}
}

func TestServer(t *testing.T) {
	go Server()
	time.Sleep(1000 * time.Millisecond)

	id, err := Create.Request(Password("123"))
	tools.Test(err, nil)

	_, err = Check.Request(ReqTCheck{id, Password("123")})
	tools.Test(err, nil)

	_, err = Check.Request(ReqTCheck{id, Password("1")})
	tools.TestE(err, packageName, 1)

	_, err = ChangePassword.Request(ReqTChangePassword{id, Password("1"), Password("1234")})
	tools.TestE(err, packageName, 1)

	_, err = ChangePassword.Request(ReqTChangePassword{id, Password("123"), Password("1234")})
	tools.Test(err, nil)

	_, err = Check.Request(ReqTCheck{id, Password("123")})
	tools.TestE(err, packageName, 1)

	_, err = Check.Request(ReqTCheck{id, Password("1234")})
	tools.Test(err, nil)

	_, err = GetHash.Request(id)
	tools.Test(err, nil)

	hash, err := GetHash.Request(Id{})
	tools.Test(hash, Hash{})
	tools.TestE(err, packageName, 2)

	hash, err = GetHash.Request(id)
	tools.Test(err, nil)

	password, err := GuessPassword(id, hash, []string{"1k", "23", "13", "", "14"})
	tools.Test(password, "1234")
	tools.Test(err, nil)

	password, err = GuessPassword(id, hash, []string{"7", "78", "ll"})
	tools.Test(password, "")
	tools.TestE(err, packageName, 3)

	password, err = GuessPassword(Id{}, hash, []string{"7", "78", "ll"})
	tools.Test(password, "")
	tools.TestE(err, packageName, 3)
}

func Test(t *testing.T) {
	for i := 0; i < 1000; i++ {
		TestCheckIfUserIdAndPasswordCorrect(&testing.T{})
	}

	i := 0
	db.View(userIdAndHash, func(key, value []byte) {
		i++
		fmt.Printf("%v\t%x\t%x\n", i, key, value)
	})
}

// func Test1(t *testing.T) {
// 	privateKey, publicKey := GenerateKey()
// 	address := CreateAddress(publicKey)

// 	StoreKey(privateKey, "1234")

// 	key, err := GetKey(address, "1234")
// 	tools.Println(key, err)

// 	err = ChangePassword(address, "1234", "12345")
// 	tools.Println(err)

// 	key, err = GetKey(address, "1234")
// 	tools.Println(key, err)

// 	key, err = GetKey(address, "12345")
// 	tools.Println(key, err)

// 	text := []byte("hashem")

// 	ciphertext, err := Encrypt(publicKey, text)
// 	tools.Println(err)

// 	text, err = Decrypt(privateKey, ciphertext)
// 	tools.Println(string(text), err)

// 	signature, err := CreateSignature(privateKey, sha512.Sum512(text))
// 	tools.Println(err)

// 	err = VerifiSignature(publicKey, sha512.Sum512(text), signature)
// 	tools.Println(err)

// }

// func Test2(t *testing.T) {
// 	c := encrypt("hi", []byte("hashem"))
// 	text, err := decrypt("hi", c)
// 	tools.Println(string(text), err)
// 	text, err = decrypt("h", c)
// 	tools.Println(string(text), err)
// }
// func Test3(t *testing.T) {
// 	_, publicKey := GenerateKey()
// 	s := tools.Time()
// 	fmt.Printf("%x\n", CreateAddress(publicKey))
// 	s.Print()
// }
