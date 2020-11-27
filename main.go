package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	EnvSlackWebhook  = "SLACK_WEBHOOK"
	EnvSlackIcon     = "SLACK_ICON"
	EnvSlackChannel  = "SLACK_CHANNEL"
	EnvSlackTitle    = "SLACK_TITLE"
	EnvSlackMessage  = "SLACK_MESSAGE"
	EnvSlackColor    = "SLACK_COLOR"
	EnvSlackUserName = "SLACK_USERNAME"
	EnvGithubActor   = "GITHUB_ACTOR"
	EnvChangeLogUrl  = "CHANGELOG_URL"
	EnvReleasesUrl   = "RELEASES_URL"
	EnvVariantsList  = "VARIANTS"
	EnvSlackPretext  = "SLACK_PRETEXT"
	EnvSlackFooter   = "SLACK_FOOTER"
)

type Webhook struct {
	Text        string       `json:"text,omitempty"`
	UserName    string       `json:"username,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links"`
	Attachments []Attachment `json:"attachments,omitmepty"`
}

type Attachment struct {
	Text       string   `json:"text,omitempty"`
	Title      string   `json:"title,omitempty"`
	Fallback   string   `json:"fallback"`
	Pretext    string   `json:"pretext,omitempty"`
	Color      string   `json:"color,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	AuthorLink string   `json:"author_link,omitempty"`
	AuthorIcon string   `json:"author_icon,omitempty"`
	Footer     string   `json:"footer,omitempty"`
	Fields     []Field  `json:"fields,omitempty"`
	Actions    []Action `json:"actions,omitempty"`
}

type Field struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

type Action struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
	Url  string `json:"url,omitempty"`
}

func main() {
	endpoint := os.Getenv(EnvSlackWebhook)
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "URL is required")
		os.Exit(1)
	}
	text := os.Getenv(EnvSlackMessage)
	if text == "" {
		fmt.Fprintln(os.Stderr, "Message is required")
		os.Exit(1)
	}
	ref := os.Getenv("GITHUB_REF")
	refStart := 10
	if strings.Contains(ref, "heads") {
		refStart = 11	
	}
	refShort := ref[refStart:len(ref)]

	fields:= []Field{
		{
			Title: "Version",
			Value: os.Getenv(EnvSlackTitle),
			Short: true,
		},
		{
			Title: "Variants",
			Value: os.Getenv(EnvVariantsList),
			Short: true,
		},
		{
			Title: "Built from",
			Value: refShort,
			Short: true,
		},
		{
			Title: "Triggered by",
			Value: envOr(EnvGithubActor, "N/A"),
			Short: true,
		},
	}
	actions:= []Action {
		{
			Type: "button",
			Text: "Changelog",
			Url: os.Getenv(EnvChangeLogUrl),
		},
		{
			Type: "button",
			Text: "Downloads",
			Url: os.Getenv(EnvReleasesUrl),
		},
	}

	msg := Webhook{
		UserName: os.Getenv(EnvSlackUserName),
		IconURL:  os.Getenv(EnvSlackIcon),
		Channel:  os.Getenv(EnvSlackChannel),
		Attachments: []Attachment{
			{
				Fallback: envOr(EnvSlackMessage, "GITHUB_ACTION=" + os.Getenv("GITHUB_ACTION") + " \n GITHUB_ACTOR=" + os.Getenv("GITHUB_ACTOR") + " \n GITHUB_EVENT_NAME=" + os.Getenv("GITHUB_EVENT_NAME") + " \n GITHUB_REF=" + os.Getenv("GITHUB_REF") + " \n GITHUB_REPOSITORY=" + os.Getenv("GITHUB_REPOSITORY") + " \n GITHUB_WORKFLOW=" + os.Getenv("GITHUB_WORKFLOW")),
				Color:    envOr(EnvSlackColor, "good"),
				Pretext:  envOr(EnvSlackPretext, ""),
				Footer:   envOr(EnvSlackFooter, ""),
				Fields:   fields,
				Actions:  actions,
			},
		},
	}

	if err := send(endpoint, msg); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %s\n", err)
		os.Exit(2)
	}
}

func envOr(name, def string) string {
	if d, ok := os.LookupEnv(name); ok {
		return d
	}
	return def
}

func send(endpoint string, msg Webhook) error {
	enc, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(enc)
	res, err := http.Post(endpoint, "application/json", b)
	if err != nil {
		return err
	}

	if res.StatusCode >= 299 {
		return fmt.Errorf("Error on message: %s\n", res.Status)
	}
	fmt.Println(res.Status)
	return nil
}
