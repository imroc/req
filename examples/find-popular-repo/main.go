package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/imroc/req/v3"
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
	req.EnableDumpAllWithoutBody().EnableDebugLog().EnableTraceAll()
}

type Repo struct {
	Name string `json:"name"`
	Star int    `json:"stargazers_count"`
}
type ErrorMessage struct {
	Message string `json:"message"`
}

func findTheMostPopularRepo(username string) (repo string, star int, err error) {

	var popularRepo Repo
	var resp *req.Response

	for page := 1; ; page++ {
		repos := []*Repo{}
		errMsg := ErrorMessage{}
		resp, err = req.SetHeader("Accept", "application/vnd.github.v3+json").
			SetQueryParams(map[string]string{
				"type":      "owner",
				"page":      strconv.Itoa(page),
				"per_page":  "100",
				"sort":      "updated",
				"direction": "desc",
			}).
			SetPathParam("username", username).
			SetResult(&repos).
			SetError(&errMsg).
			Get("https://api.github.com/users/{username}/repos")

		fmt.Println("TraceInfo:")
		fmt.Println("----------")
		fmt.Println(resp.TraceInfo())
		fmt.Println()

		if err != nil {
			return
		}

		if resp.IsSuccess() { //  HTTP status `code >= 200 and <= 299` is considred as success
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
		} else if resp.IsError() { // HTTP status `code >= 400` is considred as an error
			// Extract the error message, wrap and return err
			err = errors.New(errMsg.Message)
			return
		}

		// Unkown http status code, record and return error, here we can use
		// String() to get response body, cuz response body have already been read
		// and no error returned, do not need to use ToString().
		err = fmt.Errorf("unknown error. status code %d; body: %s", resp.StatusCode, resp.String())
		return
	}
}
