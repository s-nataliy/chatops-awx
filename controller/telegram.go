package controller

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v3"
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
	Token         string   `yaml:"token_telegram"`
	URL           string   `yaml:"url_awx"`
	LoginAWX      string   `yaml:"login_awx"`
	PassAWX       string   `yaml:"password_awx"`
	TelegramUsers []string `yaml:"telegram_users"`
}

var TemplateList []TemplateData
var foundUser = false

func TelegramBot() {
	//Чтение конфигурации из файла config.yaml
	var conf Conf
	//yml, err := ioutil.ReadFile("config.yaml")
	yml, err := ioutil.ReadFile(os.Getenv("CONFIG_FILE_PATH"))
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

			switch update.Message.Text {
			case "/list_temp":
				mainBot.Send(tgbotapi.NewMessage(chatID, BuildListTemplates(conf.URL, authAWX)))
			case "/run_temp":
				mainBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, ""))
				responseJob := RunTemplate(conf, authAWX)
				mainBot.Send(tgbotapi.NewMessage(chatID, "Статус выполнения Job'а: "+responseJob))

			}

		}
	}
}
