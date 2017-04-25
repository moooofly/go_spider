// Package page contains result catched by Downloader.
// And it alse has result parsed by PageProcesser.
package page

import (
    "github.com/PuerkitoBio/goquery"
    "github.com/bitly/go-simplejson"
    "github.com/moooofly/go_spider/core/common/mlog"
    "github.com/moooofly/go_spider/core/common/page_items"
    "github.com/moooofly/go_spider/core/common/request"
    "net/http"
    "strings"
    //"fmt"
)

// Page 对应被爬取的实体对象
type Page struct {
    // The isfail is true when crawl process is failed and errormsg is the fail reason.
    isfail   bool
    errormsg string

    // The request is crawled by spider that contains url and relevant information.
    req *request.Request    // 封装 HTTP request 相关内容
    body string             // 保存爬取到的明文结果，处理过程 resp.Body =>(UTF-8)=> destbody =>(doc.Html())=> body

    rspHeader  http.Header  // 保存 response 中的 header
    cookies []*http.Cookie  // 保存 response 中的 cookie

    // a pointer of goquery object that contains html result.
    docParser *goquery.Document // Document represents an HTML document to be manipulated

    // The jsonMap is the json result.
    jsonMap *simplejson.Json

    // 记录所有在 PageProcesser 中解析得到的 k/v 内容
    pItems *page_items.PageItems

    // 缓存等待添加到 scheduler 中的所有 request
    targetRequests []*request.Request
}

// NewPage returns initialized Page object.
// NOTE: 初始化 Page 对象；可以看出 request 被同时保存到 Page 和 PageItems 中
func NewPage(req *request.Request) *Page {
    return &Page{pItems: page_items.NewPageItems(req), req: req}
}

// SetHeader save the header of http response
func (this *Page) SetRspHeader(header http.Header) {
    this.rspHeader = header
}

// GetHeader returns the header of http response
func (this *Page) GetRspHeader() http.Header {
    return this.rspHeader
}

// SetHeader save the cookies of http response
func (this *Page) SetCookies(cookies []*http.Cookie) {
    this.cookies = cookies
}

// GetHeader returns the cookies of http response
func (this *Page) GetCookies() []*http.Cookie {
    return this.cookies
}

// IsSucc test whether download process success or not.
func (this *Page) IsSucc() bool {
    return !this.isfail
}

// Errormsg show the download error message.
func (this *Page) Errormsg() string {
    return this.errormsg
}

// SetStatus save status info about download process.
func (this *Page) SetStatus(isfail bool, errormsg string) {
    this.isfail = isfail
    this.errormsg = errormsg
}

// AddField saves KV string pair to PageItems preparing for Pipeline
func (this *Page) AddField(key string, value string) {
    this.pItems.AddItem(key, value)
}

// GetPageItems returns PageItems object that record KV pair parsed in PageProcesser.
func (this *Page) GetPageItems() *page_items.PageItems {
    return this.pItems
}

// SetSkip set label "skip" of PageItems.
// PageItems will not be saved in Pipeline when skip is set true
func (this *Page) SetSkip(skip bool) {
    this.pItems.SetSkip(skip)
}

// GetSkip returns skip label of PageItems.
func (this *Page) GetSkip() bool {
    return this.pItems.GetSkip()
}

// SetRequest saves request object of this page.
func (this *Page) SetRequest(r *request.Request) *Page {
    this.req = r
    return this
}

// GetRequest returns request object of this page.
func (this *Page) GetRequest() *request.Request {
    return this.req
}

// GetUrlTag returns name of url.
func (this *Page) GetUrlTag() string {
    return this.req.GetUrlTag()
}

// AddTargetRequest adds one new Request waiting for crawl.
// 将 request 保存到 targetRequest 中等待爬取
func (this *Page) AddTargetRequest(url string, respType string) *Page {
    this.targetRequests = append(this.targetRequests, request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil))
    return this
}

// AddTargetRequests adds new Requests waiting for crawl.
func (this *Page) AddTargetRequests(urls []string, respType string) *Page {
    for _, url := range urls {
        this.AddTargetRequest(url, respType)
    }
    return this
}

// AddTargetRequestWithProxy adds one new Request waiting for crawl.
func (this *Page) AddTargetRequestWithProxy(url string, respType string, proxyHost string) *Page {

    this.targetRequests = append(this.targetRequests, request.NewRequestWithProxy(url, respType, "", "GET", "", nil, nil, proxyHost, nil, nil))
    return this
}

// AddTargetRequestsWithProxy adds new Requests waiting for crawl.
func (this *Page) AddTargetRequestsWithProxy(urls []string, respType string, proxyHost string) *Page {
    for _, url := range urls {
        this.AddTargetRequestWithProxy(url, respType, proxyHost)
    }
    return this
}

// AddTargetRequest adds one new Request with header file for waiting for crawl.
func (this *Page) AddTargetRequestWithHeaderFile(url string, respType string, headerFile string) *Page {
    this.targetRequests = append(this.targetRequests, request.NewRequestWithHeaderFile(url, respType, headerFile))
    return this
}

// AddTargetRequest adds one new Request waiting for crawl.
// The respType is "html" or "json" or "jsonp" or "text".
// The urltag is name for marking url and distinguish different urls in PageProcesser and Pipeline.
// The method is POST or GET.
// The postdata is http body string.
// The header is http header.
// The cookies is http cookies.
// 只用于添加外部构建好的、具有 param 的 request
func (this *Page) AddTargetRequestWithParams(req *request.Request) *Page {
    this.targetRequests = append(this.targetRequests, req)
    return this
}

// AddTargetRequests adds new Requests waiting for crawl.
func (this *Page) AddTargetRequestsWithParams(reqs []*request.Request) *Page {
    for _, req := range reqs {
        this.AddTargetRequestWithParams(req)
    }
    return this
}

// GetTargetRequests returns the target requests that will put into Scheduler
func (this *Page) GetTargetRequests() []*request.Request {
    return this.targetRequests
}

// SetBodyStr saves plain string crawled in Page.
func (this *Page) SetBodyStr(body string) *Page {
    this.body = body
    return this
}

// GetBodyStr returns plain string crawled.
func (this *Page) GetBodyStr() string {
    return this.body
}

// SetHtmlParser saves goquery object bound to target crawl result.
func (this *Page) SetHtmlParser(doc *goquery.Document) *Page {
    this.docParser = doc
    return this
}

// GetHtmlParser returns goquery object bound to target crawl result.
func (this *Page) GetHtmlParser() *goquery.Document {
    return this.docParser
}

// GetHtmlParser returns goquery object bound to target crawl result.
func (this *Page) ResetHtmlParser() *goquery.Document {
    r := strings.NewReader(this.body)
    var err error
    this.docParser, err = goquery.NewDocumentFromReader(r)
    if err != nil {
        mlog.LogInst().LogError(err.Error())
        panic(err.Error())
    }
    return this.docParser
}

// SetJson saves json result.
func (this *Page) SetJson(js *simplejson.Json) *Page {
    this.jsonMap = js
    return this
}

// SetJson returns json result.
func (this *Page) GetJson() *simplejson.Json {
    return this.jsonMap
}
