package entry

import (
	"crypto/sha256"
	"determinants"
	ht "http"
	"log"
	"math"
	db "my_database"
	"net/http"
	"sync"
	"time"

	"tools"
	"user"

	"github.com/samber/lo"
)

const (
	packageName = "entry"
	host        = "localhost"
	port        = 8001
)

var (
	gvEntryIdEntryStep2DataBase            db.DB
	gvUserIdEntryIdStep2DataBase           db.DB
	gvEntryIdEntryStep3DataBase            db.DB
	gvUserIdEntryIdStep3DataBase           db.DB
	gvUserIdEntryIdStep3AccountingDataBase db.DB
	gvTokenIdTokenFolderDataBase           db.DB

	vUserIdTokenIdBalance tUserIdTokenIdBalance
	vTokenIdUserIdBalance map[TokenIdt]map[user.Id]Balancet
	vTokenIdBalance       map[TokenIdt]Balancet
	gvLastTrade           map[TokenIdt][]trade
	vDebits               map[quantityUnitMeasure]map[user.Id]map[TokenIdt]Quantity
)

type (
	tUserIdTokenIdBalance map[user.Id]map[TokenIdt]Balancet

	TokenIdt            [determinants.Len64Bit]byte
	quantityUnitMeasure string
	code                string

	info struct {
		QuantityUnitMeasure quantityUnitMeasure
		Name                string
		AccountType         string
		Description         string
		// Image               [][]byte // as jpeg because is lawer data size
	}

	Token struct {
		Code     code
		Info     info
		MarkDawn []byte
	}

	trade struct {
		time.Time
		Quantity
		Value //this is measured always as 1 gram %100 pure gold
	}

	Quantity float64
	Value    float64

	Balancet struct {
		Quantity
		Value
	}
)

func algorithmToCloseAllDebit() {
	//TODO it should be infinit for loop
	// for {
	// }
}

func numberOfUserInTheEntry(entry []SingleEntry) int {
	users := map[user.Id]bool{}
	for _, v := range entry {
		users[v.UserId] = true
	}
	return len(users)
}

func encodeTripleEntryTotUserIdTokenIdBalance(m tUserIdTokenIdBalance, s *[]SingleEntry) {
	for _, item := range *s {
		if m == nil {
			m = map[user.Id]map[TokenIdt]Balancet{}
		}
		if m[item.UserId] == nil {
			m[item.UserId] = map[TokenIdt]Balancet{}
		}

		x := m[item.UserId][item.TokenId]
		x.Quantity += item.Quantity
		x.Value += item.Value
		m[item.UserId][item.TokenId] = x
	}
}

func getAllEntries[t entryStep2 | entryStep3](userId user.Id, userDB, entryDB db.DB) ([]t, error) {
	value, err := db.Get(userDB, userId[:])
	if err != nil {
		return nil, err
	}
	var entries []t
	entriesId, err := tools.Decode[[]entryStep2Idt](value)
	tools.Panic(err)
	for _, v := range entriesId {
		value, err := db.Get(entryDB, v[:])
		if err != nil {
			log.Panicln("this error should not accurr")
		}

		entry, err := tools.Decode[t](value)
		tools.Panic(err)
		entries = append(entries, entry)
	}
	return entries, nil
}

func tokenProtocol(token Token) error {
	//TODO Check the token code,Check the token info
	return nil
}

func entryStep2GetF(entryId entryStep2Idt) (entryStep2, error) {
	value, err := db.Get(gvEntryIdEntryStep2DataBase, []byte(entryId[:]))
	if err != nil {
		return entryStep2{}, err
	}

	result, err := tools.Decode[entryStep2](value)
	tools.Panic(err)
	return result, nil
}

func checkTheAcceptsIsValid(accepts []user.ReqTCheck) error {
	for _, accept := range accepts {
		_, err := user.Check.Request(accept)
		if err != nil {
			return err
		}
	}
	return nil
}

func addTheAccepts(accepts Accepts, acceptsValidation []user.ReqTCheck) {
	for _, accept := range acceptsValidation {
		accepts = append(accepts, accept.Id)
	}
}

func setTheUserIdSentTo(entry *entryStep1) {
	entry.SendTo = append(entry.SendTo, entry.Writer)
	for _, singleEntry := range entry.TripleEntry {
		entry.SendTo = append(entry.SendTo, singleEntry.UserId)
	}
	entry.SendTo = lo.Uniq(entry.SendTo)
}

func isTheWriterAccept(writer user.Id, usersThatAccepted Accepts) bool {
	_, isHeAccept := tools.Find(writer, usersThatAccepted)
	return isHeAccept
}

func MakeOfflineAccountingCheck(tripleEntry *[]SingleEntry) error {

	if len(*tripleEntry) == 0 {
		return tools.Errorf(packageName, 1, "there is no users, should be one or two")
	}

	tripleEntry1 := tUserIdTokenIdBalance{}
	encodeTripleEntryTotUserIdTokenIdBalance(tripleEntry1, tripleEntry)

	var userId1 user.Id
	var userId2 user.Id
	var value1 Value
	var value2 Value

	totalValue := map[user.Id]Value{}

	userNumber := 0
	*tripleEntry = []SingleEntry{}

	for userId, tokenIdAndBalance := range tripleEntry1 {
		userNumber++
		if userNumber == 3 {
			break
		}
		for tokenId, balance := range tokenIdAndBalance {
			if balance.Quantity == 0 && balance.Value == 0 {
				continue
			}
			balance.Value = Value(tools.ChangeSign(float64(balance.Quantity), float64(balance.Value)))
			*tripleEntry = append(*tripleEntry, SingleEntry{userId, tokenId, balance.Quantity, balance.Value})

			totalValue[userId] += balance.Value

			switch userNumber {
			case 1:
				userId1 = userId
				value1 += Value(math.Abs(float64(balance.Value)))
			case 2:
				userId2 = userId
				value2 += Value(math.Abs(float64(balance.Value)))
			}
		}
	}

	for userId, value := range totalValue {
		if value != 0 {
			return tools.Errorf(packageName, 2, "the userId:%v have total value=%v and this is not equal to zero", userId, value)
		}
	}

	if value1 != value2 {
		return tools.Errorf(packageName, 5, "the first user:%v have total absolute value=%v and second user:%v have total absolute value=%v and the difference is %v and this is not equally", userId1, value1, userId2, value2, math.Abs(float64(value1-value2)))
	}

	return nil
}

func checkUserIdIfExistF(entry entryStep2) error {
	_, err := user.GetHash.Request(entry.Writer)
	if err != nil {
		return err
	}
	for _, userId := range entry.Accepts {
		_, err := user.GetHash.Request(userId)
		if err != nil {
			return err
		}
	}
	for _, userId := range entry.SendTo {
		_, err := user.GetHash.Request(userId)
		if err != nil {
			return err
		}
	}
	for _, v := range entry.TripleEntry {
		_, err := user.GetHash.Request(v.UserId)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkTokenIdIfExistF(entry []SingleEntry) error {
	for _, singleEntry := range entry {
		_, err := GetToken.Process(singleEntry.TokenId)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkIfTheSignOfQuantityEqualToSignOfValueAfterTheEntryIsAdded(tripleEntry []SingleEntry) error {
	for _, singleEntry := range tripleEntry {
		totalQuantityAfterEntry := vUserIdTokenIdBalance[singleEntry.UserId][singleEntry.TokenId].Quantity + singleEntry.Quantity
		totalValueAfterEntry := vUserIdTokenIdBalance[singleEntry.UserId][singleEntry.TokenId].Value + singleEntry.Value

		if !tools.IsSameSign(float64(totalQuantityAfterEntry), float64(totalValueAfterEntry)) {
			return tools.Errorf(packageName, 3, "the sign of quantity should equal the sign of value after the entry is made, see userId:%v with token:%v", singleEntry.UserId, singleEntry.TokenId)
		}
	}
	return nil
}

func runTokenCode(code code, entry []SingleEntry) error {
	//TODO
	return nil
}

func checkTheTokensIfAccepts(usersThatAccepted Accepts, entry []SingleEntry) error {
	for _, singleEntry := range entry {
		token, err := GetToken.Process(singleEntry.TokenId)
		if err != nil {
			return err
		}
		err = runTokenCode(token.Code, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func storeTheEntryIfTokensAcceptsAndTheSignIsSame(entryId entryStep2Idt, entry entryStep2) (entryStep3Id, error) {
	var err error

	err = checkIfTheSignOfQuantityEqualToSignOfValueAfterTheEntryIsAdded(entry.TripleEntry)
	if err != nil {
		entryStep2StoreF(entryId, entry)
		return entryStep3Id{}, err
	}

	err = checkTheTokensIfAccepts(entry.Accepts, entry.TripleEntry)
	if err != nil {
		entryStep2StoreF(entryId, entry)
		return entryStep3Id{}, err
	}

	var wait sync.WaitGroup
	wait.Add(2)

	go func() {
		storeEntryToVariables(encodeEntryStep2To3(entry))
		wait.Done()
	}()

	go func() {
		entryStep3StoreF(entryId, entry)
		wait.Done()
	}()

	wait.Wait()
	return entryStep3Id{}, nil
}

func entryStep2StoreF(entryId entryStep2Idt, entry entryStep2) {
	entryAsBytes, err := tools.Encode(entry)
	tools.Panic(err)
	db.Update(gvEntryIdEntryStep2DataBase, []byte(entryId[:]), entryAsBytes)
}

func entryStep3StoreF(entryId entryStep2Idt, entry entryStep2) {

	entryStep3Encoded, err := tools.Encode(entry)
	tools.Panic(err)
	entryStep3Idv := sha256.Sum256(entryStep3Encoded)
	db.Update(gvEntryIdEntryStep3DataBase, entryStep3Idv[:], entryStep3Encoded)

	userIdShouldStoreTo := map[user.Id]bool{}
	for _, singleEntry := range entry.TripleEntry {
		userIdShouldStoreTo[singleEntry.UserId] = true
	}
	for _, userId := range entry.Accepts {
		userIdShouldStoreTo[userId] = true
	}

	var wait sync.WaitGroup
	wait.Add(2)

	go func() {
		var entryStep2IdSlice []entryStep2Idt
		for userId := range userIdShouldStoreTo {
			value, err := db.Get(gvUserIdEntryIdStep2DataBase, userId[:])
			if err == nil {
				entryStep2IdSlice, err = tools.Decode[[]entryStep2Idt](value)
				tools.Panic(err)
			}

			index, isIn := tools.Find(entryId, entryStep2IdSlice)
			if isIn {
				entryStep2IdSlice = tools.DeleteAtIndex(entryStep2IdSlice, index)
				entryStep2IdSliceEncoded, err := tools.Encode(entryStep2IdSlice)
				tools.Panic(err)
				db.Update(gvUserIdEntryIdStep2DataBase, userId[:], entryStep2IdSliceEncoded)
			}
		}
		db.Delete(gvEntryIdEntryStep2DataBase, entryId[:])
		wait.Done()
	}()

	go func() {
		var entryStep3IdSlice []entryStep3Id
		for userId := range userIdShouldStoreTo {
			value, err := db.Get(gvUserIdEntryIdStep3DataBase, userId[:])
			if err == nil {
				entryStep3IdSlice, err = tools.Decode[[]entryStep3Id](value)
				tools.Panic(err)
			}

			entryStep3IdSlice = append(entryStep3IdSlice, entryStep3Idv)
			entryStep3IdSliceEncoded, err := tools.Encode(entryStep3IdSlice)
			tools.Panic(err)
			db.Update(gvUserIdEntryIdStep3DataBase, userId[:], entryStep3IdSliceEncoded)
		}
		wait.Done()
	}()

	wait.Wait()
}

func storeEntryToVariables(entry entryStep3) {
	isTwoUser := numberOfUserInTheEntry(entry.TripleEntry) == 2

	encodeTripleEntryTotUserIdTokenIdBalance(vUserIdTokenIdBalance, &entry.TripleEntry)

	for _, v := range entry.TripleEntry {
		{
			if vTokenIdUserIdBalance == nil {
				vTokenIdUserIdBalance = map[TokenIdt]map[user.Id]Balancet{}
			}
			if vTokenIdUserIdBalance[v.TokenId] == nil {
				vTokenIdUserIdBalance[v.TokenId] = map[user.Id]Balancet{}
			}

			x := vTokenIdUserIdBalance[v.TokenId][v.UserId]
			x.Quantity += v.Quantity
			x.Value += v.Value
			vTokenIdUserIdBalance[v.TokenId][v.UserId] = x
		}

		{
			if vTokenIdBalance == nil {
				vTokenIdBalance = map[TokenIdt]Balancet{}
			}

			x := vTokenIdBalance[v.TokenId]
			x.Quantity += v.Quantity
			x.Value += v.Value
			vTokenIdBalance[v.TokenId] = x
		}

		{
			if isTwoUser {
				if gvLastTrade == nil {
					gvLastTrade = map[TokenIdt][]trade{}
				}

				x := gvLastTrade[v.TokenId]
				x = append(x, trade{entry.Time, v.Quantity, v.Value})
				gvLastTrade[v.TokenId] = x
			}
		}

		{
			//TODO
			token, _ := GetToken.Process(v.TokenId)
			if vDebits == nil {
				vDebits = map[quantityUnitMeasure]map[user.Id]map[TokenIdt]Quantity{}
			}
			if vDebits[token.Info.QuantityUnitMeasure] == nil {
				vDebits[token.Info.QuantityUnitMeasure] = map[user.Id]map[TokenIdt]Quantity{}
			}
			if vDebits[token.Info.QuantityUnitMeasure][v.UserId] == nil {
				vDebits[token.Info.QuantityUnitMeasure][v.UserId] = map[TokenIdt]Quantity{}
			}

			vDebits[token.Info.QuantityUnitMeasure][v.UserId][v.TokenId] += v.Quantity
		}
	}
}

type (
	userIdSendTot []user.Id
	Accepts       []user.Id

	SingleEntry struct {
		UserId   user.Id
		TokenId  TokenIdt
		Quantity Quantity
		Value    Value
	}

	entryStep1 struct {
		Writer      user.Id
		TripleEntry []SingleEntry
		SendTo      userIdSendTot
	}

	entryStep2Idt [determinants.Len64Bit]byte

	entryStep2 struct {
		Writer      user.Id
		TripleEntry []SingleEntry
		SendTo      userIdSendTot
		Accepts     Accepts
		Time        time.Time
	}

	entryStep3Id [determinants.Len256Bit]byte

	entryStep3 struct {
		Writer      user.Id
		TripleEntry []SingleEntry
		Accepts     Accepts
		Time        time.Time
	}
)

func encodeEntryStep1To2(decoded entryStep1) entryStep2 {
	return entryStep2{
		Writer:      decoded.Writer,
		TripleEntry: decoded.TripleEntry,
		SendTo:      decoded.SendTo,
		Accepts:     Accepts{},
		Time:        time.Now().UTC(),
	}
}

func encodeEntryStep2To3(decoded entryStep2) entryStep3 {
	return entryStep3{
		Writer:      decoded.Writer,
		TripleEntry: decoded.TripleEntry,
		Accepts:     Accepts{},
		Time:        time.Now().UTC(),
	}
}

func Server() {
	var wait sync.WaitGroup
	wait.Add(6)

	open := func(database *db.DB, path string) {
		db.Open(database, determinants.DBPath(path))
		wait.Done()
	}

	go open(&gvEntryIdEntryStep2DataBase, "gvEntryIdEntryStep2DataBase")
	go open(&gvUserIdEntryIdStep2DataBase, "gvUserIdEntryIdStep2DataBase")
	go open(&gvEntryIdEntryStep3DataBase, "gvEntryIdEntryStep3DataBase")
	go open(&gvUserIdEntryIdStep3DataBase, "gvUserIdEntryIdStep3DataBase")
	go open(&gvUserIdEntryIdStep3AccountingDataBase, "gvUserIdEntryIdStep3AccountingDataBase")
	go open(&gvTokenIdTokenFolderDataBase, "gvTokenIdTokenFolderDataBase")

	wait.Wait()

	defer gvEntryIdEntryStep2DataBase.Close()
	defer gvUserIdEntryIdStep2DataBase.Close()
	defer gvEntryIdEntryStep3DataBase.Close()
	defer gvUserIdEntryIdStep3DataBase.Close()
	defer gvUserIdEntryIdStep3AccountingDataBase.Close()
	defer gvTokenIdTokenFolderDataBase.Close()

	db.View(gvEntryIdEntryStep3DataBase, func(key, value []byte) {
		wait.Add(1)
		go func() {
			d, err := tools.Decode[entryStep3](value)
			tools.Panic(err)
			storeEntryToVariables(d)
			wait.Done()
		}()
	})
	wait.Wait()

	go algorithmToCloseAllDebit()

	mux := http.NewServeMux()

	AddEntry.Handle(mux)
	AcceptEntry.Handle(mux)
	RemoveTheAccept.Handle(mux)
	EditEntry.Handle(mux)
	GetEntryStep2.Handle(mux)
	GetEntryStep3.Handle(mux)
	GetAllEntryStep2.Handle(mux)
	GetAllEntryStep3.Handle(mux)
	GetAllEntryStep3Accounting.Handle(mux)
	GetToken.Handle(mux)
	CreateToken.Handle(mux)
	UserIdBalances.Handle(mux)
	TokenIdBalances.Handle(mux)
	UserIdTokenIdBalance.Handle(mux)
	TokenIdTotalBalance.Handle(mux)
	LastTradeTable.Handle(mux)

	ht.ListenAndServe(mux, host, port)
}

type ReqTAddEntry struct {
	Entry1  entryStep1
	Accepts []user.ReqTCheck
}

var AddEntry = ht.Create(host, port, "/AddEntry", func(req ReqTAddEntry) (entryStep3Id, error) {
	accepts := req.Accepts
	entry1 := req.Entry1

	var err error

	err = checkTheAcceptsIsValid(accepts)
	if err != nil {
		return entryStep3Id{}, err
	}

	setTheUserIdSentTo(&entry1)
	entry2 := encodeEntryStep1To2(entry1)
	addTheAccepts(entry2.Accepts, accepts)

	isHeAccept := isTheWriterAccept(entry2.Writer, entry2.Accepts)
	if !isHeAccept {
		return entryStep3Id{}, tools.Errorf(packageName, 4, "the writer did not accept")
	}

	err = MakeOfflineAccountingCheck(&entry2.TripleEntry)
	if err != nil {
		return entryStep3Id{}, err
	}

	err = checkUserIdIfExistF(entry2)
	if err != nil {
		return entryStep3Id{}, err
	}

	err = checkTokenIdIfExistF(entry1.TripleEntry)
	if err != nil {
		return entryStep3Id{}, err
	}

	entryId := db.New64BitKey(gvEntryIdEntryStep2DataBase)
	entryId3, err := storeTheEntryIfTokensAcceptsAndTheSignIsSame(entryStep2Idt(entryId), encodeEntryStep1To2(entry1))
	if err != nil {
		return entryStep3Id{}, err
	}

	return entryId3, nil
})

type ReqTEntryIdAndAccepts struct {
	EntryId entryStep2Idt
	Accepts []user.ReqTCheck
}

var AcceptEntry = ht.Create(host, port, "/AcceptEntry", func(req ReqTEntryIdAndAccepts) (entryStep3Id, error) {
	accepts := req.Accepts
	entryId := req.EntryId

	var err error

	err = checkTheAcceptsIsValid(accepts)
	if err != nil {
		return entryStep3Id{}, err
	}

	entry, err := entryStep2GetF(entryId)
	if err != nil {
		return entryStep3Id{}, err
	}

	addTheAccepts(entry.Accepts, accepts)

	entryId3, err := storeTheEntryIfTokensAcceptsAndTheSignIsSame(entryId, entry)
	if err != nil {
		return entryStep3Id{}, err
	}

	return entryId3, nil
})

var RemoveTheAccept = ht.Create(host, port, "/RemoveTheAccept", func(req ReqTEntryIdAndAccepts) (ht.Useless, error) {
	accepts := req.Accepts
	entryId := req.EntryId

	var err error

	err = checkTheAcceptsIsValid(accepts)
	if err != nil {
		return ht.Useless{}, err
	}

	entry, err := entryStep2GetF(entryId)
	if err != nil {
		return ht.Useless{}, err
	}

	// addTheAccepts(entry.Accepts, accepts, false)

	isHeAccept := isTheWriterAccept(entry.Writer, entry.Accepts)
	if !isHeAccept {
		db.Delete(gvEntryIdEntryStep2DataBase, []byte(entryId[:]))
	} else {
		e, err := tools.Encode(entry)
		tools.Panic(err)
		db.Update(gvEntryIdEntryStep2DataBase, []byte(entryId[:]), e)
	}

	return ht.Useless{}, nil
})

type ReqTEditEntry struct {
	EntryId entryStep2Idt
	Writer  user.ReqTCheck
	SendTo  []user.Id
}

var EditEntry = ht.Create(host, port, "/EditEntry", func(req ReqTEditEntry) (ht.Useless, error) {
	entryId := req.EntryId
	writer := req.Writer

	var err error

	_, err = user.Check.Request(writer)
	if err != nil {
		return ht.Useless{}, err
	}

	entry, err := entryStep2GetF(entryId)
	if err != nil {
		return ht.Useless{}, err
	}

	e, err := tools.Encode(entry)
	tools.Panic(err)
	db.Update(gvEntryIdEntryStep2DataBase, []byte(entryId[:]), e)

	return ht.Useless{}, nil
})

var GetEntryStep2 = ht.Create(host, port, "/GetEntryStep2", func(req entryStep2Idt) (entryStep2, error) {
	value, err := db.Get(gvEntryIdEntryStep2DataBase, req[:])
	if err != nil {
		return entryStep2{}, err
	}
	d, err := tools.Decode[entryStep2](value)
	if err != nil {
		return entryStep2{}, err
	}
	return d, nil
})

var GetEntryStep3 = ht.Create(host, port, "/GetEntryStep3", func(req entryStep3Id) (entryStep3, error) {
	value, err := db.Get(gvEntryIdEntryStep3DataBase, req[:])
	if err != nil {
		return entryStep3{}, err
	}
	d, err := tools.Decode[entryStep3](value)
	if err != nil {
		return entryStep3{}, err
	}
	return d, nil
})

var GetAllEntryStep2 = ht.Create(host, port, "/GetAllEntryStep2", func(req user.Id) ([]entryStep2, error) {
	return getAllEntries[entryStep2](user.Id(req), gvUserIdEntryIdStep2DataBase, gvEntryIdEntryStep2DataBase)
})

var GetAllEntryStep3 = ht.Create(host, port, "/GetAllEntryStep3", func(req user.Id) ([]entryStep3, error) {
	return getAllEntries[entryStep3](user.Id(req), gvUserIdEntryIdStep3DataBase, gvEntryIdEntryStep3DataBase)
})

var GetAllEntryStep3Accounting = ht.Create(host, port, "/GetAllEntryStep3Accounting", func(req user.Id) ([]entryStep3, error) {
	return getAllEntries[entryStep3](user.Id(req), gvUserIdEntryIdStep3AccountingDataBase, gvEntryIdEntryStep3DataBase)
})

var GetToken = ht.Create(host, port, "/GetToken", func(req TokenIdt) (Token, error) {
	value, err := db.Get(gvTokenIdTokenFolderDataBase, req[:])
	if err != nil {
		return Token{}, err
	}
	d, err := tools.Decode[Token](value)
	if err != nil {
		return Token{}, err
	}
	return d, nil
})

var CreateToken = ht.Create(host, port, "/CreateToken", func(req Token) (TokenIdt, error) {
	token := Token(req)
	err := tokenProtocol(token)
	if err != nil {
		return TokenIdt{}, err
	}

	tokenByte, err := tools.Encode(token)
	tools.Panic(err)
	hash := tools.Hash64Bit(tokenByte)
	_, err = db.Get(gvTokenIdTokenFolderDataBase, hash[:])
	if err == nil {
		return TokenIdt{}, tools.Errorf(packageName, 6, "this token is have hash %v is already exist , if you whant to store it you should to make changes for the token to change the hash", hash)
	}
	db.Update(gvTokenIdTokenFolderDataBase, hash[:], tokenByte)
	return hash, nil
})

var UserIdBalances = ht.Create(host, port, "/UserIdBalances", func(req user.Id) (map[TokenIdt]Balancet, error) {
	value, ok := vUserIdTokenIdBalance[user.Id(req)]
	if !ok {
		return nil, tools.Errorf(packageName, 7, "this %v user don't have balance", user.Id(req))
	}
	return value, nil
})

var TokenIdBalances = ht.Create(host, port, "/TokenIdBalances", func(req TokenIdt) (map[user.Id]Balancet, error) {
	value, ok := vTokenIdUserIdBalance[TokenIdt(req)]
	if !ok {
		return nil, tools.Errorf(packageName, 8, "this %v token don't have balance", TokenIdt(req))
	}
	return value, nil
})

type ReqTUserIdTokenId struct {
	user.Id
	TokenIdt
}

var UserIdTokenIdBalance = ht.Create(host, port, "/UserIdTokenIdBalance", func(req ReqTUserIdTokenId) (Balancet, error) {
	value, ok := vUserIdTokenIdBalance[req.Id][req.TokenIdt]
	if !ok {
		return Balancet{}, tools.Errorf(packageName, 9, "this %v token with this %v user don't have balance", TokenIdt(req.TokenIdt), user.Id(req.Id))
	}
	return value, nil
})

var TokenIdTotalBalance = ht.Create(host, port, "/TokenIdTotalBalance", func(req TokenIdt) (Balancet, error) {
	value, ok := vTokenIdBalance[TokenIdt(req)]
	if !ok {
		return Balancet{}, tools.Errorf(packageName, 10, "this %v token don't have balance", TokenIdt(req))
	}
	return value, nil
})

var LastTradeTable = ht.Create(host, port, "/LastTradeTable", func(req TokenIdt) ([]trade, error) {
	value, ok := gvLastTrade[TokenIdt(req)]
	if !ok {
		return []trade{}, tools.Errorf(packageName, 11, "this %v token didn't used in triple entry", TokenIdt(req))
	}
	return value, nil
})

// // /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// type TokenAddress determinants.Sha256
// type token struct {
// 	TokenAddress
// 	Token
// 	// data []byte
// }
// type Contract struct { // that will help me to make decentralized organizations and banks
// 	ContractAddress     user.Address
// 	ContractCodeAndData []byte
// }
// type singleEntry struct {
// 	Address user.Address
// 	TokenAddress
// 	Quantity
// 	Value
// }

// type TripleEntry []singleEntry

// type BigEntry struct {
// 	BalanceEvidence   []byte
// 	Contracts         []Contract
// 	Tokens            []token             // this will help me to prevent database for tokens
// 	TripleEntries     []TripleEntry       // just to know i need it to be slice of triple entry because that will help me to write big entry with more than two person and that will help to close the circle debit
// 	Notes             string              //
// 	Nonce             uint                // this should allways bigger by one from the state to prevent double spending attack
// 	WriterAddress     user.Address        //
// 	PreviousBlockHash determinants.Sha512 // this will help me to make proof of work easy by using proof of Signature and to prevent double spending attack
// 	// Time              time.Time // it work like nonce to prevent double spending attack, but i prefer to remove it to prevent the mutablity as possuble because he can use it as proof of work
// }

// type Accepte struct {
// 	PublicKey rsa.PublicKey
// 	Signature user.Signature // Signature for the acceptor of Hash
// }

// type Document struct { // this ordered by the init sequance
// 	BigEntry      BigEntry
// 	Hash          determinants.Sha512 // this hash should be allways uniqe in all the database
// 	AllTheAccepte []Accepte           // this should be  one-time signature scheme
// }

// type Block struct {
// 	Number                uint
// 	PreviousBlockHash     determinants.Sha512
// 	MerkleRootOfDocuments determinants.Sha512 // this will help me to delete the document to reduce the size of data and to remove the public key just for security
// 	Balances              []byte              // this will stored as binary tree or Merkle Patricia tree or verkle tree and the values is the wallets hashes
// 	Nonce                 []byte
// }

// // Documents         []Document          // this will not hashed with the rest of block it just embeded with merkle tree

// type singleEntry1 struct {
// 	TokenAddress
// 	Quantity
// 	Value
// }

// type wallet struct {
// 	writeNumber uint
// 	data        []byte
// 	b           []singleEntry1
// }
