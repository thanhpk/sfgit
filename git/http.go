package git

import (
	"github.com/valyala/fasthttp"
	"time"
	"fmt"
	"encoding/base64"
)

var client = &fasthttp.Client{
	MaxConnsPerHost: 1000,
}

func sendHTTPRequest(method, url string, auth ...string) string {
	//now := time.Now().Unix()
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.Header.SetUserAgent("Gitbot/4.012")
	if len(auth) == 2 {
		req.Header.Set("Authorization", "Basic " +
			base64.StdEncoding.EncodeToString([]byte(auth[0] + ":" + auth[1])))
	}
	res := fasthttp.AcquireResponse()

	err := client.DoTimeout(req, res, 4 * time.Second)
	if err != nil {
		fmt.Println(err)
	}
	return string(res.Body())
}
