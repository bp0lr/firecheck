//
// @bp0lr - 11/11/2020
//

package main

import (
	"os"
	"fmt"
	"net"
	"sync"
	"time"
	"bytes"
	"bufio"
	"strings"
	"net/url"
	"net/http"
	"math/rand"
	"crypto/tls"
	
	"encoding/json"

	flag "github.com/spf13/pflag"
)

var (
	workersArg	      int
	headerArg         []string
	urlArg            string
	userArg			  string
	proxyArg          string
	outputFileArg     string
	verboseArg        bool
	useRandomAgentArg bool
	notFancyArg       bool
)

type bbTest struct {
	Txt    string `json:"txt"`
	User   string  `json:"user"`
}

func main() {

	flag.StringArrayVarP(&headerArg, "header", "H", nil, "Add custom Headers to the request")
	flag.IntVarP(&workersArg, "workers", "w", 50, "Workers amount")
	flag.StringVarP(&urlArg, "url", "u", "", "The firebase url to test")
	flag.StringVarP(&userArg, "user", "m", "", "Add your username for write POC")
	flag.BoolVarP(&verboseArg, "verbose", "v", false, "Display extra info about what is going on")
	flag.StringVarP(&proxyArg, "proxy", "p", "", "Add a HTTP proxy")
	flag.BoolVarP(&useRandomAgentArg, "random-agent", "r", false, "Set a random User Agent")
	flag.StringVarP(&outputFileArg, "output", "o", "", "Output file to save the results to")
	flag.BoolVarP(&notFancyArg, "simple", "s", false, "Display only the url without R W D")
	
	flag.Parse()

	//concurrency
	workers := 50
	if workersArg > 0  && workersArg < 100 {
		workers = workersArg
	}

	client := newClient(proxyArg)

	jobs := make(chan string)
	var wg sync.WaitGroup

	var outputFile *os.File
	var err0 error
	if outputFileArg != "" {
		outputFile, err0 = os.OpenFile(outputFileArg, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err0 != nil {
			fmt.Printf("cannot write %s: %s", outputFileArg, err0.Error())
			return
		}
		
		defer outputFile.Close()
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			for raw := range jobs {
				
				u, err := url.ParseRequestURI(raw)
				if err != nil {
					if verboseArg {
						fmt.Printf("[-] Invalid url: %s\n", raw)
					}
					continue
				}				
				
				processRequest(&wg, u, client, outputFile)

			}
			wg.Done()			
		}()
	}

	if len(urlArg) < 1 {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			jobs <- sc.Text()
		}
	} else {
		jobs <- urlArg
	}

	close(jobs)	
	wg.Wait()	
}

func processRequest(wg *sync.WaitGroup, u *url.URL, client *http.Client, outputFile *os.File) {

	if verboseArg {
		fmt.Printf("[+] Testing: %v\n", u.String())
	}
	
	//check read
	/////////////////////////////////////////////////////////////////////////
	read, resp, err :=check("R", u, client)
	write, _, _ :=check("W", u, client)
	delete, _, _ :=check("D", u, client)
	
	if(read || write || delete){
		if(notFancyArg){
			fmt.Printf("%v\n", u.String())
		} else {
			str:= "[+] " + u.String() + " => "
			if(read){
				str+= " R "
			}
			if(write){
				str+= " W "
			}
			if(delete){
				str+= " D "
			}

			if outputFileArg != "" {
				outputFile.WriteString(u.String() + "\n")
			}

			fmt.Println(str)
		}
	} else {
		if(err != nil && verboseArg){
			fmt.Printf("[-] Error: %v [%v]\n", err, resp)
		} else {
			if(verboseArg){
				fmt.Printf("[-] %v [%v]\n", u.String(), resp)
			}
		}
	}
}

func newClient(proxy string) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:    30,
		IdleConnTimeout: time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout: time.Second * 10,
		}).DialContext,
	}

	if proxy != "" {
		if p, err := url.Parse(proxy); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}
	
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
	}
	
	return client
}

func getUserAgent() string {
	payload := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 8_4_1 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12H321 Safari/600.1.4",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (iPad; CPU OS 7_1_2 like Mac OS X) AppleWebKit/537.51.2 (KHTML, like Gecko) Version/7.0 Mobile/11D257 Safari/9537.53",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(payload))

	pick := payload[randomIndex]

	return pick
}

//CheckRead desc
func check(checkType string, u *url.URL, client *http.Client) (bool, int, error){

	var url string
	var req *http.Request
	var err error

	// Read
	if(checkType == "R"){
		url = u.Scheme + "://" + u.Host + "/.json"
		req, err = http.NewRequest("GET", url, nil)
	}

	// Write
	if(checkType == "W"){
		url = u.Scheme + "://" + u.Host + "/BountyTest.json"

		bb := bbTest{Txt:  "Bounty test" }
		if(len(userArg) > 0){
			bb.User = userArg
		} else {
			bb.User = "firecheck"
		}

		j, _:=json.Marshal(bb)

		req, err = http.NewRequest("PUT", url, bytes.NewBuffer(j))
	}

	// Delete
	if(checkType == "D"){
		url = u.Scheme + "://" + u.Host + "/BountyTest.json"
		req, err = http.NewRequest("DELETE", url, nil)
	}

	if err != nil {
		if verboseArg {
			fmt.Printf("[-] Error: %v\n", err)
		}
		return false, 0, err
	}

	if useRandomAgentArg {
		req.Header.Set("User-Agent", getUserAgent())
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; firecheck/1.0)")
	}

	// add headers to the request
	for _, h := range headerArg {
		parts := strings.SplitN(h, ":", 2)

		if len(parts) != 2 {
			continue
		}
		req.Header.Set(parts[0], parts[1])
	}

	// send the request
	resp, err := client.Do(req)
	
	if err != nil {
		if verboseArg {
			fmt.Printf("[-] Error: %v\n", err)
		}
		return false, 0, err
	}	
	
	defer resp.Body.Close()

	if(resp.StatusCode == 200){
		return true, 200, nil
	}
	
	return false, resp.StatusCode, nil
}