package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type outgoingRequest struct {
	endpoint string
	params   url.Values
}

type endpointHit struct {
	duration time.Duration
	when     time.Time
}

var transcount uint64
var endpointTimes map[string][]endpointHit
var endpointMutex sync.Mutex

// go run WorkloadGen.go serverAddr:port workloadfile
func main() {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: server address, workloadfile, delay(ms), getlog(bool)")
		return
	}

	serverAddr := os.Args[1]
	workloadFile := os.Args[2]
	delayMs, _ := strconv.Atoi(os.Args[3])
	getLogStr := os.Args[4]
	var getLog bool
	if getLogStr != "true" {
		getLog = false
	} else {
		getLog = true
	}
	endpointTimes = make(map[string][]endpointHit)

	fmt.Printf("Testing %v on serverAddr %v with delay of %vms\n", workloadFile, serverAddr, delayMs)

	users := splitUsersFromFile(workloadFile)
	fmt.Printf("Found %d users...\n", len(users))
	go countTPS()

	runRequests(serverAddr, users, delayMs, getLog)
	fmt.Printf("Done!\n")

	printEndpointStats()
	saveEndpointStats()
}

func runRequests(serverAddr string, users map[string][]outgoingRequest, delay int, getLog bool) {
	var wg sync.WaitGroup
	for userName, commands := range users {
		fmt.Printf("Running user %v's commands...\n", userName)

		wg.Add(1)
		go runUserRequests(serverAddr, delay, userName, commands, &wg)
	}

	// Wait for commands, then manually post the final dumplog
	wg.Wait()
	if getLog {
		resp, httpErr := http.PostForm("http://"+serverAddr+"/DUMPLOG/", url.Values{"filename": {"./output.xml"}})
		if httpErr != nil {
			panic(httpErr)
		}

		data, decodeErr := ioutil.ReadAll(resp.Body)
		if decodeErr != nil {
			panic(decodeErr)
		}

		file, createErr := os.Create("./output.xml")
		defer file.Close()
		if createErr != nil {
			fmt.Printf("error: %v %v\n", createErr, file)
		}
		_, writeErr := file.Write(data)
		if writeErr != nil {
			panic(writeErr)
		}
	}
}

func runUserRequests(serverAddr string, delay int, userName string, commands []outgoingRequest, wg *sync.WaitGroup) {
	defer wg.Done()

	//timeout := time.Duration(3 * time.Second)
	client := http.Client{
	//Timeout: timeout,
	}

	// Issue login before executing any commands
	resp, err := client.PostForm("http://"+serverAddr+"/"+"LOGIN"+"/", url.Values{"username": {userName}})
	if err != nil {
		fmt.Println(err)
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	for _, command := range commands {
		time.Sleep(time.Duration(rand.Intn(delay)) * time.Millisecond)

		var resp *http.Response
		var err error
		time0 := time.Now()

		for {
			resp, err = client.PostForm("http://"+serverAddr+"/"+command.endpoint+"/", command.params)
			defer resp.Body.Close()
			if err != nil {
				fmt.Println(err.Error())
			} else {
				break
			}
		}

		responseTime := time.Since(time0)

		endpointMutex.Lock()
		hitEvent := endpointHit{
			responseTime,
			time0,
		}
		endpointTimes[command.endpoint] = append(endpointTimes[command.endpoint], hitEvent)
		endpointMutex.Unlock()
		atomic.AddUint64(&transcount, 1)
	}
}

func splitUsersFromFile(filename string) map[string][]outgoingRequest {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	// https://regex101.com/r/O6xaTp/3
	re := regexp.MustCompile(`\[\d+\] ((?P<endpoint>\w+),(?P<user>\w+)(,-*\w*\.*\d*)*)`)
	outputCommands := make(map[string][]outgoingRequest)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)

		if matches != nil {
			commandString := matches[1]
			parsedCommand := parseCommand(commandString)
			//endpoint := matches[2]
			user := matches[3]
			outputCommands[user] = append(outputCommands[user], parsedCommand)
		} else {
			fmt.Println("Error parsing command: ", line)
		}
	}

	return outputCommands
}

// Parse a single line command into the corresponding endpoint and values
func parseCommand(cmd string) outgoingRequest {
	subcmd := strings.Split(cmd, ",")
	endpoint := subcmd[0]
	var v url.Values

	// username, stock, amount, filename
	switch endpoint {
	case "ADD":
		v = url.Values{
			"username": {subcmd[1]},
			"amount":   {subcmd[2]},
		}
	case "QUOTE", "CANCEL_SET_BUY", "CANCEL_SET_SELL":
		v = url.Values{
			"username": {subcmd[1]},
			"stock":    {subcmd[2]},
		}
	case "SELL", "BUY", "SET_BUY_AMOUNT", "SET_BUY_TRIGGER", "SET_SELL_AMOUNT", "SET_SELL_TRIGGER":
		v = url.Values{
			"username": {subcmd[1]},
			"stock":    {subcmd[2]},
			"amount":   {subcmd[3]},
		}
	case "COMMIT_BUY", "CANCEL_BUY", "COMMIT_SELL", "CANCEL_SELL", "DISPLAY_SUMMARY":
		v = url.Values{
			"username": {subcmd[1]},
		}
	}

	out := outgoingRequest{
		endpoint,
		v,
	}
	return out
}

func countTPS() {
	var tpsStart uint64
	var tpsEnd uint64
	elapsedtime := 0
	for {
		tpsStart = transcount
		time.Sleep(time.Second)
		tpsEnd = transcount

		fmt.Printf("%d Running at %d TPS, %d trans\n", elapsedtime, tpsEnd-tpsStart, transcount)
		elapsedtime++
	}
}

func printEndpointStats() {
	for endpoint := range endpointTimes {
		responseTimes := endpointTimes[endpoint]
		totalTime, _ := time.ParseDuration("0")
		numCommands := float64(len(responseTimes))

		for _, event := range responseTimes {
			totalTime += event.duration
		}

		// Some gross time type conversions
		avgTimeFloat := totalTime.Seconds() / numCommands
		avgTimeString := strconv.FormatFloat(avgTimeFloat, 'f', 6, 64)
		avgTimeTime, _ := time.ParseDuration(avgTimeString + "s")

		fmt.Printf("%v: %v\n", endpoint, avgTimeTime)
	}
}

func saveEndpointStats() {
	f, err := os.Create("./endpointStats.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.Write([]byte("ENDPOINT,when,duration\n"))
	for endpoint := range endpointTimes {
		hits := endpointTimes[endpoint]

		for _, hit := range hits {
			outline := fmt.Sprintf("%v,%v,%v\n", endpoint, hit.when.UnixNano(), hit.duration.Nanoseconds())
			f.Write([]byte(outline))
		}
	}
}
