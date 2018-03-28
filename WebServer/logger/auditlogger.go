package logger

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Logger interface {
	UserCommand(server string, transNum int,
		command string, username interface{}, stock interface{},
		filename interface{}, funds interface{})

	QuoteServer(server string, transNum int,
		price string, stock string, user string, qsTime uint64, key string)

	AccountTransaction(server string, transNum int,
		action string, user interface{}, funds interface{})

	SystemError(server string, transNum int,
		command string, user interface{}, stock interface{},
		filename interface{},
		funds interface{}, errorMsg interface{})

	SystemEvent(server string, transNum int,
		command string, username interface{}, stock interface{},
		filename interface{}, funds interface{})

	DumpLog(filename string, username interface{})
}

type AuditLogger struct {
	Addr string
}

func (al AuditLogger) DumpLog(filename string, username interface{}) {
	params := map[string]string{
		"filename": filename,
	}
	if username != nil {
		params["username"] = username.(string)
	}
	al.SendLog("/dumpLog", params)
}

func (al AuditLogger) UserCommand(server string, transNum int, command string,
	username interface{}, stock interface{}, filename interface{}, funds interface{}) {
	params := map[string]string{
		"server":         server,
		"transactionNum": strconv.Itoa(transNum),
		"command":        command,
	}
	if username != nil {
		params["username"] = username.(string)
	}
	if stock != nil {
		params["stockSymbol"] = stock.(string)
	}
	if filename != nil {
		params["filename"] = filename.(string)
	}
	if funds != nil {
		params["funds"] = funds.(string)
	}
	al.SendLog("/userCommand", params)
}

func (al AuditLogger) SystemEvent(server string, transNum int, command string, username interface{}, stock interface{},
	filename interface{}, funds interface{}) {
	params := map[string]string{
		"server":         server,
		"transactionNum": strconv.Itoa(transNum),
		"command":        command,
	}
	if username != nil {
		params["username"] = username.(string)
	}
	if stock != nil {
		params["stockSymbol"] = stock.(string)
	}
	if filename != nil {
		params["filename"] = filename.(string)
	}
	if funds != nil {
		params["funds"] = funds.(string)
	}
	al.SendLog("/systemEvent", params)
}

func (al AuditLogger) SystemError(server string, transNum int, command string, user interface{}, stock interface{}, filename interface{},
	funds interface{}, errorMsg interface{}) {
	return
	params := map[string]string{
		"server":         server,
		"transactionNum": strconv.Itoa(transNum),
		"command":        command,
	}
	if user != nil {
		params["username"] = user.(string)
	}
	if stock != nil {
		params["stockSymbol"] = stock.(string)
	}
	if filename != nil {
		params["filename"] = filename.(string)
	}
	if funds != nil {
		params["funds"] = funds.(string)
	}
	if errorMsg != nil {
		params["errorMessage"] = errorMsg.(string)
	}
	al.SendLog("/errorEvent", params)
}

func (al AuditLogger) AccountTransaction(server string, transactionNum int, action string, user interface{}, funds interface{}) {
	params := map[string]string{
		"server":         server,
		"transactionNum": strconv.Itoa(transactionNum),
		"action":         action,
	}
	if user != nil {
		params["username"] = user.(string)
	}
	if funds != nil {
		params["funds"] = funds.(string)
	}
	al.SendLog("/accountTransaction", params)
}

func (al AuditLogger) QuoteServer(server string, transactionNum int,
	price string, stock string, user string, qsTime uint64, key string) {
	params := map[string]string{
		"server":          server,
		"transactionNum":  strconv.Itoa(transactionNum),
		"price":           price,
		"stockSymbol":     stock,
		"username":        user,
		"quoteServerTime": strconv.FormatUint(qsTime, 10),
		"cryptokey":       key,
	}
	al.SendLog("/quoteServer", params)
}

func (al AuditLogger) SendLog(slash string, params map[string]string) {
	req, err := http.NewRequest("get", al.Addr+slash, nil)
	req.Header.Set("Connection", "close")
	if err != nil {
		log.Print(err)
	}

	url := req.URL.Query()
	for k, v := range params {
		url.Add(k, v)
	}

	req.URL.RawQuery = url.Encode()
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{Transport: transport}
	var resp *http.Response
	for {
		resp, err = client.Do(req)

		if err != nil { // trans server down? retry
			fmt.Println(err.Error())
		} else {
			break
		}
	}
	resp.Body.Close()
}
