package git

import (
	"github.com/tidwall/gjson"
	"time"
	"github.com/thanhpk/log"
	"github.com/keegancsmith/shell"
	"fmt"
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

	data, err := sendHTTPRequest("GET", bbapihost + "user", email, password)
	if err != nil {
		panic(err)
	}
	username := gjson.Get(data, "username").Str
	if username == "" {
		panic("invalid email and password, cannot get bitbucket username")
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
	var data string
	var err error
	for i := 0; i < 3; i++ {
		data, err = sendHTTPRequest("GET", bbapihost + "repositories/" + repo, m.email, m.password)
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Log(err)
		return time.Time{}
	}

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
	cmd := shell.Commandf("git clone %s/%s.git %s", m.GetAuthUrl(), repo, m.root + repo)
	out, err :=  cmd.Output()
	if err != nil {
		log.Logf("git clone %s/%s.git %s", m.GetAuthUrl(), repo, m.root + repo)

		log.Log(string(out))
	}
//	o, _ := cmd.Output()
	return err
}

func (m Bb) GetAuthUrl() string {
	return fmt.Sprintf("https://%s:%s@bitbucket.org", m.username, m.password)
}

func (m Bb) GetAuth() (string, string) {
	return m.username, m.password
}
