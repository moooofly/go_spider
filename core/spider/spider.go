// craw master module
package spider

import (
    "github.com/moooofly/go_spider/core/common/mlog"
    "github.com/moooofly/go_spider/core/common/page"
    "github.com/moooofly/go_spider/core/common/page_items"
    "github.com/moooofly/go_spider/core/common/request"
    "github.com/moooofly/go_spider/core/common/resource_manage"
    "github.com/moooofly/go_spider/core/downloader"
    "github.com/moooofly/go_spider/core/page_processer"
    "github.com/moooofly/go_spider/core/pipeline"
    "github.com/moooofly/go_spider/core/scheduler"
    "math/rand"
    //"net/http"
    "time"
    //"fmt"
)

type Spider struct {
    taskname string
    pPageProcesser page_processer.PageProcesser
    pDownloader downloader.Downloader
    pScheduler scheduler.Scheduler  // 为当前 spider 指定的 scheduler ，其中存放所有 request
    pPipelines []pipeline.Pipeline  // 每个 pipeline 对应一种输出形式
    rcManager resource_manage.ResourceManager  // 资源管理
    rcNum uint  // 控制可用资源数量（用于限制并发 goroutine 数目）
    exitWhenComplete bool
    // If sleeptype is "fixed", the s is the sleep time and e is useless.
    // If sleeptype is "rand", the sleep time is rand between s and e.
    startSleeptime uint
    endSleeptime   uint
    sleeptype      string  // 用于控制失败重试时休眠时间的选择方式："fixed" or "rand"
}

// Spider is the scheduler module for all the other modules, like downloader, pipeline, scheduler and etc.
// taskname => could be empty string, or it can be used in Pipeline for record the result crawled by which task
func NewSpider(pageinst page_processer.PageProcesser, taskname string) *Spider {
    // 初始化日志输出功能（输出到 os.Stderr 上）
    mlog.StraceInst().Open()

    ap := &Spider{taskname: taskname, pPageProcesser: pageinst}

    // 关闭 filelog 日志输出功能
    ap.CloseFileLog()
    ap.exitWhenComplete = true
    ap.sleeptype = "fixed"
    ap.startSleeptime = 0

    // 指定 queue scheduler 给当前 spider
    // false 表明当前 queue 不去重
    if ap.pScheduler == nil {
        ap.SetScheduler(scheduler.NewQueueScheduler(false))
    }

    // 新建 HTTP 下载器
    if ap.pDownloader == nil {
        ap.SetDownloader(downloader.NewHttpDownloader())
    }

    mlog.StraceInst().Println("** start spider **")
    ap.pPipelines = make([]pipeline.Pipeline, 0)

    return ap
}

func (this *Spider) Taskname() string {
    return this.taskname
}

// Deal with one url and return the PageItems.
func (this *Spider) Get(url string, respType string) *page_items.PageItems {
    req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
    return this.GetByRequest(req)
}

// Deal with several urls and return the PageItems slice.
func (this *Spider) GetAll(urls []string, respType string) []*page_items.PageItems {
    for _, u := range urls {
        req := request.NewRequest(u, respType, "", "GET", "", nil, nil, nil, nil)
        this.AddRequest(req)
    }

    pip := pipeline.NewCollectPipelinePageItems()
    this.AddPipeline(pip)

    this.Run()

    return pip.GetCollected()
}

// Deal with one url and return the PageItems with other setting.
func (this *Spider) GetByRequest(req *request.Request) *page_items.PageItems {
    var reqs []*request.Request
    reqs = append(reqs, req)
    items := this.GetAllByRequest(reqs)
    if len(items) != 0 {
        return items[0]
    }
    return nil
}

// Deal with several urls and return the PageItems slice
func (this *Spider) GetAllByRequest(reqs []*request.Request) []*page_items.PageItems {
    // push url
    for _, req := range reqs {
        //req := request.NewRequest(u, respType, urltag, method, postdata, header, cookies)
        this.AddRequest(req)
    }

    pip := pipeline.NewCollectPipelinePageItems()
    this.AddPipeline(pip)

    this.Run()

    return pip.GetCollected()
}

func (this *Spider) Run() {
    if this.rcNum == 0 {
        this.rcNum = 1
    }
    this.rcManager = resource_manage.NewResourceManageChan(this.rcNum)

    for {
        // 取出一个待处理 request
        req := this.pScheduler.Poll()

        // NOTE: rcManager is not atomic
        // 当满足：
        // 1. 尚有资源被占用（存在未处理完的东东）
        // 2. scheduler 中没有更多的 request 待处理
        // 3. 所有 request 处理结束后退出当前程序
        if this.rcManager.Used() == 0 && req == nil && this.exitWhenComplete {
	        mlog.StraceInst().Println("** executed callback **")
            // 清理工作
	        this.pPageProcesser.Finish()
            mlog.StraceInst().Println("** end spider **")
            break
        } else if req == nil {  // 或者存在资源尚被占用情况，或者设置了完成后不退出
            time.Sleep(500 * time.Millisecond)
            //mlog.StraceInst().Println("scheduler is empty")
            continue
        }
        // 获取一个可用资源（无资源可用会阻塞）
        this.rcManager.GetOne()

        // 通过上述资源管理控制并发 goroutine 数量
        // Asynchronous fetching
        go func(req *request.Request) {
            defer this.rcManager.FreeOne()
            //time.Sleep( time.Duration(rand.Intn(5)) * time.Second)
            mlog.StraceInst().Println("start crawl : " + req.GetUrl())
            // 开始页面处理（爬网页）
            this.pageProcess(req)
        }(req)
    }
    this.close()
}

// spider 状态重置
func (this *Spider) close() {
    this.SetScheduler(scheduler.NewQueueScheduler(false))
    this.SetDownloader(downloader.NewHttpDownloader())
    this.pPipelines = make([]pipeline.Pipeline, 0)
    this.exitWhenComplete = true
}

func (this *Spider) AddPipeline(p pipeline.Pipeline) *Spider {
    this.pPipelines = append(this.pPipelines, p)
    return this
}

func (this *Spider) SetScheduler(s scheduler.Scheduler) *Spider {
    this.pScheduler = s
    return this
}

func (this *Spider) GetScheduler() scheduler.Scheduler {
    return this.pScheduler
}

func (this *Spider) SetDownloader(d downloader.Downloader) *Spider {
    this.pDownloader = d
    return this
}

func (this *Spider) GetDownloader() downloader.Downloader {
    return this.pDownloader
}

func (this *Spider) SetRCNum(i uint) *Spider {
    this.rcNum = i
    return this
}

func (this *Spider) GetRCNum() uint {
    return this.rcNum
}

// If exit when each crawl task is done.
// If you want to keep spider in memory all the time and add url from outside, you can set it true.
func (this *Spider) SetExitWhenComplete(e bool) *Spider {
    this.exitWhenComplete = e
    return this
}

func (this *Spider) GetExitWhenComplete() bool {
    return this.exitWhenComplete
}

// The OpenFileLog initialize the log path and open log.
// If log is opened, error info or other useful info in spider will be logged in file of the filepath.
// Log command is mlog.LogInst().LogError("info") or mlog.LogInst().LogInfo("info").
// Spider's default log is closed.
// The filepath is absolute path.
func (this *Spider) OpenFileLog(filePath string) *Spider {
    mlog.InitFilelog(true, filePath)
    return this
}

// OpenFileLogDefault open file log with default file path like "WD/log/log.2014-9-1".
func (this *Spider) OpenFileLogDefault() *Spider {
    mlog.InitFilelog(true, "")
    return this
}

// The CloseFileLog close file log.
func (this *Spider) CloseFileLog() *Spider {
    mlog.InitFilelog(false, "")
    return this
}

// The OpenStrace open strace that output progress info on the screen.
// Spider's default strace is opened.
func (this *Spider) OpenStrace() *Spider {
    mlog.StraceInst().Open()
    return this
}

// The CloseStrace close strace.
func (this *Spider) CloseStrace() *Spider {
    mlog.StraceInst().Close()
    return this
}

// The SetSleepTime set sleep time after each crawl task.
// The unit is millisecond.
// If sleeptype is "fixed", the s is the sleep time and e is useless.
// If sleeptype is "rand", the sleep time is rand between s and e.
func (this *Spider) SetSleepTime(sleeptype string, s uint, e uint) *Spider {
    this.sleeptype = sleeptype
    this.startSleeptime = s
    this.endSleeptime = e
    if this.sleeptype == "rand" && this.startSleeptime >= this.endSleeptime {
        panic("startSleeptime must smaller than endSleeptime")
    }
    return this
}

func (this *Spider) sleep() {
    if this.sleeptype == "fixed" {
        // s is the sleep time and e is useless
        time.Sleep(time.Duration(this.startSleeptime) * time.Millisecond)
    } else if this.sleeptype == "rand" {
        // sleep time is rand between s and e
        sleeptime := rand.Intn(int(this.endSleeptime-this.startSleeptime)) + int(this.startSleeptime)
        time.Sleep(time.Duration(sleeptime) * time.Millisecond)
    }
}

// url => 目标 URL
// respType => 预期应答类型
func (this *Spider) AddUrl(url string, respType string) *Spider {
    // 构建一个 GET request 结构
    req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
    // 将 req 放入 scheduler
    this.AddRequest(req)
    return this
}

func (this *Spider) AddUrlEx(url string, respType string, headerFile string, proxyHost string) *Spider {
    req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
    this.AddRequest(req.AddHeaderFile(headerFile).AddProxyHost(proxyHost))
    return this
}

func (this *Spider) AddUrlWithHeaderFile(url string, respType string, headerFile string) *Spider {
    req := request.NewRequestWithHeaderFile(url, respType, headerFile)
    this.AddRequest(req)
    return this
}

func (this *Spider) AddUrls(urls []string, respType string) *Spider {
    for _, url := range urls {
        req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
        this.AddRequest(req)
    }
    return this
}

func (this *Spider) AddUrlsWithHeaderFile(urls []string, respType string, headerFile string) *Spider {
    for _, url := range urls {
        req := request.NewRequestWithHeaderFile(url, respType, headerFile)
        this.AddRequest(req)
    }
    return this
}

func (this *Spider) AddUrlsEx(urls []string, respType string, headerFile string, proxyHost string) *Spider {
    for _, url := range urls {
        req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
        this.AddRequest(req.AddHeaderFile(headerFile).AddProxyHost(proxyHost))
    }
    return this
}

// 将 request 添加到 scheduler 中
func (this *Spider) AddRequest(req *request.Request) *Spider {
    if req == nil {
        mlog.LogInst().LogError("request is nil")
        return this
    } else if req.GetUrl() == "" {
        mlog.LogInst().LogError("request is empty")
        return this
    }
    this.pScheduler.Push(req)
    return this
}

//
func (this *Spider) AddRequests(reqs []*request.Request) *Spider {
    for _, req := range reqs {
        this.AddRequest(req)
    }
    return this
}

// 开始页面处理
func (this *Spider) pageProcess(req *request.Request) {
    var p *page.Page

    defer func() {
        if err := recover(); err != nil { // do not affect other
            if strerr, ok := err.(string); ok {
                mlog.LogInst().LogError(strerr)
            } else {
                mlog.LogInst().LogError("pageProcess error")
            }
        }
    }()

    // 默认重试 3 次
    for i := 0; i < 3; i++ {
        this.sleep()
        // 针对初始 URL 进行爬取
        p = this.pDownloader.Download(req)
        if p.IsSucc() { // if fail retry 3 times
            break
        }
    }

    if !p.IsSucc() { // if fail do not need process
        return
    }

    // 调用自定义页面处理代码（从初始页面上获取并构建其它链接）
    this.pPageProcesser.Process(p)
    // 将等待放入 sheduler 的其他 request 加到其中
    for _, req := range p.GetTargetRequests() {
        this.AddRequest(req)
    }

    if !p.GetSkip() {
        for _, pipe := range this.pPipelines {
            //fmt.Println("%v",p.GetPageItems().GetAll())
            // 处理爬取到所有内容（例如输出到 console）
            pipe.Process(p.GetPageItems(), this)
        }
    }
}
