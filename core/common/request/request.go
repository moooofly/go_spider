// Package request implements request entity contains url and other relevant information.
package request

import (
    "github.com/bitly/go-simplejson"
    "io/ioutil"
    "net/http"
    "os"

    "github.com/moooofly/go_spider/core/common/mlog"
)

// Request represents object waiting for being crawled.
type Request struct {
    Url string       // 当前 HTTP request 的目标 URL
    RespType string  // 指定的 response type (html/json/jsonp/text)
    Method string    // 当前 HTTP request 使用的 method (GET or POST)
    Postdata string  // 当发起 POST request 时使用
    Urltag string    // 用于在 PageProcesser 和 Pipeline 中区分不同 URL 的 tag
    Header http.Header      // 设置到 HTTP request 中的 header
    Cookies []*http.Cookie  // 设置到 HTTP request 中的 cookie(s)
    ProxyHost string        // 用于指定 Proxy 地址，例如 'localhost:80'

    // Redirect function for downloader used in http.Client
    // If CheckRedirect returns an error, the Client's Get
    // method returns both the previous Response.
    // If CheckRedirect returns error.New("normal"), the error process after client.Do will ignore the error.
    checkRedirect func(req *http.Request, via []*http.Request) error

    Meta interface{}
}

// NewRequest returns initialized Request object.
// The respType is json, jsonp, html, text
/*
func NewRequestSimple(url string, respType string, urltag string) *Request {
    return &Request{url:url, respType:respType}
}
*/

func NewRequest(url string, respType string, urltag string, method string,
    postdata string, header http.Header, cookies []*http.Cookie,
    checkRedirect func(req *http.Request, via []*http.Request) error,
    meta interface{}) *Request {
    return &Request{url, respType, method, postdata, urltag, header, cookies, "", checkRedirect, meta}
}

func NewRequestWithProxy(url string, respType string, urltag string, method string,
    postdata string, header http.Header, cookies []*http.Cookie, proxyHost string,
    checkRedirect func(req *http.Request, via []*http.Request) error,
    meta interface{}) *Request {
    return &Request{url, respType, method, postdata, urltag, header, cookies, proxyHost, checkRedirect, meta}
}

func NewRequestWithHeaderFile(url string, respType string, headerFile string) *Request {
    _, err := os.Stat(headerFile)
    if err != nil {
        //file is not exist , using default mode
        return NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
    }

    h := readHeaderFromFile(headerFile)

    return NewRequest(url, respType, "", "GET", "", h, nil, nil, nil)
}

func readHeaderFromFile(headerFile string) http.Header {
    //read file , parse the header and cookies
    b, err := ioutil.ReadFile(headerFile)
    if err != nil {
        //make be:  share access error
        mlog.LogInst().LogError(err.Error())
        return nil
    }
    js, _ := simplejson.NewJson(b)
    //constructed to header

    h := make(http.Header)
    h.Add("User-Agent", js.Get("User-Agent").MustString())
    h.Add("Referer", js.Get("Referer").MustString())
    h.Add("Cookie", js.Get("Cookie").MustString())
    h.Add("Cache-Control", "max-age=0")
    h.Add("Connection", "keep-alive")
    return h
}

//point to a json file
/* xxx.json
{
	"User-Agent":"curl/7.19.3 (i386-pc-win32) libcurl/7.19.3 OpenSSL/1.0.0d",
	"Referer":"http://weixin.sogou.com/gzh?openid=oIWsFt6Sb7aZmuI98AU7IXlbjJps",
	"Cookie":""
}
*/
func (this *Request) AddHeaderFile(headerFile string) *Request {
    _, err := os.Stat(headerFile)
    if err != nil {
        return this
    }
    h := readHeaderFromFile(headerFile)
    this.Header = h
    return this
}

// @host  http://localhost:8765/
func (this *Request) AddProxyHost(host string) *Request {
    this.ProxyHost = host
    return this
}

func (this *Request) GetUrl() string {
    return this.Url
}

func (this *Request) GetUrlTag() string {
    return this.Urltag
}

func (this *Request) GetMethod() string {
    return this.Method
}

func (this *Request) GetPostdata() string {
    return this.Postdata
}

func (this *Request) GetHeader() http.Header {
    return this.Header
}

func (this *Request) GetCookies() []*http.Cookie {
    return this.Cookies
}

func (this *Request) GetProxyHost() string {
    return this.ProxyHost
}

func (this *Request) GetResponseType() string {
    return this.RespType
}

func (this *Request) GetRedirectFunc() func(req *http.Request, via []*http.Request) error {
    return this.checkRedirect
}

func (this *Request) GetMeta() interface{} {
    return this.Meta
}
