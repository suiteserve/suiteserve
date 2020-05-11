package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const maxConns = 5

var (
	baseUri = flag.String("baseuri", "https://localhost:8080",
		"The base URI to which all requests are made.")

	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxConns,
		},
	}
	connGrp = semaphore.NewWeighted(int64(maxConns))
)

func main() {
	flag.Parse()
	log.Printf("-baseuri=%s\n", *baseUri)

	start := nowTimeMillis()
	var wg sync.WaitGroup
	wg.Add(30)
	for i := 0; i < 30; i++ {
		go newSuite(&wg, "Suite "+strconv.Itoa(i), randUint(30))
	}
	wg.Wait()
	end := nowTimeMillis()
	log.Printf("Total run time: %.1fs", float64(end-start)*time.Millisecond.Seconds())
}

func nowTimeMillis() int64 {
	now := time.Now()
	// Doesn't use now.UnixNano() to avoid Y2K262.
	return now.Unix()*time.Second.Milliseconds() +
		time.Duration(now.Nanosecond()).Milliseconds()
}

func randUint(max int32) int {
	return int(rand.Int31n(max))
}

func postJson(url string, body interface{}, resBody interface{}) http.Header {
	b, err := json.Marshal(body)
	if err != nil {
		log.Fatalln(err)
	}
	res, err := httpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println(err)
		}
	}()
	resBodyB, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Fatalf("non 2xx status: %v <<EOF\n%vEOF\n", res.StatusCode, string(resBodyB))
	}
	if resBody != nil {
		if err := json.Unmarshal(resBodyB, resBody); err != nil {
			log.Fatalln(err)
		}
	}
	return res.Header
}
