package entry

import (
	"fmt"
	"math"
	"os"
	"testing"
	"text/tabwriter"
	"tools"
	"user"
)

func Print1(v tUserIdTokenIdBalance) {
	fmt.Println(tools.ColorBlue, tools.Stack(8), tools.ColorReset)

	table := tabwriter.NewWriter(os.Stdout, 4, 1, 1, ' ', 0)
	i := 0
	for k1, v1 := range v {
		for k2, v2 := range v1 {
			i++
			fmt.Fprintf(table, "%v%v%v:\t%v\t%v\t%v\t%v\n", tools.ColorYellow, i, tools.ColorReset, k1, k2, v2.Quantity, v2.Value)
		}
	}
	table.Flush()
}

// func Print2(v map[quantityUnitMeasure]map[string]map[string]Quantity) {
// 	fmt.Println(tools.ColorBlue, tools.Stack(8), tools.ColorReset)

// 	table := tabwriter.NewWriter(os.Stdout, 4, 1, 1, ' ', 0)
// 	i := 0
// 	for k1, v1 := range v {
// 		for k2, v2 := range v1 {
// 			for k3, v3 := range v2 {
// 				i++
// 				fmt.Fprintf(table, "%v%v%v:\t%v\t%v\t%v\t%v\n", tools.ColorYellow, i, tools.ColorReset, k1, k2, k3, v3)
// 			}
// 		}
// 	}
// 	table.Flush()
// }

// here start tests/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func TestCheckAndSetAndStoreTheEntryToDataBase(t *testing.T) {
	user1 := user.Id{95, 200, 32, 195, 111, 240, 216, 230}
	password1 := user.Password("password")
	user2 := user.Id{96, 200, 32, 195, 111, 240, 216, 230}

	x := map[user.Id]bool{}
	x[user1] = true
	x[user2] = true

	{
		_, err := AddEntry.Process(ReqTAddEntry{
			Entry1: entryStep1{
				SendTo:      []user.Id{user1},
				Writer:      user1,
				TripleEntry: []SingleEntry{},
			},
			Accepts: []user.ReqTCheck{{
				Id:       user1,
				Password: password1,
			}}})
		tools.TestE(err, packageName, 1)
	}
	{
		user1 = user.Id{60, 229, 153, 6, 218, 169, 101, 179}
		password1 = user.Password("9139256583631120136")
		_, err := AddEntry.Process(ReqTAddEntry{
			Entry1: entryStep1{
				SendTo:      []user.Id{user1},
				Writer:      user1,
				TripleEntry: []SingleEntry{},
			},
			Accepts: []user.ReqTCheck{{
				Id:       user1,
				Password: password1,
			}}})
		tools.Test(err, nil)
	}
}

func TestMakeOfflineAccountingCheck(t *testing.T) {
	{
		a := []SingleEntry{}
		e := []SingleEntry{}
		err := MakeOfflineAccountingCheck(&a)
		tools.TestE(err, packageName, 1)
		tools.Test(a, e)
	}
	{
		a := []SingleEntry{
			{user.Id{}, TokenIdt{}, 0, 0},
			{user.Id{}, TokenIdt{}, 0, 0},
		}
		e := []SingleEntry{}
		err := MakeOfflineAccountingCheck(&a)
		tools.Test(err, nil)
		tools.Test(a, e)
	}
	{
		a := []SingleEntry{
			{user.Id{1}, TokenIdt{}, 0, 1},
			{user.Id{}, TokenIdt{}, 0, 0},
		}
		e := []SingleEntry{
			{user.Id{1}, TokenIdt{}, 0, 1},
		}
		err := MakeOfflineAccountingCheck(&a)
		tools.Test(err, nil)
		tools.Test(a, e)
	}
}

func Test_encodeTripleEntryTotUserIdTokenIdBalance(t *testing.T) {
	m := tUserIdTokenIdBalance{}
	s := []SingleEntry{
		{user.Id{1}, TokenIdt{}, 1, 1},
		{user.Id{1}, TokenIdt{88}, 2, 3},
		{user.Id{1}, TokenIdt{88}, 0, 1},
		{user.Id{}, TokenIdt{88}, 2, 0},
		{user.Id{255, 77, 99, 45, 75, 62}, TokenIdt{1}, 2, 1},
	}
	encodeTripleEntryTotUserIdTokenIdBalance(m, &s)
	Print1(m)
}

func Test_algorithmToCloseAllDebit(t *testing.T) {
	s := []struct {
		quantityUnitMeasure
		user1 string
		user2 string
		Quantity
	}{
		{"dollar", "hashem", "saba", 10},
		{"dollar", "saba", "yasa", 11},
		{"dollar", "yasa", "zaid", 12},
		{"dollar", "zaid", "hashem", 13},
		{"dollar", "zaid", "zozo", 50},
	}

	var smallestNumber Quantity = math.MaxFloat64
	m := map[quantityUnitMeasure]map[string]map[bool]Quantity{}
	for _, v := range s {
		if m[v.quantityUnitMeasure] == nil {
			m[v.quantityUnitMeasure] = map[string]map[bool]Quantity{}
		}
		if m[v.quantityUnitMeasure][v.user1] == nil {
			m[v.quantityUnitMeasure][v.user1] = map[bool]Quantity{}
		}
		if m[v.quantityUnitMeasure][v.user2] == nil {
			m[v.quantityUnitMeasure][v.user2] = map[bool]Quantity{}
		}

		m[v.quantityUnitMeasure][v.user1][true] += v.Quantity
		m[v.quantityUnitMeasure][v.user2][false] -= v.Quantity

		if v.Quantity < smallestNumber {
			smallestNumber = v.Quantity
		}
	}

	for k1, v1 := range m {
		for k2, v2 := range v1 {
			for k3, v3 := range v2 {
				m[k1][k2][k3] = Quantity(math.Abs(float64(v3))) - smallestNumber
			}
		}
	}
	tools.Println(smallestNumber)

	table := tabwriter.NewWriter(os.Stdout, 4, 1, 1, ' ', 0)
	i := 0
	for k1, v1 := range m {
		for k2, v2 := range v1 {
			for k3, v3 := range v2 {
				i++
				fmt.Fprintf(table, "%v%v%v:\t%v\t%v\t%v\t%v\n", tools.ColorYellow, i, tools.ColorReset, k1, k2, k3, v3)
			}
		}
	}
	table.Flush()
}

func Test(t *testing.T) {
	a := tools.Rand[[]SingleEntry]()
	tools.Println(a)
	err := MakeOfflineAccountingCheck(&a)
	tools.Println(a, err)
	// for _, v := range a {
	// 	fmt.Println(v)
	// }
}
