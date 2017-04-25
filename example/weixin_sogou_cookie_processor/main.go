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
    "fmt"

    "github.com/PuerkitoBio/goquery"
    "github.com/moooofly/go_spider/core/common/page"
    "github.com/moooofly/go_spider/core/pipeline"
    "github.com/moooofly/go_spider/core/spider"
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

    query.Find(`div[class="wx-rb bg-blue wx-rb_v1 _item"]`).Each(func(i int, s *goquery.Selection) {
        name := s.Find("div.txt-box > h3").Text()
        href, _ := s.Attr("href")

        fmt.Printf("WeName:%v link:http://http://weixin.sogou.com%v \r\n", name, href)
        // the entity we want to save by Pipeline
        p.AddField("name", name)
        p.AddField("href", href)
    })

    next_page_href, _ := query.Find("#sogou_next").Attr("href")
    if next_page_href == "" {
        p.SetSkip(true)
    } else {
        //fmt.Printf("====> next_page_href: %s\n", next_page_href)
        p.AddTargetRequestWithHeaderFile("http://weixin.sogou.com/weixin"+next_page_href, "html", "weixin.sogou.com.json")
    }
}

func (this *MyPageProcesser) Finish() {
    fmt.Printf("TODO: before end spider \r\n")
}

func main() {
    req_url := "http://weixin.sogou.com/weixin?query=%E4%BA%91%E6%B5%AE&type=1&page=1&ie=utf8"
    spider.NewSpider(NewMyPageProcesser(), "TaskName").AddUrlWithHeaderFile(req_url, "html", "weixin.sogou.com.json").
        AddPipeline(pipeline.NewPipelineConsole()).SetRCNum(3).Run()
}
