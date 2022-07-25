// Package main provides an utility for list your github followers
// and the ones recently unfollowed you.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	gh "github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

var user = flag.String("user", "hitzhangjie", "github user (specify shell var GITHUB_ACCESS_TOKEN before running")

func init() {
	flag.Parse()
}

func main() {
	// list followers
	followers, err := listFollowers(*user)
	if err != nil {
		if _, ok := err.(*gh.RateLimitError); ok {
			fmt.Println("hit rate limit...wait 1s")
		}
		fmt.Println("list followers fail: %v", err)
		os.Exit(1)
	}

	// record current result
	fmt.Printf("%s has %d followers:\n", *user, len(followers))
	// record current followers
	//prettyprint(followers)
	if err := recordFollowers(followers); err != nil {
		fmt.Printf("record followers fail: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("recording current followers done")

	// read last recorded result
	pattern := fmt.Sprintf("%s/followers.*", configdir())
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("glob followers fail: %v\n", err)
		os.Exit(1)
	}
	// if we want to diff, at least 2 files should be found
	if len(matches) <= 1 {
		fmt.Println("last recorded followers not found")
		fmt.Println("diff cannot be finished...exit")
		return
	}

	// parse last recorded result
	sort.Strings(matches)
	last := matches[len(matches)-2]
	buf, err := ioutil.ReadFile(last)
	if err != nil {
		fmt.Printf("read %s fail: %v", last, err)
		os.Exit(1)
	}

	followersBefore := map[string]struct{}{}
	vals := strings.Split(string(buf), "\n")
	for _, v := range vals {
		v := strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		followersBefore[v] = struct{}{}
	}

	// diff the followers and followersBefore
	unfollowed := []string{}
	newfollowed := []string{}

	for u := range followersBefore {
		if _, ok := followers[u]; !ok {
			unfollowed = append(unfollowed, u)
		}
	}
	for u := range followers {
		if _, ok := followersBefore[u]; !ok {
			newfollowed = append(newfollowed, u)
		}
	}
	fmt.Println("unfollowed users: ", unfollowed)
	fmt.Println("newfollowed users: ", newfollowed)
}

func authenticatedClient() *http.Client {
	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.TODO(), ts)

	return gh.NewClient(tc).Client()
}

func listFollowers(user string) (map[string]struct{}, error) {
	var client = gh.NewClient(authenticatedClient())
	var followers = make(map[string]struct{})
	var pageIdx int = 1
	var pageSize int = 10

	for {
		users, resp, err := client.Users.ListFollowers(context.TODO(), user, &gh.ListOptions{
			Page:    pageIdx,
			PerPage: pageSize,
		})
		if err != nil {
			return nil, err
		}

		for _, u := range users {
			followers[u.GetLogin()] = struct{}{}
		}
		if resp.NextPage == 0 {
			break
		}
		pageIdx = resp.NextPage
	}
	return followers, nil
}

func prettyprint(users map[string]*gh.User) {
	i := 1
	for k := range users {
		fmt.Printf("%03d. %s\n", i, k)
	}
}

func recordFollowers(users map[string]struct{}) error {
	layout := "2006-01-02 03:04:05.999"
	fname := fmt.Sprintf("followers.%s", time.Now().Format(layout))

	dir := configdir()
	_ = os.MkdirAll(dir, os.ModePerm)

	buf := bytes.Buffer{}
	for k := range users {
		fmt.Fprintf(&buf, "%s\n", k)
	}

	fname = filepath.Join(dir, fname)
	return ioutil.WriteFile(fname, buf.Bytes(), 0666)
}

func configdir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(home, ".config/gh-followers")
	_ = os.MkdirAll(dir, os.ModePerm)
	return dir
}
