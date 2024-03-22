package controller

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

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
	Token         string   `json:"token_telegram"`
	URL           string   `json:"url_awx"`
	LoginAWX      string   `json:"login_awx"`
	PassAWX       string   `json:"password_awx"`
	TelegramUsers []string `json:"telegram_users"`
}

var TemplateList []TemplateData
var foundUser = false

func TelegramBot() {
	//Чтение конфигурации из файла config.json
	var conf Conf
	jsn, err := ioutil.ReadFile(os.Getenv("CONFIG_FILE_PATH"))
	if err != nil {
		log.Println(err)
	}

	json.Unmarshal(jsn, &conf)
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

			foundUser = false
			for _, user := range conf.TelegramUsers {
				if user == userName {
					foundUser = true
				}
			}

			if foundUser {
				log.Printf("Authorized on account %s", mainBot.Self.UserName)
			} else {
				log.Printf("Пользователь %v не авторизован", userName)
				continue
			}

			log.Printf("[%s](%d) %s", userName, userID, textMess)

			command := strings.Split(textMess, " ")
			switch command[0] {
			case "/list_temp":
				mainBot.Send(tgbotapi.NewMessage(chatID, BuildListTemplates(conf.URL, authAWX)))
			case "/run_temp":
				if len(command) != 3 {
					mainBot.Send(tgbotapi.NewMessage(chatID, "Неверно введена команда. Шаблон: /run_temp template_name server_name"))
				} else {
					responseJob := RunTemplate(command, conf, authAWX)
					mainBot.Send(tgbotapi.NewMessage(chatID, "Статус выполнения Job'а: "+responseJob))
				}
			}

		}
	}
}
