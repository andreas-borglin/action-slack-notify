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
	EnvSlackWebhook   = "SLACK_WEBHOOK"
	EnvSlackChannel   = "SLACK_CHANNEL"
	EnvSlackTitle     = "SLACK_TITLE"
	EnvSlackMessage   = "SLACK_MESSAGE"
	EnvSlackColor     = "SLACK_COLOR"
	EnvSlackFooter    = "SLACK_FOOTER"
	EnvGithubActor    = "GITHUB_ACTOR"
	EnvVariants       = "VARIANTS"
	EnvEnvironments   = "ENVIRONMENTS"
	EnvChangeLogUrl   = "CHANGELOG_URL"
	EnvReleaseUrl     = "RELEASE_URL"
	EnvSlackPretext   = "SLACK_PRETEXT"
	EnvVersionName    = "VERSION_NAME"
	EnvBaseUrl        = "BASE_URL"
)

type Webhook struct {
	Text        string       `json:"text,omitempty"`
	UserName    string       `json:"username,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback   string  `json:"fallback"`
	Pretext    string  `json:"pretext,omitempty"`
	Color      string  `json:"color,omitempty"`
	AuthorName string  `json:"author_name,omitempty"`
	AuthorLink string  `json:"author_link,omitempty"`
	AuthorIcon string  `json:"author_icon,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
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
	
	ref := os.Getenv("GITHUB_REF")
	refStart := 10
	isTag := true
	if strings.Contains(ref, "heads") {
		refStart = 11	
		isTag = false
	}
	refShort := ref[refStart:len(ref)]

	fields := []Field{}
	
	version := os.Getenv(EnvVersionName)
	if version == "" && isTag {
		version = refShort	
	}
	if version != "" {
		versionfield := Field{
			Title: "Version",
			Value: version,
			Short: true,
		}
		fields = append(fields, versionfield)
	}
	
	variants := os.Getenv(EnvVariants)
	if variants != "" {
		variantfield := Field{
			Title: "Variants",
			Value: variants,
			Short: true,
		}
		fields = append(fields, variantfield)
	}
	
	environments := os.Getenv(EnvEnvironments)
	if environments != "" {
		envfield := Field{
			Title: "Environments",
			Value: environments,
			Short: true,
		}
		fields = append(fields, envfield)
	}
	
	if !isTag {
		builtFromField := Field{
			Title: "Built from",
			Value: refShort,
			Short: true,
		}
		fields = append(fields, builtFromField)
	}
	
	actionedByField := Field{
		Title: "Actioned by",
		Value: envOr(EnvGithubActor, "Unknown"),
		Short: true,
	}
	fields = append(fields, actionedByField)	
	
	baseUrl := os.Getenv(EnvBaseUrl)
	if baseUrl != "" {
		baseUrlField := Field{
			Title: "Base URL",
			Value: baseUrl,
			Short: true,
		}
		fields = append(fields, baseUrlField)
	}
	
	actions := []Action{}
	
	changeLogUrl := os.Getenv(EnvChangeLogUrl)
	if changeLogUrl != "" {
		changeLogUrlAction := Action{
			Type: "button",
			Text: "Changelog",
			Url: changeLogUrl,
		}
		actions = append(actions, changeLogUrlAction)
	}

	releaseUrl := os.Getenv(EnvReleaseUrl)
	if releaseUrl != "" {
		releaseUrlAction := Action{
			Type: "button",
			Text: "View release",
			Url: releaseUrl,
		}
		actions = append(actions, releaseUrlAction)
	}

	msg := Webhook{
		Channel:   os.Getenv(EnvSlackChannel),
		Attachments: []Attachment{
			{
				Fallback:   envOr(EnvSlackMessage, "GITHUB_ACTION="+os.Getenv("GITHUB_ACTION")+" \n GITHUB_ACTOR="+os.Getenv("GITHUB_ACTOR")+" \n GITHUB_EVENT_NAME="+os.Getenv("GITHUB_EVENT_NAME")+" \n GITHUB_REF="+os.Getenv("GITHUB_REF")+" \n GITHUB_REPOSITORY="+os.Getenv("GITHUB_REPOSITORY")+" \n GITHUB_WORKFLOW="+os.Getenv("GITHUB_WORKFLOW")),
				Color:      envOr(EnvSlackColor, "good"),
				Pretext:    envOr(EnvSlackPretext, ""),
				Footer:     envOr(EnvSlackFooter, "AEVI Slack Notification"),
				Fields:     fields,
				Actions:    actions,
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
