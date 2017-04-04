//
package main

/*
Packages must be imported:
    "core/common/page"
    "core/spider"
Pckages may be imported:
    "core/pipeline": scawler result persistent;
    "github.com/PuerkitoBio/goquery": html dom parser.
*/
import (
    "github.com/PuerkitoBio/goquery"
    "github.com/moooofly/go_spider/core/common/page"
    "github.com/moooofly/go_spider/core/pipeline"
    "github.com/moooofly/go_spider/core/spider"
    "strings"
    "fmt"
)

type MyPageProcesser struct {
}

func NewMyPageProcesser() *MyPageProcesser {
    return &MyPageProcesser{}
}

// Parse html dom here and record the parse result that we want to Page.
// Package goquery (http://godoc.org/github.com/PuerkitoBio/goquery) is used to parse html.
func (this *MyPageProcesser) Process(p *page.Page) {
    if !p.IsSucc() {
        println(p.Errormsg())
        return
    }

    query := p.GetHtmlParser()
    var urls []string
    query.Find("h3[class='repo-list-name'] a").Each(func(i int, s *goquery.Selection) {
        href, _ := s.Attr("href")
        urls = append(urls, "http://github.com/"+href)
    })
    // these urls will be saved and crawed by other coroutines.
    p.AddTargetRequests(urls, "html")

    name := query.Find(".entry-title .author").Text()
    name = strings.Trim(name, " \t\n")
    repository := query.Find(".entry-title .js-current-repository").Text()
    repository = strings.Trim(repository, " \t\n")
    //readme, _ := query.Find("#readme").Html()
    if name == "" {
        p.SetSkip(true)
    }
    // the entity we want to save by Pipeline
    p.AddField("author", name)
    p.AddField("project", repository)
    //p.AddField("readme", readme)
}

func (this *MyPageProcesser) Finish() {
    fmt.Printf("TODO:before end spider \r\n")
}

func main() {
    // Taskname 用于在 Pipeline 中识别爬取到的内容属于哪个 task
    // 针对指定 URL 进行爬取，并指定 response type 为 "html" ，当前支持 "html", "json", "jsonp", "text" 四种
    // NOTE: 如何事先知道 response type 是什么的呢？
    spider.NewSpider(NewMyPageProcesser(), "TaskName").AddUrl("https://github.com/moooofly?tab=repositories", "html").
        AddPipeline(pipeline.NewPipelineConsole()).SetThreadnum(3).Run()
}
