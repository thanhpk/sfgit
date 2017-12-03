# Sfgit
![](http://media.esports.vn/images/article_avartar/Avatar.nghi-van-dut-cap-aag-Copy.jpg)
Sfgit statically cache your repo, help you develop your app locally

## Dependencies
- go 1.8
- docker 17.09

## Installation

First, clone the repo:
```sh
  git clone https://github.com/thanhpk/sfgit
  cd sfgit
```

Create config file at sfgit root: `config.yaml` with the following content:
```yaml
  database: pushed_db
  root: /srv/git/
  bitbucket:
    email: your_bitbucket_email@gmail.com
    password: your_bitbucket_password
  github:
    email: your_github_email@gmail.com
    password: your_github_password
  origincf:
    username: your_origin_username
    password: your_origin_password
  listenport: 10292
  gitport: 12085
```
Then, tell your git to use sfgit instead of directly connect to remote host by append to your your `~/.gitconfig` file:
```
  [url "ssh://git@origin.cf:10022/"]
    insteadOf = https://bitbucket.org/

  [url "http://origin.dev/"]
    insteadOf = http://origin.cf/

  [url "http://github.dev/"]
    insteadOf = https://github.com/

  [url "http://github.dev/"]
    insteadOf = http://github.com/

  [url "http://github.dev/"]
    insteadOf = git://github.com/

  [url "http://github.dev/"]
    insteadOf = git@github.com:
```
D32 2:31:23 sfgit/main.go:189 server is running at port 10292

Now, run the service
```sh
  ./up
```
you should see the server is running at port 10292:
```
D32 2:31:23 sfgit/main.go:189 server is running at port 10292
```

Now test it with some thing really heavy
```sh
  git clone https://github.com/moby/moby
  # took a while for the first time
  git rm -rf moby
  git clone https://github.com/moby/moby
  # super fast for the second time.
```

## Architecture

![](https://www.lucidchart.com/publicSegments/view/3452ce5d-8337-4632-9fcc-7840cc48df78/image.png)
