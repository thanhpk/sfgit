package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"net/http/httputil"
	"net/url"
	"github.com/thanhpk/sfgit/db"
	"github.com/thanhpk/sfgit/git"
	"github.com/jinzhu/configor"
)

type DB interface {
	IsRepoExists(service, repo string) bool
	Close()
	Touch(service, repo string) error
	LastUpdate(service, repo string) time.Time
}

type API interface {
	GetService() string
	LastUpdate(repo string) time.Time
	//ShouldUpdate(repo string) bool
	PullRepo(repo string) error
	CloneRepo(repo string) error
}

var store DB
var github API
var bitbucket API

func shouldUpdate(db DB, api API, service, repo string) bool {
	t := api.LastUpdate(repo)
	ut := db.LastUpdate(service, repo)
	if ut.Sub(t).Seconds() < 3 {
		return true
	}
	return false
}

func updateIfOutdated(db DB, api API, service, repo string) {
	if !shouldUpdate(db, api, service, repo) {
		return
	}
	err := api.PullRepo(repo)
	if err != nil {
		fmt.Println(err)
	}
	db.Touch(service, repo)
}

func extractRepo(url string) string {
	repo := strings.Split(url, "/info/")[0]
	return repo[1:]
}

func handler(w http.ResponseWriter, r *http.Request) {
	var api API
	var reurl *url.URL
	if strings.Contains(r.Host, "github") {
		api = github
	} else if strings.Contains(r.Host, "bitbucket") {
		api = bitbucket
	} else {
		http.Error(w, "unsupported service " + r.Host, 400)
		return
	}
	reurl, _ = url.Parse("http://127.0.0.1:" + Config.GitPort + "/" + api.GetService() + "/")


	if r.Method == "GET" {
		repo := extractRepo(r.URL.Path)
		if store.IsRepoExists(api.GetService(), repo) {
			updateIfOutdated(store, api, api.GetService(), repo)
		} else {
			err := api.CloneRepo(repo)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	reverseproxy := httputil.NewSingleHostReverseProxy(reurl)
	reverseproxy.ServeHTTP(w, r)
}

var Config = struct {
	Database string `default:"pushed_bk"`
	Root string `default:"/srv/git/"`
	Bitbucket struct {
		Email, Password string
	}
	ListenPort string `default:"10292"` // if this change, git.conf should change too
	GitPort string `default:"12085"` // if this change, git.conf should change too
}{}

func main() {
	configor.Load(&Config, "./config.yaml")
	store = db.NewLocalDB(Config.Root, Config.Database)
	defer store.Close()

	github = git.NewGithubAPI(Config.Root + "github.com/")
	bitbucket = git.NewBitbucketAPI(Config.Root + "bitbucket.org/", Config.Bitbucket.Email, Config.Bitbucket.Password)

	http.HandleFunc("/", handler)
	fmt.Println("server is running at port " + Config.ListenPort)
	http.ListenAndServe(":" + Config.ListenPort, nil)
}
