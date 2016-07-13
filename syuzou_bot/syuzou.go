package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Data struct {
	Message     string `json:"message"`
	CreatedTime string `json:"created_time"`
	FullPicture string `json:"full_picture"`
	Id          string `json:"id"`
}
type Paging struct {
	Previous string `json:"previous"`
	Next     string `json:"next"`
}
type Posts struct {
	Datas  []Data `json:"data"`
	Paging `json:"paging"`
}

type Response struct {
	Posts `json:"posts"`
	Id    string `json:"id"`
}

type Slack struct {
	Text      string `json:"text"`
	Username  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
	Channel   string `json:"channel"`
}

const (
	APP_ID       = "shuzou.0209"
	APP_TOKEN    = ""
	INCOMING_URL = "https://hooks.slack.com/services/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"
)

func main() {
	request_url := fmt.Sprintf("https://graph.facebook.com/v2.5/%s?fields=posts.limit(5){message,created_time,full_picture}&access_token=%s", APP_ID, APP_TOKEN)
	res, _ := http.Get(request_url)
	body, _ := ioutil.ReadAll(res.Body)

	var response Response

	err := json.Unmarshal([]byte(body), &response)
	if err != nil {
		fmt.Println("error:", err)
	}
	now := time.Now()
	for _, v := range response.Posts.Datas {
		t, err := time.Parse(
			"2006-01-02T15:04:05-0700",
			v.CreatedTime,
		)
		if err != nil {
			fmt.Println("time error:", err)
			continue
		}
		t = t.Add(9 * time.Hour)
		if int(t.Month()) == int(now.Month()) && t.Day() == now.Day() {
			if strings.Index(v.Message, "本日のしゅぞう定食は") != -1 {
				postSlack(t, v)
			} else {
				continue
			}
		} else {
			continue
		}
	}
}

func postSlack(today time.Time, data Data) {
	msg := fmt.Sprintf("%s\n\n%s", data.Message, data.FullPicture)
	params, _ := json.Marshal(Slack{
		msg,
		"しゅぞうBot",
		":shuzou:",
		"#lunch"})
	resp, _ := http.PostForm(
		INCOMING_URL,
		url.Values{"payload": {string(params)}},
	)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Println(string(body))
	//fmt.Println(data.Message, data.FullPicture)
}
