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

// 自定义页面处理器
type MyPageProcesser struct {
}

func NewMyPageProcesser() *MyPageProcesser {
    return &MyPageProcesser{}
}

// 页面处理
// 1. 解析 HTML DOM 并记录解析结果
// 2. 基于 goquery (http://godoc.org/github.com/PuerkitoBio/goquery) 完成 HTML 解析
// 这里的代码已经无法正常工作，因为 github 的页面代码已发生变化
func (this *MyPageProcesser) Process(p *page.Page) {
    if !p.IsSucc() {
        println(p.Errormsg())
        return
    }

    // 获取包含 HTML 结果的对象
    query := p.GetHtmlParser()
    var urls []string
    // 基于 selector 进行过滤，再对过滤后得到对每个元素执行 func()
    // 将所有超链接拼接后保存到 urls 中
    query.Find("h3[class='repo-list-name'] a").Each(func(i int, s *goquery.Selection) {
        href, _ := s.Attr("href")
        urls = append(urls, "http://github.com/"+href)
    })
    // 此处的 urls 应该理解成由原始 URL 衍生出的其他链接
    // 供后续爬取使用
    p.AddTargetRequests(urls, "html")

    name := query.Find(".entry-title .author").Text()
    name = strings.Trim(name, " \t\n")
    repository := query.Find(".entry-title .js-current-repository").Text()
    repository = strings.Trim(repository, " \t\n")
    //readme, _ := query.Find("#readme").Html()

    // 若没有找到 name 则设置跳过该 pageItem
    if name == "" {
        p.SetSkip(true)
    }
    // the entity we want to save by Pipeline
    p.AddField("author", name)
    p.AddField("project", repository)
    //p.AddField("readme", readme)
}

// 资源处理结束的清理
func (this *MyPageProcesser) Finish() {
    fmt.Printf("TODO:before end spider \r\n")
}

func main() {
    // Taskname 用于在 Pipeline 中识别爬取到的内容属于哪个 task
    // 针对指定 URL 进行爬取，并指定 response type 为 "html" ，当前支持 "html", "json", "jsonp", "text" 四种
    // NOTE: 如何事先知道 response type 是什么的呢？
    // AddUrl => 直接将 request 放入 scheduler 中，此处的 URL 可以当作原始 URL
    spider.NewSpider(NewMyPageProcesser(), "TaskName").AddUrl("https://github.com/moooofly?tab=repositories", "html").
        AddPipeline(pipeline.NewPipelineConsole()).SetRCNum(3).Run()
}
