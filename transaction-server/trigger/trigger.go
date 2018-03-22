package triggerclient

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Trigger struct {
	username  string
	stockname string
	amount    decimal.Decimal
	price     decimal.Decimal
	action    string
	transNum  int
}

func (t Trigger) getPriceStr() string {
	return t.price.String()
}

func (t Trigger) getAmountStr() string {
	return t.amount.String()
}

func (t Trigger) String() string {
	str := fmt.Sprintf("{%v %v %v %v %v}", t.username, t.stockname, t.getPriceStr(), t.getAmountStr(), t.action)
	return str
}

func (t Trigger) GetCost() decimal.Decimal {
	return t.amount.Mul(t.price)
}

func (t Trigger) GetAmount() decimal.Decimal {
	return t.amount
}

func (t Trigger) GetPrice() decimal.Decimal {
	return t.price
}

func newSellTrigger(transNum int, username string, stockname string, amount decimal.Decimal) Trigger {
	t := Trigger{
		transNum:  transNum,
		username:  username,
		stockname: stockname,
		amount:    amount,
		action:    "SELL",
	}

	return t
}

func newBuyTrigger(transNum int, username string, stockname string, amount decimal.Decimal) Trigger {
	t := Trigger{
		transNum:  transNum,
		username:  username,
		stockname: stockname,
		amount:    amount,
		action:    "BUY",
	}

	return t
}
