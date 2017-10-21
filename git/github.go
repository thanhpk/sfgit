package git

import (
	"github.com/tidwall/gjson"
	"time"
	"fmt"
	"github.com/keegancsmith/shell"
)

type Gh struct {
	root string
}

const (
	ghhost = "https://github.com/"
	ghapihost =  "https://api.github.com/repos/"
)

func NewGithubAPI(root string) *Gh {
	return &Gh{
		root: root,
	}
}

func (m Gh) GetService() string {
	return "github.com"
}

func (m Gh) LastUpdate(repo string) time.Time {
	data := sendHTTPRequest("GET", ghapihost + repo)
	pushed := gjson.Get(data, "pushed_at")
	t, err := time.Parse(time.RFC3339Nano, pushed.Str)
	if err != nil {
		fmt.Println("dd", err)
		return time.Time{}
	}
	return t
}

func (m Gh) PullRepo(repo string) error {
	cmd := shell.Commandf("cd %s && for remote in `git branch -r `; do git branch --track $remote; done || git remote update && git pull --all", m.root + repo)
	return cmd.Run()
}

func (m Gh) CloneRepo(repo string) error {
	cmd := shell.Commandf("git clone %s %s", ghhost + repo, m.root + repo)
	return cmd.Run()
}
