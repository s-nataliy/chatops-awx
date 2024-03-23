package controller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

func BuildListTemplates(url string, authAWX string) string {
	var replyMsg string
	templates := GetStructTemplates(url, authAWX)
	for _, template := range templates {
		replyMsg = replyMsg + "\n" + template.name
	}
	return replyMsg
}

func GetStructTemplates(url string, authAWX string) []TemplateData {
	job_templates := RequestAPI("GET", url+"/api/v2/job_templates/?format=json", nil, authAWX)

	//чтение шаблонов
	templates := make([]TemplateData, 0)
	var templateData TemplateData
	for _, item := range job_templates.(map[string]interface{})["results"].([]interface{}) {
		templateData.id = (item.(map[string]interface{})["id"]).(float64)
		templateData.url = (item.(map[string]interface{})["url"]).(string)
		templateData.name = (item.(map[string]interface{})["name"]).(string)
		templateData.description = (item.(map[string]interface{})["description"]).(string)
		templates = append(templates, templateData)
	}
	return templates
}

func RunTemplate(command []string, conf Conf, authAWX string) (string, string) {
	TemplateList := GetStructTemplates(conf.URL, authAWX)
	temp_name := command[1]
	server_name := command[2]
	jsonServer, _ := json.Marshal(map[string]string{
		"limit": server_name,
	})

	var newURL string

	for _, selectTemplates := range TemplateList {
		if selectTemplates.name == temp_name {
			newURL = conf.URL + selectTemplates.url + "launch/"
		}
	}

	response := RequestAPI("POST", newURL, jsonServer, authAWX)
	return temp_name, GetStatusJob(response, conf, authAWX)
}

func GetStatusJob(responseJob interface{}, conf Conf, authAWX string) string {

	urlJob := responseJob.(map[string]interface{})["url"].(string)

	duration := time.Duration(6) * time.Second
	time.Sleep(duration)
	getJob := RequestAPI("GET", conf.URL+urlJob, nil, authAWX)
	statusJob := getJob.(map[string]interface{})["status"].(string)

	return statusJob
}

func RunSilence(command []string, replyMess *tgbotapi.Message, conf Conf, authAWX string) string {
	TemplateList := GetStructTemplates(conf.URL, authAWX)
	temp_name := "SilenceAlert"

	var newURL string

	replyText = replyMess.Text
	replyUser = replyMess.From.FirstName
	//dataCenter = strings.ToLower(replyUser[:strings.IndexByte(replyUser, '-')])

	alertName = replyText[:strings.IndexByte(replyText, ':')-1]

	instanceValue = replyText[strings.Index(replyText, "Instance: ")+10:]
	instanceValue = instanceValue[:strings.IndexByte(instanceValue, '\n')]
	if len(command) < 3 {
		durationValue = "1h"
	} else {
		durationValue = command[2]
	}
	dataCenter = strings.ToLower(command[1])
	jsonSilence, _ := json.Marshal(map[string]interface{}{
		"extra_vars": map[string]string{
			"datacenter": dataCenter,
			"alertname":  alertName,
			"instance":   instanceValue,
			"author":     userName,
			"duration":   durationValue,
		},
	})
	for _, selectTemplates := range TemplateList {
		if selectTemplates.name == temp_name {
			newURL = conf.URL + selectTemplates.url + "launch/"
		}
	}

	response := RequestAPI("POST", newURL, jsonSilence, authAWX)
	return GetStatusJob(response, conf, authAWX)
}
