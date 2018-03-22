package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/shopspring/decimal"
)

// UserDatabase holds all of the supported database commands
type UserDatabase interface {
	GetUserInfo(user string) (info string, err error)

	AddFunds(string, decimal.Decimal) error
	GetFunds(string) (decimal.Decimal, error)
	RemoveFunds(string, decimal.Decimal) error

	AddStock(user string, stock string, shares decimal.Decimal) error
	GetStock(user string, stock string) (decimal.Decimal, error)
	RemoveStock(user string, stock string, amount decimal.Decimal) error

	AddReserveFunds(string, decimal.Decimal) error
	GetReserveFunds(string) (decimal.Decimal, error)
	RemoveReserveFunds(string, decimal.Decimal) error

	AddReserveStock(user string, stock string, shares decimal.Decimal) error
	GetReserveStock(user string, stock string) (decimal.Decimal, error)
	RemoveReserveStock(user string, stock string, amount decimal.Decimal) error

	AddSellTrigger(user string, stock string, shares decimal.Decimal) error
	RemoveSellTrigger(user string, stock string, shares decimal.Decimal) error
	AddBuyTrigger(user string, stock string, amount decimal.Decimal) error
	RemoveBuyTrigger(user string, stock string) error

	PushBuy(user string, stock string, cost decimal.Decimal, shares decimal.Decimal) error
	PopBuy(user string) (stock string, cost decimal.Decimal, shares decimal.Decimal, err error)
	PushSell(user string, stock string, cost decimal.Decimal, shares decimal.Decimal) error
	PopSell(user string) (stock string, cost decimal.Decimal, shares decimal.Decimal, err error)

	DbRequestWorker()
	MakeDbRequests([]*Query)
}

// Typical structure of a redis command
type Query struct {
	Command    string
	UserString string
	Params     []interface{}
}

// Represents a response from a redis database
type Response struct {
	r   interface{}
	err error
}

// RedisDatabase holds the address of the redisDB
type RedisDatabase struct {
	Addr         string
	Port         string
	DbRequests   chan *Query
	BatchSize    int
	PollRate     time.Duration
	BatchResults chan Response
}

func (u RedisDatabase) getConn() redis.Conn {
	c, err := redis.Dial(u.Addr, u.Port)
	if err != nil {
		panic(err)
	}
	return c
}

// GetUserInfo returns all of a users information in the database
func (u RedisDatabase) GetUserInfo(user string) (info string, err error) {
	c := u.getConn()
	c.Send("MULTI")
	c.Send("GET", user+":Balance")
	c.Send("GET", user+":Stocks")
	c.Send("GET", user+":SellOrders")
	c.Send("GET", user+":BuyOrders")
	c.Send("GET", user+":SellTriggers")
	c.Send("GET", user+":BuyTriggers")
	c.Send("GET", user+":BalanceReserve")
	c.Send("GET", user+":StocksReserve")
	c.Send("GET", user+":History")
	r, err := c.Do("EXEC")
	if err != nil {
		return "", err
	}

	c.Close()
	return fmt.Sprintf("%v", r), err
}

// PushSell adds a record of the users requested sell to their account
func (u RedisDatabase) PushSell(user string, stock string, cost decimal.Decimal, shares decimal.Decimal) error {
	return u.pushOrder("Sell", user, stock, cost, shares)
}

// PopSell removes a users most recent requested sell
func (u RedisDatabase) PopSell(user string) (stock string, cost decimal.Decimal, shares decimal.Decimal, err error) {
	return u.popOrder("Sell", user)
}

// PushBuy adds a record of the users requested buy to their account
func (u RedisDatabase) PushBuy(user string, stock string, cost decimal.Decimal, shares decimal.Decimal) error {
	// Expires in 60s
	return u.pushOrder("Buy", user, stock, cost, shares)
}

// PopBuy removes a users most recent requested buy
func (u RedisDatabase) PopBuy(user string) (stock string, cost decimal.Decimal, shares decimal.Decimal, err error) {
	return u.popOrder("Buy", user)
}

func (u RedisDatabase) pushOrder(transType string, user string,
	stock string, cost decimal.Decimal, shares decimal.Decimal) error {
	accountSuffix := ""
	if transType == "Buy" {
		accountSuffix = ":BuyOrders"
	} else if transType == "Sell" {
		accountSuffix = ":SellOrders"
	} else {
		return errors.New("Bad transaction type of " + transType)
	}

	query := new(Query)
	query.Command = "RPUSH"
	query.UserString = user + accountSuffix
	query.Params = append(query.Params, u.encodeOrder(stock, cost, shares))

	u.DbRequests <- query

	resp := <-u.BatchResults

	_, err := redis.Int64(resp.r, resp.err)
	return err
}

func (u RedisDatabase) popOrder(transType string, user string) (stock string, cost decimal.Decimal, shares decimal.Decimal, err error) {
	accountSuffix := ""
	if transType == "Buy" {
		accountSuffix = ":BuyOrders"
	} else if transType == "Sell" {
		accountSuffix = ":SellOrders"
	} else {
		return stock, cost, shares, errors.New("Bad transaction type of " + transType)
	}
	query := new(Query)
	query.Command = "RPOP"
	query.UserString = user + accountSuffix

	u.DbRequests <- query

	resp := <-u.BatchResults

	recv, err := redis.String(resp.r, resp.err)

	stock, cost, shares = u.decodeOrder(recv)
	return stock, cost, shares, err
}

// Encodes a buy or sell order into a string, to be pushed onto the pending orders stack
// Returns a string following the format of:
//		"stock:cost:shares"
func (u RedisDatabase) encodeOrder(stock string, cost decimal.Decimal, shares decimal.Decimal) string {
	return stock + ":" + cost.String() + ":" + shares.String()
}

// Performs the opposite of encodeOrder
func (u RedisDatabase) decodeOrder(order string) (stock string, cost decimal.Decimal, shares decimal.Decimal) {
	split := strings.Split(order, ":")
	if len(split) == 3 {
		stock = split[0]
		cost, _ = decimal.NewFromString(split[1])
		shares, _ = decimal.NewFromString(split[2])
	} else {
		stock = ""
		cost, _ = decimal.NewFromString("0")
		shares, _ = decimal.NewFromString("0")
	}

	return stock, cost, shares
}

// AddFunds adds amount dollars to the user account
func (u RedisDatabase) AddFunds(user string, amount decimal.Decimal) error {
	_, err := u.fundAction("Add", user, ":Balance", amount)
	return err
}

// GetFunds returns the amount of available funds in a users account
func (u RedisDatabase) GetFunds(user string) (decimal.Decimal, error) {
	amount := decimal.NewFromFloat(0.0)
	return u.fundAction("Get", user, ":Balance", amount)
}

// RemoveFunds remove n funds from the user's account
// amount is the absolute value of the funds being removed
func (u RedisDatabase) RemoveFunds(user string, amount decimal.Decimal) error {
	_, err := u.fundAction("Remove", user, ":Balance", amount)
	return err
}

// AddReserveFunds adds funds to a user's reserve account
func (u RedisDatabase) AddReserveFunds(user string, amount decimal.Decimal) error {
	_, err := u.fundAction("Add", user, ":BalanceReserve", amount)
	return err
}

// GetReserveFunds returns the amount of funds present in a users reserve account
func (u RedisDatabase) GetReserveFunds(user string) (decimal.Decimal, error) {
	amount := decimal.NewFromFloat(0.0)
	return u.fundAction("Get", user, ":BalanceReserve", amount)
}

// RemoveReserveFunds removes n funds from a users account
// Pass in the absoloute value of funds to be removed.
func (u RedisDatabase) RemoveReserveFunds(user string, amount decimal.Decimal) error {
	_, err := u.fundAction("Add", user, ":BalanceReserve", amount)
	return err
}

// stockAction handles the generic stock commands
func (u RedisDatabase) fundAction(action string, user string,
	accountSuffix string, amount decimal.Decimal) (decimal.Decimal, error) {
	command := ""
	if action == "Add" {
		command = "INCRBYFLOAT"
	} else if action == "Get" {
		command = "GET"
	} else if action == "Remove" {
		command = "INCRBYFLOAT"
		amount = amount.Neg()
	} else {
		return decimal.NewFromFloat(0.0), errors.New("Bad action attempt on funds")
	}

	query := new(Query)
	query.Command = command
	query.UserString = user + accountSuffix
	if action != "Get" {
		query.Params = append(query.Params, amount)
	}

	u.DbRequests <- query
	resp := <-u.BatchResults

	r, err := redis.Float64(resp.r, resp.err)

	return decimal.NewFromFloat(r), err
}

// GetStock returns the users available balance of said stock
func (u RedisDatabase) GetStock(user string, stock string) (decimal.Decimal, error) {
	return u.stockAction("Get", user, ":Stocks", stock, decimal.NewFromFloat(0.0))
}

// RemoveStock removes int stocks from the users account
// Send the absolute value of the stock being removed
func (u RedisDatabase) RemoveStock(user string, stock string, amount decimal.Decimal) error {
	_, err := u.stockAction("Remove", user, ":Stocks", stock, amount)
	return err
}

// AddStock adds shares to the user account
func (u RedisDatabase) AddStock(user string, stock string, shares decimal.Decimal) error {
	_, err := u.stockAction("Add", user, ":Stocks", stock, shares)
	return err
}

// AddReserveStock adds n shares of stock to a user's account
func (u RedisDatabase) AddReserveStock(user string, stock string, amount decimal.Decimal) error {
	_, err := u.stockAction("Add", user, ":StocksReserve", stock, amount)
	return err
}

// GetReserveStock returns the amount of shares present in a user's reserve account
func (u RedisDatabase) GetReserveStock(user string, stock string) (decimal.Decimal, error) {
	return u.stockAction("Get", user, ":StocksReserve", stock, decimal.NewFromFloat(0.0))
}

// RemoveReserveStock removes n shares of stock from a user's reserve account
func (u RedisDatabase) RemoveReserveStock(user string, stock string, amount decimal.Decimal) error {
	_, err := u.stockAction("Remove", user, ":StocksReserve", stock, amount)
	return err
}

// stockAction handles the generic stock commands
func (u RedisDatabase) stockAction(action string, user string,
	accountSuffix string, stock string, amount decimal.Decimal) (decimal.Decimal, error) {
	command := ""
	if action == "Add" {
		command = "HINCRBYFLOAT"
	} else if action == "Get" {
		command = "HGET"
	} else if action == "Remove" {
		command = "HINCRBYFLOAT"
		amount = amount.Neg()
	} else {
		return decimal.NewFromFloat(0.0), errors.New("Bad action attempt on stocks")
	}

	query := new(Query)
	query.Command = command
	query.UserString = user + accountSuffix
	query.Params = append(query.Params, stock)
	if action != "Get" {
		query.Params = append(query.Params, amount)
	}

	u.DbRequests <- query

	var r float64
	var err error
	resp := <-u.BatchResults

	r, err = redis.Float64(resp.r, resp.err)
	if err != nil {
		return decimal.Decimal{}, err
	}

	rDec := decimal.NewFromFloat(r)
	return rDec, nil
}

// DeleteKey deletes a key in the database
// use this function with caution...
func (u RedisDatabase) DeleteKey(key string) {
	conn := u.getConn()
	conn.Do("DEL", key)
	conn.Close()
}

func (u RedisDatabase) DbRequestWorker() {
	reqQue := []*Query{}
	for {
		// Block until request received
		select {
		case request := <-u.DbRequests:
			reqQue = append(reqQue, request)
			if len(reqQue) >= u.BatchSize {
				u.MakeDbRequests(reqQue)
				reqQue = nil
				reqQue = []*Query{}
			}
		case <-time.After(u.PollRate * time.Millisecond):
			u.MakeDbRequests(reqQue)
			reqQue = nil
			reqQue = []*Query{}
		}
	}
}

func (u RedisDatabase) MakeDbRequests(requestQue []*Query) {
	// Batch size has been reached or poll time has passed,
	conn := u.getConn()
	for _, query := range requestQue {
		if len(query.Params) == 0 {
			conn.Send(query.Command, query.UserString)
		} else if len(query.Params) == 1 {
			conn.Send(query.Command, query.UserString, query.Params[0])
		} else if len(query.Params) == 2 {
			conn.Send(query.Command, query.UserString, query.Params[0], query.Params[1])
		} else {
			// Should never happen...
			panic("More params then 3!")
		}
	}
	conn.Flush()

	for i := 0; i < len(requestQue); i++ {
		r, err := conn.Receive()
		// Recieve results from queries
		resp := Response{r, err}
		// Notify waiting processes of batch execution
		u.BatchResults <- resp
	}
	conn.Close()
}
