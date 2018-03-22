package main

import (
	"fmt"
	"seng468/triggerserver/quote"
	"time"

	"github.com/shopspring/decimal"
)

type trigger struct {
	username        string
	stockname       string
	amount          decimal.Decimal
	price           decimal.Decimal
	action          string
	transNum        int
	done            bool
	successListener chan trigger
}

func (t trigger) getSuccessString() string {
	return fmt.Sprintf("TRIGGER_SUCCESS,%v,%v,%v,%v,%v\n",
		t.username, t.stockname, t.price, t.amount, t.action)
}

func (t trigger) getPriceStr() string {
	return t.price.String()
}

func (t trigger) getAmountStr() string {
	return t.amount.String()
}

func (t trigger) String() string {
	str := fmt.Sprintf("{%v %v %v %v %v}", t.username, t.stockname, t.getPriceStr(), t.getAmountStr(), t.action)
	return str
}

func (t trigger) StartPolling() {
	t.done = false
	if t.checkTriggerStatus() {
		successListener <- t
		return
	}

	ticker := time.NewTicker((time.Second * 60) + time.Millisecond)
	defer ticker.Stop()

	for _ = range ticker.C {
		if t.done {
			return
		}
		if t.checkTriggerStatus() {
			successListener <- t
			return
		}
	}
}

func (t trigger) Cancel() {
	t.done = true
	return
}

func (t trigger) checkTriggerStatus() bool {
	result := t.hitQuoteServer()
	return t.checkResult(result)
}

func (t trigger) hitQuoteServer() decimal.Decimal {
	result, err := quoteclient.Query(t.username, t.stockname, t.transNum)
	if err != nil {
		panic(err)
	}

	return result
}

// See if the result from the quoteserver is enough to stop the trigger
func (t trigger) checkResult(result decimal.Decimal) bool {
	switch t.action {
	case "BUY":
		return t.price.GreaterThanOrEqual(result)
	case "SELL":
		return t.price.LessThanOrEqual(result)
	}

	panic("Should never reach here...")
}

func newSellTrigger(sls chan trigger, transNum int, username string, stockname string, amount decimal.Decimal) trigger {
	t := trigger{
		transNum:        transNum,
		username:        username,
		stockname:       stockname,
		amount:          amount,
		action:          "SELL",
		done:            false,
		successListener: sls,
	}

	return t
}

func newBuyTrigger(sls chan trigger, transNum int, username string, stockname string, amount decimal.Decimal) trigger {
	t := trigger{
		transNum:        transNum,
		username:        username,
		stockname:       stockname,
		amount:          amount,
		action:          "BUY",
		done:            false,
		successListener: sls,
	}

	return t
}
