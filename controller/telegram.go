package controller

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v3"
)

// var mainBot *tgbotapi.BotAPI
var textMess string
var chatID int64
var userID int
var userName string

type TemplateData struct {
	id          float64
	url         string
	name        string
	description string
}

// Структура конфигурационного файла в формате yaml
type Conf struct {
	Token         string   `yaml:"token_telegram"`
	URL           string   `yaml:"url_awx"`
	LoginAWX      string   `yaml:"login_awx"`
	PassAWX       string   `yaml:"password_awx"`
	TelegramUsers []string `yaml:"telegram_users"`
}

var foundUser = false

func TelegramBot() {
	//Чтение конфигурации из файла config.yaml
	var conf Conf
	yml, err := ioutil.ReadFile("config.yaml")
	//yml, err := ioutil.ReadFile("myconfig.yaml")
	if err != nil {
		log.Println(err)
	}

	yaml.Unmarshal(yml, &conf)
	authAWX := base64.StdEncoding.EncodeToString([]byte(conf.LoginAWX + ":" + conf.PassAWX))

	mainBot, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		log.Panic(err)
	}

	// Setup long-polling request
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := mainBot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // Есть новое сообщение
			textMess = update.Message.Text          // Текст сообщения
			chatID = update.Message.Chat.ID         //  ID чата
			userID = update.Message.From.ID         // ID пользователя
			userName = update.Message.From.UserName // Имя пользователя

			for _, user := range conf.TelegramUsers {
				if user == userName {
					foundUser = true
				}
			}

			if foundUser {
				log.Printf("Authorized on account %s", mainBot.Self.UserName)
				//continue
			} else {
				log.Printf("Пользователь %v не авторизован", userName)
				//break
			}

			log.Printf("[%s](%d) %s", userName, userID, textMess)

			url := conf.URL

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
			//парсинг команды
			command := strings.Split(textMess, " ")
			switch command[0] {
			case "/list_temp":
				replyMsg := ""
				for _, template := range templates {
					replyMsg = replyMsg + "\n" + template.name
				}
				mainBot.Send(tgbotapi.NewMessage(chatID, replyMsg))
			case "/run_temp":
				/*if len(command) != 3 {
					mainBot.Send(tgbotapi.NewMessage(chatID, "Неверно введена команда. Шаблон: /run_temp template_name server_name"))
				}*/
				temp_name := "ping"  //command[1]
				server_name := "all" //command[2]
				jsonServer, _ := json.Marshal(map[string]string{
					"limit": server_name,
				})

				var newURL string

				for _, selectTemplates := range templates {
					if selectTemplates.name == temp_name {
						newURL = url + selectTemplates.url + "launch/"
					}
				}

				postJob := RequestAPI("POST", newURL, jsonServer, authAWX)
				urlJob := postJob.(map[string]interface{})["url"].(string)

				duration := time.Duration(6) * time.Second
				time.Sleep(duration)
				getJob := RequestAPI("GET", url+urlJob, nil, authAWX)
				statusJob := getJob.(map[string]interface{})["status"].(string)

				mainBot.Send(tgbotapi.NewMessage(chatID, "Статус выполнения Job'а "+temp_name+": "+statusJob))
			}

		}
	}
}
