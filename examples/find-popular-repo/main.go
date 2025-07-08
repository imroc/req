package main

import (
	"fmt"
	"strconv"

	"github.com/0xobjc/req/v3"
)

// Change the name if you want
var username = "imroc"

func main() {
	repo, star, err := findTheMostPopularRepo(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("The most popular repo of %s is %s, which have %d stars\n", username, repo, star)
}

func init() {
	req.EnableDebugLog().
		EnableTraceAll().
		EnableDumpEachRequest().
		SetCommonErrorResult(&ErrorMessage{}).
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if resp.Err != nil {
				return nil
			}
			if errMsg, ok := resp.ErrorResult().(*ErrorMessage); ok {
				resp.Err = errMsg
				return nil
			}
			if !resp.IsSuccessState() {
				resp.Err = fmt.Errorf("bad status: %s\nraw content:\n%s", resp.Status, resp.Dump())
			}
			return nil
		})
}

type Repo struct {
	Name string `json:"name"`
	Star int    `json:"stargazers_count"`
}
type ErrorMessage struct {
	Message string `json:"message"`
}

func (msg *ErrorMessage) Error() string {
	return fmt.Sprintf("API Error: %s", msg.Message)
}

func findTheMostPopularRepo(username string) (repo string, star int, err error) {
	var popularRepo Repo
	var resp *req.Response

	for page := 1; ; page++ {
		repos := []*Repo{}
		resp, err = req.SetHeader("Accept", "application/vnd.github.v3+json").
			SetQueryParams(map[string]string{
				"type":      "owner",
				"page":      strconv.Itoa(page),
				"per_page":  "100",
				"sort":      "updated",
				"direction": "desc",
			}).
			SetPathParam("username", username).
			SetSuccessResult(&repos).
			Get("https://api.github.com/users/{username}/repos")

		fmt.Println("TraceInfo:")
		fmt.Println("----------")
		fmt.Println(resp.TraceInfo())
		fmt.Println()

		if err != nil {
			return
		}

		if !resp.IsSuccessState() { //  HTTP status `code >= 200 and <= 299` is considered as success by default
			return
		}
		for _, repo := range repos {
			if repo.Star >= popularRepo.Star {
				popularRepo = *repo
			}
		}
		if len(repo) == 100 { // Try Next page
			continue
		}
		// All repos have been traversed, return the final result
		repo = popularRepo.Name
		star = popularRepo.Star
		return
	}
}
