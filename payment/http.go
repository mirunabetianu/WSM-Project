package main

import (
	"encoding/json"
	"net/http"
)

const BASE_URL = "http://localhost:"

func call(port string, endpoint string, target interface{}) error {
	url := BASE_URL + port + endpoint
	resp, err := http.Get(url)
	if err != nil {
		panic("error")
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
