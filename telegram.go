package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var mainURL = "https://api.telegram.org/bot%s"

type TelegramBot struct {
	BotToken string
	ChatID   string
}

func NewTelegramBot(botToken, chatID string) TelegramBot {
	return TelegramBot{
		BotToken: botToken,
		ChatID:   chatID,
	}
}

func (t *TelegramBot) getURL(api string) string {
	return fmt.Sprintf(mainURL, t.BotToken) + "/" + api
}

func (t *TelegramBot) getChatID() string {
	return t.ChatID
}

func (t *TelegramBot) SendMessage(s string) error {
	var (
		err error
	)
	payload := strings.NewReader("chat_id=" + t.getChatID() + "&text=" + s)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, t.getURL("sendMessage"), payload)
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
