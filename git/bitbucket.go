package git

import (
	"github.com/tidwall/gjson"
	"time"
	"github.com/thanhpk/log"
	"github.com/keegancsmith/shell"
)

type Bb struct {
	root, username, password, email string
}

const (
	bbhost = "https://bitbucket.com/"
	bbapihost =  "https://api.bitbucket.org/2.0/"
)

func NewBitbucketAPI(root, email, password string) *Bb {
	if email == "" || password == "" {
		panic("username and password must not be empty")
	}

	data := sendHTTPRequest("GET", bbapihost + "user", email, password)
	username := gjson.Get(data, "username").Str
	if username == "" {
		panic("invalid email and password, cannot get username")
	}

	return &Bb{
		username: username,
		email: email,
		password: password,
		root: root,
	}
}

func (m Bb) GetService() string {
	return "bitbucket.org"
}

func (m Bb) LastUpdate(repo string) time.Time {
	data := sendHTTPRequest("GET", bbapihost + "repositories/" + repo, m.email, m.password)
	pushed := gjson.Get(data, "updated_on")
	t, err := time.Parse(time.RFC3339Nano, pushed.Str)
	if err != nil {
		log.Log(err)
		return time.Time{}
	}
	return t
}

func (m Bb) PullRepo(repo string) error {
	cmd := shell.Commandf("cd %s && git fetch --all && for i in $(git branch -r | grep  -v HEAD | sed -e 's/origin\\///'); do git checkout $i && git reset --hard origin/$i; done", m.root + repo)
	return cmd.Run()
}

func (m Bb) CloneRepo(repo string) error {
	cmd := shell.Commandf("git clone https://%s:%s@bitbucket.org/%s.git %s", m.username, m.password, repo, m.root + repo)
	out, err :=  cmd.Output()
	if err != nil {
		log.Logf("git clone https://%s:%s@bitbucket.org/%s.git %s", m.username, m.password, repo, m.root + repo)

		log.Log(string(out))
	}
//	o, _ := cmd.Output()
	return err
}
