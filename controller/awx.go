package controller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

var client = &http.Client{}

func RequestAPI(method string, url string, body []byte, login string) interface{} {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Basic "+login)

	if method == "POST" {
		req.Header.Add("Content-Type", "application/json")
	}

	if err != nil {
		log.Println("Ошибка при выполнении запроса к API", err)

	}
	resp, _ := client.Do(req)

	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var resData interface{}
	if err := json.Unmarshal([]byte(data), &resData); err != nil {
		log.Println("Ошибка при чтении JSON:", err)
	}

	return resData
}
