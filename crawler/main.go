package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing target URL")
		os.Exit(1)
	}

	target_url := os.Args[1]
	log.Printf("Begin crawling %s", target_url)

	resp, err := http.Get(target_url)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// This is a blocking call. In real world, we should chunk reading the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Printf("Response status: %s", resp.Status)
	log.Printf("Response headers: %s", resp.Header)
	log.Printf("Response body: %s", body)

}
