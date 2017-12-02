package git

import (
	"github.com/tidwall/gjson"
	"time"
	"github.com/thanhpk/log"
	"github.com/keegancsmith/shell"
)

type OriginCf struct {
	root, email, password string
}

const (
	origincfhost = "http://origin.cf/"
	origincfapihost = "http://origin.cf/api/v1/repos/"
)

func NewOriginCfAPI(root, email, password string) *OriginCf {
	return &OriginCf{
		root: root,
		email: email,
		password: password,
	}
}

func (m OriginCf) GetService() string {
	return "origin.cf"
}

func (m OriginCf) LastUpdate(repo string) time.Time {
	var data string
	var err error
	for i := 0; i < 3; i++ {
		data, err = sendHTTPRequest("GET", origincfapihost + repo, m.email, m.password)
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Log(err)
		return time.Time{}
	}

	updated := gjson.Get(data, "updated_at")
	t, err := time.Parse(time.RFC3339Nano, updated.Str)
	if err != nil {
		log.Log(err)
		return time.Time{}
	}
	return t
}

func (m OriginCf) PullRepo(repo string) error {
	cmd := shell.Commandf("cd %s && git fetch --all && for i in $(git branch -r | grep  -v HEAD | sed -e 's/origin\\///'); do git checkout $i && git reset --hard origin/$i; done", m.root + repo)
	return cmd.Run()
}

func (m OriginCf) CloneRepo(repo string) error {
	log.Log(repo)
	cmd := shell.Commandf("git clone http://%s:%s@origin.cf/%s.git %s", m.email, m.password, repo, m.root + repo)
	out, err :=  cmd.Output()
	if err != nil {
		log.Logf("git clone https://%s:%s@origin.cf/%s.git %s", m.email, m.password, repo, m.root + repo)

		log.Log(string(out))
	}
//	o, _ := cmd.Output()
	return err
}
