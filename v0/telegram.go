package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var mainURL = "https://api.telegram.org/bot%s"

type TelegramBot struct {
	BotToken string
	ChatID   string
}

type TelegramUpdate struct {
	LastUpdateID int64
	Result       []TelegramUpdateItem `json:"result"`
}

type TelegramUpdateItem struct {
	UpdateID    int64                     `json:"update_id"`
	ChannelPost TelegramUpdateChannelPost `json:"channel_post"`
}

type TelegramUpdateChannelPost struct {
	Text string `json:"text"`
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
		log.Println(err)
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (t *TelegramBot) GetUpdate() (TelegramUpdate, error) {
	var (
		err error
		b   TelegramUpdate
	)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, t.getURL("getUpdates"), nil)
	if err != nil {
		log.Println(err)
		return b, err
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return b, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return b, err
	}

	if err = json.Unmarshal(body, &b); err != nil {
		log.Println(err)
		return b, err
	}

	return b, nil
}

func (tu *TelegramUpdate) GetLastUpdateID() error {
	var (
		pv  string
		err error
	)

	if pv, err = getKV("telegram_bot_last_update_id"); err != nil {
		return err
	}
	if tu.LastUpdateID, err = strconv.ParseInt(pv, 10, 64); err != nil {
		return err
	}

	return err
}

func (tu *TelegramUpdate) GetLastUpdateMessage(prefix string) string {
	var (
		m      string
		lastID int64
	)

	for _, rm := range tu.Result {
		if rm.UpdateID > tu.LastUpdateID {
			set := strings.Split(rm.ChannelPost.Text, " ")
			if len(set) > 0 && set[0] == prefix {
				m = rm.ChannelPost.Text
			}
			lastID = rm.UpdateID
		}
	}
	if lastID > 0 {
		if err := tu.SetLastUpdateID(lastID); err != nil {
			log.Println(err)
		}
	}

	log.Println(">>", m)
	return m
}

func (tu *TelegramUpdate) SetLastUpdateID(updateID int64) error {
	var (
		err error
	)

	tu.LastUpdateID = updateID
	if err = setKV("telegram_bot_last_update_id", fmt.Sprint(tu.LastUpdateID)); err != nil {
		return err
	}

	return err
}
