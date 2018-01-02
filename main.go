package main

import (
	"net/http"
	"strings"
	"time"
	"net/http/httputil"
	"net/url"
	"github.com/thanhpk/log"
	"github.com/thanhpk/sfgit/db"
	"github.com/thanhpk/sfgit/git"
	"github.com/jinzhu/configor"
	"github.com/thanhpk/goslice"
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
	GetAuth() (string, string)
	GetAuthUrl() string
}

var store DB
var github API
var bitbucket API
var origincf API

var pullingrepo = make([]string, 0)
var cloningrepo = make([]string, 0)

func printStatus() {
	log.Logf("cloning %d repos %v\npulling %d repos %v",
		len(cloningrepo), cloningrepo,
		len(pullingrepo), pullingrepo)
}

func addPullingRepo(repo string) {
	pullingrepo = append(pullingrepo, repo)
	printStatus()
}

func removePullingRepo(repo string) {
	pullingrepo = slice.Substract(pullingrepo, []string{repo})
	printStatus()
}

func addCloningRepo(repo string) {
	cloningrepo = append(cloningrepo, repo)
	printStatus()
}

func removeCloningRepo(repo string) {
	cloningrepo = slice.Substract(cloningrepo, []string{repo})
	printStatus()
}

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
	addPullingRepo(repo)
	err := api.PullRepo(repo)
	removePullingRepo(repo)

	if err != nil {
		log.WithStack(err)
	}
	db.Touch(service, repo)
}

func extractRepo(url string) string {
	repo := strings.Join(strings.Split(url, "/")[1:3], "/")
	if strings.HasSuffix(repo, ".git") {
		repo = repo[:len(repo)-4]
	}
	return repo
}

func reverseProxy(w http.ResponseWriter, r *http.Request, u *url.URL, username, password string) {
	reverseproxy := httputil.NewSingleHostReverseProxy(u)
	r.URL.Scheme="https"
	r.URL.Host = u.Host
	r.URL.User = u.User
	r.Host = u.Host
	r.RequestURI = r.URL.String()
	if username != "" {
		r.SetBasicAuth(username, password)
	}
	reverseproxy.ServeHTTP(w, r)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var api API
	var reurl *url.URL
	if strings.Contains(r.Host, "github") {
		api = github
	} else if strings.Contains(r.Host, "bitbucket") {
		api = bitbucket
	} else if strings.Contains(r.Host, "origin") {
		api = origincf
	} else {
		http.Error(w, "unsupported service " + r.Host, 400)
		return
	}
	u, p := api.GetAuth()
	reurl, _ = url.Parse("http://127.0.0.1:" + Config.GitPort + "/" + api.GetService() + "/")
	repo := extractRepo(r.URL.Path)
	log.Log(r.Method, r.Host + "/" + repo, r.URL.String())

	ps := strings.Split(r.URL.String(), "/")
	if r.Method == "GET" && len(ps) > 4 && ps[4] == "refs?service=git-receive-pack" {
		reurl, _ = url.Parse(api.GetAuthUrl())
		reverseProxy(w, r, reurl, u, p)
		return
	}

	if r.Method == "POST" && len(ps) > 4 && ps[4] == "refs?service=git-upload-pack" {
		// add .git to path
		ps[2] = strings.Split(ps[2], ".git")[0]
		r.URL.Path = strings.Join(ps, "/")

		reurl, _ = url.Parse(api.GetAuthUrl())
		reverseProxy(w, r, reurl, u, p)
		return
	}

	// Push to remote
	if r.Method == "POST" && len(ps) > 3 && ps[len(ps) -1] == "git-receive-pack" {
		reurl, _ = url.Parse(api.GetAuthUrl())
		reverseProxy(w, r, reurl, u, p)
		return
	}

	if r.Method == "GET" {
		if store.IsRepoExists(api.GetService(), repo) {
			updateIfOutdated(store, api, api.GetService(), repo)
		} else {
			addCloningRepo(repo)
			err := api.CloneRepo(repo)
			if err != nil {
				log.Log(api.GetService())
				log.WithStack(err, api.GetService())
			}
			removeCloningRepo(repo)
		}
	}
	pathsplit := strings.Split(r.URL.Path, "/")
	r.URL.Path = strings.Join(append([]string{"/" + repo}, pathsplit[3:]...), "/")
	reverseProxy(w, r, reurl, "", "")
}

var Config = struct {
	Database string `default:"pushed_bk"`
	Root string `default:"/srv/git/"`
	Bitbucket struct {
		Email, Password string
	}
	Github struct {
		Email, Password string
	}
	Origincf struct {
		Username, Password string
	}
	ListenPort string `default:"10292"` // if this change, git.conf should change too
	GitPort string `default:"12085"` // if this change, git.conf should change too
}{}

func main() {
	configor.Load(&Config, "./config.yaml")
	store = db.NewLocalDB(Config.Root, Config.Database)
	defer store.Close()
	c := make(chan bool, 3)
	go func() {
		github = git.NewGithubAPI(Config.Root + "github.com/", Config.Github.Email, Config.Github.Password)
		c <- true
	}()
	go func() {
		bitbucket = git.NewBitbucketAPI(Config.Root + "bitbucket.org/", Config.Bitbucket.Email, Config.Bitbucket.Password)
		c <- true
	}()

	go func() {
		origincf = git.NewOriginCfAPI(Config.Root + "origin.cf/", Config.Origincf.Username, Config.Origincf.Password)
		c <- true
	}()
	_, _, _ = <-c, <-c, <-c

	http.HandleFunc("/", handler)
	log.Log("server is running at port " + Config.ListenPort)
	http.ListenAndServe(":" + Config.ListenPort, nil)
}
