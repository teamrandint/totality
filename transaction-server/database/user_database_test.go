package database

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

func TestAddUser(t *testing.T) {
	db := RedisDatabase{"tcp", ":6379"}
	_, err := db.GetUserInfo("AAA")
	if err != nil {
		t.Error(err)
	}
}

func TestAddFunds(t *testing.T) {
	db := RedisDatabase{"tcp", ":6379"}
	dollar, err := decimal.NewFromString("23.01")
	err2 := db.AddFunds("AAA", dollar)
	if err != nil || err2 != nil {
		t.Error(err, err2)
	}
	db.DeleteKey("AAA")
}

func TestGetUserInfo(t *testing.T) {
	db := RedisDatabase{"tcp", ":6379"}
	dollar, _ := decimal.NewFromString("23.01")
	db.AddFunds("AAA", dollar)
	r, error := db.GetUserInfo("AAA")
	if error != nil {
		t.Error(error)
	} else {
		fmt.Println(r)
	}
	db.DeleteKey("AAA:Balance")
}

func TestRemoveFunds(t *testing.T) {
	db := RedisDatabase{"tcp", ":6379"}
	dollar, err := decimal.NewFromString("23.01")
	err2 := db.AddFunds("F", dollar)
	if err != nil || err2 != nil {
		t.Error(err, err2)
	}
	err = db.RemoveFunds("F", dollar)
	zero, _ := db.GetFunds("F")

	if zero.String() != "0" {
		t.Error("Account should be 0")
	}
	db.DeleteKey("F:Balance")
}

func TestGetFunds(t *testing.T) {
	db := RedisDatabase{"tcp", ":6379"}
	dollar, err := decimal.NewFromString("23.01")

	err2 := db.AddFunds("fundGetter", dollar)
	amount, err2 := db.GetFunds("fundGetter")

	if err != nil || err2 != nil {
		t.Error(err, err2)
	}

	if amount.String() != dollar.String() {
		t.Error("Amounts not equal, 23.01,", amount)
	}
	db.DeleteKey("fundGetter:Balance")
}

func TestStocks(t *testing.T) {
	db := RedisDatabase{"tcp", ":6379"}
	db.AddStock("F", "stockname", decimal.NewFromFloat(22.00))

	amt, err := db.GetStock("F", "stockname")
	if !amt.Equals(decimal.NewFromFloat(22.00)) {
		t.Error("Wrong value for stocks, should be 22, is ", amt)
	}
	if err != nil {
		t.Error(err)
	}

	amt, err = db.GetStock("F", "wrongstockname")
	if !amt.Equals(decimal.NewFromFloat(0.0)) {
		t.Error("Should get no value for stocks")
	}

	err = db.RemoveStock("F", "stockname", decimal.NewFromFloat(2.0))
	if err != nil {
		t.Error(err)
	}

	amt, err = db.GetStock("F", "stockname")
	if !amt.Equals(decimal.NewFromFloat(20.0)) {
		t.Error("Failed to remove stock")
	} else if err != nil {
		t.Error(err)
	}

	db.DeleteKey("F:Stocks")
}

func TestOrders(t *testing.T) {
	db := RedisDatabase{"tcp", ":6379"}
	err := db.PushSell("SELLER", "AAA", decimal.NewFromFloat(11.11), decimal.NewFromFloat(3))
	if err != nil {
		t.Error(err)
	}
	err = db.PushSell("SELLER", "BBB", decimal.NewFromFloat(11.11), decimal.NewFromFloat(3))
	if err != nil {
		t.Error(err)
	}

}
