package git

import (
	"github.com/tidwall/gjson"
	"time"
	"github.com/thanhpk/log"
	"github.com/keegancsmith/shell"
	"os"
)

type Gh struct {
	root, email, password string
}

const (
	ghhost = "https://github.com/"
	ghapihost =  "https://api.github.com/repos/"
)

func NewGithubAPI(root, email, password string) *Gh {
	return &Gh{
		root: root,
		email: email,
		password: password,
	}
}

func (m Gh) GetService() string {
	return "github.com"
}

func (m Gh) LastUpdate(repo string) time.Time {
	// skip existed repo
	_, err := os.Stat(m.root + "/" +  repo + "/.git")
	if err == nil {
		return time.Time{}
	}
	var data string
	for i := 0; i < 3; i++ {
		data, err = sendHTTPRequest("GET", ghapihost + repo, m.email, m.password)
		if err != nil {
			break
		}
		time.Sleep(time.Duration(i + 1) * time.Second)
	}
	if err != nil {
		log.Log(err)
		return time.Time{}
	}

	pushed := gjson.Get(data, "pushed_at")
	t, err := time.Parse(time.RFC3339Nano, pushed.Str)
	if err != nil {
		log.Log(data)
		log.Log(repo, err)
		return time.Time{}
	}
	return t
}

func (m Gh) PullRepo(repo string) error {
	cmd := shell.Commandf("cd %s && git fetch --all && for i in $(git branch -r | grep  -v HEAD | sed -e 's/origin\\///'); do git checkout $i && git reset --hard origin/$i; done", m.root + repo)
	return cmd.Run()
}

func (m Gh) CloneRepo(repo string) error {
	cmd := shell.Commandf("git clone %s %s", ghhost + repo, m.root + repo)
	out, err := cmd.Output()
	if err != nil {
		log.Log(out)
	}
	return err
}
