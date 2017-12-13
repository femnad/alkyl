package notifications

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	BaseNotificationUrl = "https://api.github.com"
)

type RepositoryType struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

type SubjectType struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type Notification struct {
	Id         string         `json:"id"`
	Url        string         `json:"url"`
	Repository RepositoryType `json:"repository"`
	Subject    SubjectType    `json:"subject"`
}

type Issue struct {
	Body  string `json:"body"`
	Title string `json:"title"`
	Id    uint64 `json:"id"`
}

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func githubRequest(url string, token string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	authHeader := fmt.Sprintf("token %s", token)
	req.Header.Add("Authorization", authHeader)
	resp, err := client.Do(req)
	Check(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	Check(err)

	return body
}

func GetEndpointUrl(endpoint string) string {
	return fmt.Sprintf("%s/%s", BaseNotificationUrl, endpoint)
}

func GetIssuesUrl(repo string) string {
	UrlSuffix := fmt.Sprintf("repos/%s/issues", repo)
	return GetEndpointUrl(UrlSuffix)
}

func GetIssues(repo string, token string) []Issue {
	log.Println("Getting issues")
	url := GetIssuesUrl(repo)
	body := githubRequest(url, token)
	var issues []Issue
	err := json.Unmarshal(body, &issues)
	Check(err)

	return issues
}
