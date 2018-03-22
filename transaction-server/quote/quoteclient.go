package quoteclient

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/shopspring/decimal"
)

func Query(user string, stock string, transNum int) (decimal.Decimal, error) {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
	req, err := http.NewRequest("GET", "http://"+os.Getenv("quoteaddr")+":"+os.Getenv("quoteport")+"/quote", nil)
	if err != nil {
		log.Print(err)
		panic(err)
	}
	q := req.URL.Query()
	q.Add("user", user)
	q.Add("stock", stock)
	q.Add("transNum", strconv.Itoa(transNum))
	req.URL.RawQuery = q.Encode()

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error connecting to the quote server: %s", err.Error())
		return decimal.Decimal{}, err
	}
	amount, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %s", err.Error())
		return decimal.Decimal{}, err
	}
	return decimal.NewFromString(string(amount))
}
