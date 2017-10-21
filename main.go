package main

import (
	"fmt"
	"net/http"
	"html"
	"io/ioutil"
	"strings"
	"os"
	"github.com/valyala/fasthttp"
	"time"
	"github.com/tidwall/gjson"
	"github.com/boltdb/bolt"
	"github.com/keegancsmith/shell"
	"net/http/httputil"
	"net/url"
)

var db *bolt.DB
var pushed_bk = []byte("repo_pushed")
var root = "/srv/git/"
const (
	github = "https://github.com/"
)

var client = &fasthttp.Client{
	MaxConnsPerHost: 1000,
}

var reverseproxy *httputil.ReverseProxy

func sendHTTPRequest(method, url string) string {
	//now := time.Now().Unix()
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.Header.SetUserAgent("Gitbot/4.012")
	res := fasthttp.AcquireResponse()

	err := client.DoTimeout(req, res, 4 * time.Second)
	if err != nil {
		fmt.Println(err)
	}
	return string(res.Body())
}


func isRepoExist(root, repo string) bool {
	_, err := os.Stat(root + repo + "/.git")
	return err == nil
}

func updateLastUpdated(repo string, t time.Time) (err error) {
	var ut time.Time
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(pushed_bk)
		utb := b.Get([]byte(repo))
		ut, _ = time.Parse(time.RFC3339Nano, string(utb))
		return nil
	})

	if err != nil {
		return err
	}

	if t.Sub(ut).Nanoseconds() < 0 {
		return nil
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(pushed_bk)
		err := b.Put([]byte(repo), []byte(t.Format(time.RFC3339Nano)))
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func updateIfOutdated(root, repo string) {
	if !shouldUpdate(root, repo) {
		return
	}

	err := fetchRepo(root, repo)
	if err != nil {
		fmt.Println(err)
	}
	updateLastUpdated(repo, time.Now())
}

func shouldUpdate(root, repo string) bool {
	data := sendHTTPRequest("GET", "https://api.github.com/repos/" + repo)
	pushed := gjson.Get(data, "pushed_at")
	t, err := time.Parse(time.RFC3339Nano, pushed.Str)
	if err != nil {
		fmt.Println(err)
		return true
	}
	var sh bool
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(pushed_bk)
		utb := b.Get([]byte(repo))
		if len(utb) == 0 {
			sh = true
			return nil
		}
		ut, _ := time.Parse(time.RFC3339Nano, string(utb))
		if ut.Sub(t).Seconds() < 3 {
			sh = true
			return nil
		}
		sh = false
		return nil
	})

	if err != nil {
		return true
	}

	return sh
}

func cloneRepo(root, repo string) {
	cmd := shell.Commandf("git clone %s %s", github + repo, root + repo)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func fetchRepo(root, repo string) error {
	println("fetching")
	cmd := shell.Commandf("cd %s && for remote in `git branch -r `; do git branch --track $remote; done || git remote update && git pull --all", root + repo)
	return cmd.Run()
}

func extractRepo(url string) string {
	repo := strings.Split(url, "/info/")[0]
	return repo[1:]
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		repo := extractRepo(r.URL.Path)
		if isRepoExist(root, repo) {
			updateIfOutdated(root, repo)
		} else {
			cloneRepo(root, repo)
		}
	}
	reverseproxy.ServeHTTP(w, r)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	var root = "/git"
	//flusher := w.(http.Flusher)
	repochan := getHead(root, listAllGitRepo(root))
	for r := range repochan {
		fmt.Fprint(w, "<code>" + html.EscapeString(r.commit)  + "</code>&nbsp;" + html.EscapeString(r.name) + "<br/>")
	}
}

type Repo struct {
	name, commit string
}

func getHead(root string, in <-chan string) <-chan Repo {
	repochan := make(chan Repo)
	go func() {
		for rname := range in {
			repochan <- Repo{
				name: rname,
				commit: getLatestCommit(root + "/" + rname),
			}
		}
		close(repochan)
	}()
	return repochan
}

func listAllGitRepo(root string) <-chan string {
	repochan := make(chan string)
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				fmt.Println(r)
			}
			close(repochan)
		}()
		files, err := ioutil.ReadDir(root)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, f := range files {
			if !f.IsDir() {
				continue
			}

			repfiles, err := ioutil.ReadDir(root + "/" + f.Name())
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, r := range repfiles {
				if !f.IsDir() {
					continue
				}
				repochan <- f.Name() + "/" + r.Name()
			}
		}
	}()
	return repochan
}

func getLatestCommit(repo string) string {
	head, err := ioutil.ReadFile(repo + ".git/HEAD")
	if err != nil {
		return ""
	}
	commit, err := ioutil.ReadFile(repo + ".git/" + string(head))
	if err != nil {
		return ""
	}
	return string(commit)
}

func setupDB() *bolt.DB {
	db, err := bolt.Open("git.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(pushed_bk)
		return err
	})
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	reurl, err := url.Parse("http://127.0.0.1:12085/")
	if err != nil {
		panic(err)
	}
	reverseproxy = httputil.NewSingleHostReverseProxy(reurl)
	db = setupDB()
	defer db.Close()
	http.HandleFunc("/", handler)
	//http.HandleFunc("/list", listHandler)
	println("server is running at port 10292")
	http.ListenAndServe(":10292", nil)
}
