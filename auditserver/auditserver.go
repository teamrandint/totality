package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"seng468/auditserver/commands"
	"seng468/auditserver/log"
	"sync"
	"time"
	// _ "net/http/pprof"
)

func auditWorker() {
	for {
		// receive from channel, or be blocked
		command := <-logChannel
		eventlog.Insert(command)
	}
}

func userCommandHandler(w http.ResponseWriter, r *http.Request) {
	timestamp := makeTimestamp()
	query := r.URL.Query()
	fmt.Printf("Received userCommand at %v\n", timestamp)

	v := &commands.UserCommand{
		Timestamp:      timestamp,
		Server:         query.Get("server"),
		TransactionNum: query.Get("transactionNum"),
		Command:        query.Get("command"),
		Username:       query.Get("username"),
		StockSymbol:    query.Get("stockSymbol"),
		Filename:       query.Get("filename"),
		Funds:          query.Get("funds"),
	}
	logChannel <- v

	w.Write([]byte("OK"))
}

func quoteServerHandler(w http.ResponseWriter, r *http.Request) {
	timestamp := makeTimestamp()
	query := r.URL.Query()
	fmt.Printf("Received quoteServer at %v\n", timestamp)

	v := &commands.QuoteServer{
		Timestamp:       timestamp,
		Server:          query.Get("server"),
		TransactionNum:  query.Get("transactionNum"),
		Username:        query.Get("username"),
		StockSymbol:     query.Get("stockSymbol"),
		Price:           query.Get("price"),
		QuoteServerTime: query.Get("quoteServerTime"),
		Cryptokey:       query.Get("cryptokey"),
	}
	logChannel <- v

	w.Write([]byte("OK"))
}

func accountTransactionHandler(w http.ResponseWriter, r *http.Request) {
	timestamp := makeTimestamp()
	query := r.URL.Query()
	fmt.Printf("Received accountTransaction at %v\n", timestamp)

	v := &commands.AccountTransaction{
		Timestamp:      timestamp,
		Server:         query.Get("server"),
		TransactionNum: query.Get("transactionNum"),
		Action:         query.Get("action"),
		Username:       query.Get("username"),
		Funds:          query.Get("funds"),
	}
	logChannel <- v

	w.Write([]byte("OK"))
}

func systemEventHandler(w http.ResponseWriter, r *http.Request) {
	timestamp := makeTimestamp()
	query := r.URL.Query()
	fmt.Printf("Received systemEvent at %v\n", timestamp)

	v := &commands.SystemEvent{
		Timestamp:      timestamp,
		Server:         query.Get("server"),
		TransactionNum: query.Get("transactionNum"),
		Command:        query.Get("command"),
		Username:       query.Get("username"),
		StockSymbol:    query.Get("stockSymbol"),
		Filename:       query.Get("filename"),
		Funds:          query.Get("funds"),
	}
	logChannel <- v

	w.Write([]byte("OK"))
}

func errorEventHandler(w http.ResponseWriter, r *http.Request) {
	timestamp := makeTimestamp()
	query := r.URL.Query()
	fmt.Printf("Received errorEvent at %v\n", timestamp)

	v := &commands.ErrorEvent{
		Timestamp:      timestamp,
		Server:         query.Get("server"),
		TransactionNum: query.Get("transactionNum"),
		Command:        query.Get("command"),
		Username:       query.Get("username"),
		StockSymbol:    query.Get("stockSymbol"),
		Filename:       query.Get("filename"),
		Funds:          query.Get("funds"),
		ErrorMessage:   query.Get("errorMessage"),
	}
	logChannel <- v

	w.Write([]byte("OK"))
}

func dumpLogHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	dumpfile := query.Get("filename")
	userLog := query.Get("username")
	dumpfileB := string(bytes.Trim([]byte(dumpfile), "\x00"))
	//if dumpfileB != "./test.log" {
	//	panic(fmt.Sprintf("Names not equal %q ./test.log\n len=%v", dumpfile, len(dumpfile)))
	//}

	file, err := os.Create(string(dumpfileB))
	if err != nil {
		fmt.Printf("error: %v %v\n", err, file)
	}
	fmt.Printf("Dumping log to %v, with user set as %v", dumpfileB, userLog)

	mutex.Lock()
	eventlog.Write(file)
	mutex.Unlock()

	file.Close()
}

func dumpLogRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("filename")
	http.ServeFile(w, r, filename)
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

var eventlog = log.Log{
	Entries: make([]commands.Command, 5000000),
}
var mutex sync.Mutex
var logChannel = make(chan commands.Command, 10000)

func main() {
	http.HandleFunc("/userCommand", userCommandHandler)
	http.HandleFunc("/quoteServer", quoteServerHandler)
	http.HandleFunc("/accountTransaction", accountTransactionHandler)
	http.HandleFunc("/systemEvent", systemEventHandler)
	http.HandleFunc("/errorEvent", errorEventHandler)
	http.HandleFunc("/dumpLog", dumpLogHandler)
	http.HandleFunc("/dumpLogRetrieve", dumpLogRetrieveHandler)

	fmt.Printf("Audit server listening on %s:%s\n", os.Getenv("auditaddr"), os.Getenv("auditport"))
	go auditWorker()
	if err := http.ListenAndServe(":"+os.Getenv("auditport"), nil); err != nil {
		panic(err)
	}
}
