//
package main

import (
    "fmt"
    "github.com/moooofly/go_spider/core/common/page"
    "github.com/moooofly/go_spider/core/common/request"
    "github.com/moooofly/go_spider/core/spider"
    "strings"
)

type MyPageProcesser struct {
}

func NewMyPageProcesser() *MyPageProcesser {
    return &MyPageProcesser{}
}

func (this *MyPageProcesser) Process(p *page.Page) {
    if !p.IsSucc() {
        println(p.Errormsg())
        return
    }

    query := p.GetHtmlParser()

    name := query.Find(".lemmaTitleH1").Text()
    name = strings.Trim(name, " \t\n")

    summary := query.Find(".card-summary-content .para").Text()
    summary = strings.Trim(summary, " \t\n")

    // the entity we want to save by Pipeline
    p.AddField("name", name)
    p.AddField("summary", summary)
}

func (this *MyPageProcesser) Finish() {
    fmt.Printf("TODO: before end spider \r\n")
}

func main() {
    sp := spider.NewSpider(NewMyPageProcesser(), "TaskName")

    //  Params:
    //  1. url ==> custom Url.
    //  2. respType ==> "html" or "json" or "jsonp" or "text".
    //  3. urltag ==> name for marking url and distinguish different urls in PageProcesser and Pipeline.
    //  4. method ==> POST or GET.
    //  5. postdata ==> body string sent to server.
    //  6. header ==> header for http request.
    //  7. cookies ==> Cookies
    //  8. checkRedirect ==> Http redirect function
    //  9. meta ==> custom data
    req := request.NewRequest("http://baike.baidu.com/view/1628025.htm?fromtitle=http&fromid=243074&type=syn", "html", "", "GET", "", nil, nil, nil, nil)
    // 内部会针对当前 URL 运行一次 spider
    pageItems := sp.GetByRequest(req)
    //pageItems := sp.Get("http://baike.baidu.com/view/1628025.htm?fromtitle=http&fromid=243074&type=syn", "html")

    url := pageItems.GetRequest().GetUrl()
    println("-----------------------------------spider.Get---------------------------------")
    println("url\t:\t" + url)
    for name, value := range pageItems.GetAll() {
        println(name + "\t:\t" + value)
    }

    println("\n--------------------------------spider.GetAll---------------------------------")
    urls := []string{
        "http://baike.baidu.com/view/1628025.htm?fromtitle=http&fromid=243074&type=syn",
        "http://baike.baidu.com/view/383720.htm?fromtitle=html&fromid=97049&type=syn",
    }
    var reqs []*request.Request
    for _, url := range urls {
        req := request.NewRequest(url, "html", "", "GET", "", nil, nil, nil, nil)
        reqs = append(reqs, req)
    }

    // 内部会针对当前 URLs 运行一次 spider
    pageItemsArr := sp.SetRCNum(2).GetAllByRequest(reqs)
    //pageItemsArr := sp.SetRCNum(2).GetAll(urls, "html")
    for _, item := range pageItemsArr {
        url = item.GetRequest().GetUrl()
        println("url\t:\t" + url)
        fmt.Printf("item\t:\t%s\n", item.GetAll())
    }
}
