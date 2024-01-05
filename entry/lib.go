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
	tools.PanicIfNotNil(err)
	for _, v := range entriesId {
		value, err := db.Get(entryDB, v[:])
		if err != nil {
			log.Panicln("this error should not accurr")
		}

		entry, err := tools.Decode[t](value)
		tools.PanicIfNotNil(err)
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
	tools.PanicIfNotNil(err)
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

	var tripleEntry1 tUserIdTokenIdBalance
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
			return tools.Errorf(packageName, 2, "the userId:%v have total value=%v and this is not equal to zero", userId, totalValue)
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

func storeTheEntryIfTokensAcceptsAndTheSignIsSame(entryId entryStep2Idt, entry entryStep2) error {
	var err error

	err = checkIfTheSignOfQuantityEqualToSignOfValueAfterTheEntryIsAdded(entry.TripleEntry)
	if err != nil {
		entryStep2StoreF(entryId, entry)
		return err
	}

	err = checkTheTokensIfAccepts(entry.Accepts, entry.TripleEntry)
	if err != nil {
		entryStep2StoreF(entryId, entry)
		return err
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
	return nil
}

func entryStep2StoreF(entryId entryStep2Idt, entry entryStep2) {
	entryAsBytes := tools.Encode(entry)
	db.Update(gvEntryIdEntryStep2DataBase, []byte(entryId[:]), entryAsBytes)
}

func entryStep3StoreF(entryId entryStep2Idt, entry entryStep2) {

	entryStep3Encoded := tools.Encode(entry)
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
				tools.PanicIfNotNil(err)
			}

			index, isIn := tools.Find(entryId, entryStep2IdSlice)
			if isIn {
				entryStep2IdSlice = tools.DeleteAtIndex(entryStep2IdSlice, index)
				entryStep2IdSliceEncoded := tools.Encode(entryStep2IdSlice)
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
				tools.PanicIfNotNil(err)
			}

			entryStep3IdSlice = append(entryStep3IdSlice, entryStep3Idv)
			entryStep3IdSliceEncoded := tools.Encode(entryStep3IdSlice)
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

	db.Read(gvEntryIdEntryStep3DataBase, func(key, value []byte) {
		wait.Add(1)
		go func() {
			d, err := tools.Decode[entryStep3](value)
			tools.PanicIfNotNil(err)
			storeEntryToVariables(d)
			wait.Done()
		}()
	})
	wait.Wait()

	go algorithmToCloseAllDebit()

	mux := http.NewServeMux()

	ht.HandleFunc(mux, AddEntry.Pattern, AddEntry.Handle)
	ht.HandleFunc(mux, AcceptEntry.Pattern, AcceptEntry.Handle)
	ht.HandleFunc(mux, RemoveTheAccept.Pattern, RemoveTheAccept.Handle)
	ht.HandleFunc(mux, EditEntry.Pattern, EditEntry.Handle)
	ht.HandleFunc(mux, GetEntryStep2.Pattern, GetEntryStep2.Handle)
	ht.HandleFunc(mux, GetEntryStep3.Pattern, GetEntryStep3.Handle)
	ht.HandleFunc(mux, GetAllEntryStep2.Pattern, GetAllEntryStep2.Handle)
	ht.HandleFunc(mux, GetAllEntryStep3.Pattern, GetAllEntryStep3.Handle)
	ht.HandleFunc(mux, GetAllEntryStep3Accounting.Pattern, GetAllEntryStep3Accounting.Handle)
	ht.HandleFunc(mux, GetToken.Pattern, GetToken.Handle)
	ht.HandleFunc(mux, CreateToken.Pattern, CreateToken.Handle)
	ht.HandleFunc(mux, UserIdBalances.Pattern, UserIdBalances.Handle)
	ht.HandleFunc(mux, TokenIdBalances.Pattern, TokenIdBalances.Handle)
	ht.HandleFunc(mux, UserIdTokenIdBalance.Pattern, UserIdTokenIdBalance.Handle)
	ht.HandleFunc(mux, TokenIdTotalBalance.Pattern, TokenIdTotalBalance.Handle)
	ht.HandleFunc(mux, LastTradeTable.Pattern, LastTradeTable.Handle)

	ht.ListenAndServe(mux, host, port)
}

type ReqTAddEntry struct {
	Entry1  entryStep1
	Accepts []user.ReqTCheck
}

var AddEntry = ht.Create[ReqTAddEntry, any](host, port, "/AddEntry", func(req ReqTAddEntry) (any, error) {
	accepts := req.Accepts
	entry1 := req.Entry1

	var err error

	err = checkTheAcceptsIsValid(accepts)
	if err != nil {
		return nil, err
	}

	setTheUserIdSentTo(&entry1)
	entry2 := encodeEntryStep1To2(entry1)
	addTheAccepts(entry2.Accepts, accepts)

	isHeAccept := isTheWriterAccept(entry2.Writer, entry2.Accepts)
	if !isHeAccept {
		return nil, tools.Errorf(packageName, 4, "the writer did not accept")
	}

	err = MakeOfflineAccountingCheck(&entry2.TripleEntry)
	if err != nil {
		return nil, err
	}

	err = checkUserIdIfExistF(entry2)
	if err != nil {
		return nil, err
	}

	err = checkTokenIdIfExistF(entry1.TripleEntry)
	if err != nil {
		return nil, err
	}

	entryId := db.NewKey(gvEntryIdEntryStep2DataBase, determinants.Len256Bit)
	err = storeTheEntryIfTokensAcceptsAndTheSignIsSame(entryStep2Idt(entryId), encodeEntryStep1To2(entry1))
	if err != nil {
		return nil, err
	}

	return nil, nil
})

type ReqTEntryIdAndAccepts struct {
	EntryId entryStep2Idt
	Accepts []user.ReqTCheck
}

var AcceptEntry = ht.Create[ReqTEntryIdAndAccepts, any](host, port, "/AcceptEntry", func(req ReqTEntryIdAndAccepts) (any, error) {
	accepts := req.Accepts
	entryId := req.EntryId

	var err error

	err = checkTheAcceptsIsValid(accepts)
	if err != nil {
		return nil, err
	}

	entry, err := entryStep2GetF(entryId)
	if err != nil {
		return nil, err
	}

	addTheAccepts(entry.Accepts, accepts)

	err = storeTheEntryIfTokensAcceptsAndTheSignIsSame(entryId, entry)
	if err != nil {
		return nil, err
	}

	return nil, nil
})

var RemoveTheAccept = ht.Create[ReqTEntryIdAndAccepts, any](host, port, "/RemoveTheAccept", func(req ReqTEntryIdAndAccepts) (any, error) {
	accepts := req.Accepts
	entryId := req.EntryId

	var err error

	err = checkTheAcceptsIsValid(accepts)
	if err != nil {
		return nil, err
	}

	entry, err := entryStep2GetF(entryId)
	if err != nil {
		return nil, err
	}

	// addTheAccepts(entry.Accepts, accepts, false)

	isHeAccept := isTheWriterAccept(entry.Writer, entry.Accepts)
	if !isHeAccept {
		db.Delete(gvEntryIdEntryStep2DataBase, []byte(entryId[:]))
	} else {
		db.Update(gvEntryIdEntryStep2DataBase, []byte(entryId[:]), tools.Encode(entry))
	}

	return nil, nil
})

type ReqTEditEntry struct {
	EntryId entryStep2Idt
	Writer  user.ReqTCheck
	SendTo  []user.Id
}

var EditEntry = ht.Create[ReqTEditEntry, any](host, port, "/EditEntry", func(req ReqTEditEntry) (any, error) {
	entryId := req.EntryId
	writer := req.Writer

	var err error

	_, err = user.Check.Request(writer)
	if err != nil {
		return nil, err
	}

	entry, err := entryStep2GetF(entryId)
	if err != nil {
		return nil, err
	}

	db.Update(gvEntryIdEntryStep2DataBase, []byte(entryId[:]), tools.Encode(entry))

	return nil, nil
})

var GetEntryStep2 = ht.Create[entryStep2Idt, entryStep2](host, port, "/GetEntryStep2", func(req entryStep2Idt) (entryStep2, error) {
	value, err := db.Get(gvEntryIdEntryStep2DataBase, req[:])
	if err != nil {
		return entryStep2{}, err
	}
	d, err := tools.Decode[entryStep2](value)
	return d, err
})

var GetEntryStep3 = ht.Create[entryStep3Id, entryStep3](host, port, "/GetEntryStep3", func(req entryStep3Id) (entryStep3, error) {
	value, err := db.Get(gvEntryIdEntryStep3DataBase, req[:])
	if err != nil {
		return entryStep3{}, err
	}
	d, err := tools.Decode[entryStep3](value)
	return d, err
})

var GetAllEntryStep2 = ht.Create[user.Id, []entryStep2](host, port, "/GetAllEntryStep2", func(req user.Id) ([]entryStep2, error) {
	return getAllEntries[entryStep2](user.Id(req), gvUserIdEntryIdStep2DataBase, gvEntryIdEntryStep2DataBase)
})

var GetAllEntryStep3 = ht.Create[user.Id, []entryStep3](host, port, "/GetAllEntryStep3", func(req user.Id) ([]entryStep3, error) {
	return getAllEntries[entryStep3](user.Id(req), gvUserIdEntryIdStep3DataBase, gvEntryIdEntryStep3DataBase)
})

var GetAllEntryStep3Accounting = ht.Create[user.Id, []entryStep3](host, port, "/GetAllEntryStep3Accounting", func(req user.Id) ([]entryStep3, error) {
	return getAllEntries[entryStep3](user.Id(req), gvUserIdEntryIdStep3AccountingDataBase, gvEntryIdEntryStep3DataBase)
})

var GetToken = ht.Create[TokenIdt, Token](host, port, "/GetToken", func(req TokenIdt) (Token, error) {
	value, err := db.Get(gvTokenIdTokenFolderDataBase, req[:])
	if err != nil {
		return Token{}, err
	}
	d, err := tools.Decode[Token](value)
	if err != nil {
		return Token{}, err
	}
	return d, err
})

var CreateToken = ht.Create[Token, TokenIdt](host, port, "/CreateToken", func(req Token) (TokenIdt, error) {
	token := Token(req)
	err := tokenProtocol(token)
	if err != nil {
		return TokenIdt{}, err
	}

	tokenByte := tools.Encode(token)
	hash := tools.Hash64Bit(tokenByte)
	_, err = db.Get(gvTokenIdTokenFolderDataBase, hash[:])
	if err == nil {
		return TokenIdt{}, tools.Errorf(packageName, 6, "this token is have hash %v is already exist , if you whant to store it you should to make changes for the token to change the hash", hash)
	}
	db.Update(gvTokenIdTokenFolderDataBase, hash[:], tokenByte)
	return hash, nil
})

var UserIdBalances = ht.Create[user.Id, map[TokenIdt]Balancet](host, port, "/UserIdBalances", func(req user.Id) (map[TokenIdt]Balancet, error) {
	value, ok := vUserIdTokenIdBalance[user.Id(req)]
	if !ok {
		return nil, tools.Errorf(packageName, 0, "this %v user don't have balance", user.Id(req))
	}
	return value, nil
})

var TokenIdBalances = ht.Create[TokenIdt, map[user.Id]Balancet](host, port, "/TokenIdBalances", func(req TokenIdt) (map[user.Id]Balancet, error) {
	value, ok := vTokenIdUserIdBalance[TokenIdt(req)]
	if !ok {
		return nil, tools.Errorf(packageName, 0, "this %v token don't have balance", TokenIdt(req))
	}
	return value, nil
})

type ReqTUserIdTokenId struct {
	user.Id
	TokenIdt
}

var UserIdTokenIdBalance = ht.Create[ReqTUserIdTokenId, Balancet](host, port, "/UserIdTokenIdBalance", func(req ReqTUserIdTokenId) (Balancet, error) {
	value, ok := vUserIdTokenIdBalance[req.Id][req.TokenIdt]
	if !ok {
		return Balancet{}, tools.Errorf(packageName, 0, "this %v token with this %v user don't have balance", TokenIdt(req.TokenIdt), user.Id(req.Id))
	}
	return value, nil
})

var TokenIdTotalBalance = ht.Create[TokenIdt, Balancet](host, port, "/TokenIdTotalBalance", func(req TokenIdt) (Balancet, error) {
	value, ok := vTokenIdBalance[TokenIdt(req)]
	if !ok {
		return Balancet{}, tools.Errorf(packageName, 0, "this %v token don't have balance", TokenIdt(req))
	}
	return value, nil
})

var LastTradeTable = ht.Create[TokenIdt, []trade](host, port, "/LastTradeTable", func(req TokenIdt) ([]trade, error) {
	value, ok := gvLastTrade[TokenIdt(req)]
	if !ok {
		return []trade{}, tools.Errorf(packageName, 0, "this %v token didn't used in triple entry", TokenIdt(req))
	}
	return value, nil
})
