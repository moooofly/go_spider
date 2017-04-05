package downloader

import (
    "bytes"

    "github.com/PuerkitoBio/goquery"
    "github.com/bitly/go-simplejson"
    //    iconv "github.com/djimenez/iconv-go"
    "github.com/moooofly/go_spider/core/common/mlog"
    "github.com/moooofly/go_spider/core/common/page"
    "github.com/moooofly/go_spider/core/common/request"
    "github.com/moooofly/go_spider/core/common/util"
    //    "golang.org/x/text/encoding/simplifiedchinese"
    //    "golang.org/x/text/transform"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    //"fmt"
    "golang.org/x/net/html/charset"
    //    "regexp"
    //    "golang.org/x/net/html"
    "strings"
	"compress/gzip"
)

// The HttpDownloader download page by package net/http.
// The "html" content is contained in dom parser of package goquery.
// The "json" content is saved.
// The "jsonp" content is modified to json.
// The "text" content will save body plain text only.
// The page result is saved in Page.
type HttpDownloader struct {
}

func NewHttpDownloader() *HttpDownloader {
    return &HttpDownloader{}
}

func (this *HttpDownloader) Download(req *request.Request) *page.Page {
    var mtype string
    // 将 request 关联到 Page 和 PageItems 中
    var p = page.NewPage(req)
    // 获取指定的 response type
    mtype = req.GetResponseType()
    switch mtype {
    case "html":
        return this.downloadHtml(p, req)
    case "json":
        fallthrough
    case "jsonp":
        return this.downloadJson(p, req)
    case "text":
        return this.downloadText(p, req)
    default:
        mlog.LogInst().LogError("error request type:" + mtype)
    }
    return p
}

/*
// The acceptableCharset is test for whether Content-Type is UTF-8 or not
func (this *HttpDownloader) acceptableCharset(contentTypes []string) bool {
    // each type is like [text/html; charset=UTF-8]
    // we want the UTF-8 only
    for _, cType := range contentTypes {
        if strings.Index(cType, "UTF-8") != -1 || strings.Index(cType, "utf-8") != -1 {
            return true
        }
    }
    return false
}


// The getCharset used for parsing the header["Content-Type"] string to get charset of the page.
func (this *HttpDownloader) getCharset(header http.Header) string {
    reg, err := regexp.Compile("charset=(.*)$")
    if err != nil {
        mlog.LogInst().LogError(err.Error())
        return ""
    }

    var charset string
    for _, cType := range header["Content-Type"] {
        substrings := reg.FindStringSubmatch(cType)
        if len(substrings) == 2 {
            charset = substrings[1]
        }
    }

    return charset
}




// Use golang.org/x/text/encoding. Get page body and change it to utf-8
func (this *HttpDownloader) changeCharsetEncoding(charset string, sor io.ReadCloser) string {
    ischange := true
    var tr transform.Transformer
    cs := strings.ToLower(charset)
    if cs == "gbk" {
        tr = simplifiedchinese.GBK.NewDecoder()
    } else if cs == "gb18030" {
        tr = simplifiedchinese.GB18030.NewDecoder()
    } else if cs == "hzgb2312" || cs == "gb2312" || cs == "hz-gb2312" {
        tr = simplifiedchinese.HZGB2312.NewDecoder()
    } else {
        ischange = false
    }

    var destReader io.Reader
    if ischange {
        transReader := transform.NewReader(sor, tr)
        destReader = transReader
    } else {
        destReader = sor
    }

    var sorbody []byte
    var err error
    if sorbody, err = ioutil.ReadAll(destReader); err != nil {
        mlog.LogInst().LogError(err.Error())
        return ""
    }
    bodystr := string(sorbody)

    return bodystr
}

// Use go-iconv. Get page body and change it to utf-8

func (this *HttpDownloader) changeCharsetGoIconv(charset string, sor io.ReadCloser) string {
    var err error
    var converter *iconv.Converter
    if charset != "" && strings.ToLower(charset) != "utf-8" && strings.ToLower(charset) != "utf8" {
        converter, err = iconv.NewConverter(charset, "utf-8")
        if err != nil {
            mlog.LogInst().LogError(err.Error())
            return ""
        }
        defer converter.Close()
    }

    var sorbody []byte
    if sorbody, err = ioutil.ReadAll(sor); err != nil {
        mlog.LogInst().LogError(err.Error())
        return ""
    }
    bodystr := string(sorbody)

    var destbody string
    if converter != nil {
        // convert to utf8
        destbody, err = converter.ConvertString(bodystr)
        if err != nil {
            mlog.LogInst().LogError(err.Error())
            return ""
        }
    } else {
        destbody = bodystr
    }
    return destbody
}
*/

// Charset auto determine. Use golang.org/x/net/html/charset. Get page body and change it to utf-8
func (this *HttpDownloader) changeCharsetEncodingAuto(contentTypeStr string, sor io.ReadCloser) string {
    var err error
    // 获取能够将 sor 中的内容转换成 UTF-8 的 io.Reader
    destReader, err := charset.NewReader(sor, contentTypeStr)
    if err != nil {
        mlog.LogInst().LogError(err.Error())
        destReader = sor
    }

    var sorbody []byte
    if sorbody, err = ioutil.ReadAll(destReader); err != nil {
        mlog.LogInst().LogError(err.Error())
        // For gb2312, an error will be returned.
        // Error like: simplifiedchinese: invalid GBK encoding
        // return ""
    }
    //e,name,certain := charset.DetermineEncoding(sorbody,contentTypeStr)
    bodystr := string(sorbody)

    return bodystr
}

func (this *HttpDownloader) changeCharsetEncodingAutoGzipSupport(contentTypeStr string, sor io.ReadCloser) string {
	var err error
    // gzipReader is an io.Reader that can be read to retrieve
    // uncompressed data from a gzip-format compressed file.
	gzipReader, err := gzip.NewReader(sor)
	if err != nil {
		mlog.LogInst().LogError(err.Error())
		return ""
	}
	defer gzipReader.Close()

    // 获取能够将 gzipReader 中的内容转换成 UTF-8 的 io.Reader
	destReader, err := charset.NewReader(gzipReader, contentTypeStr)
	if err != nil {
		mlog.LogInst().LogError(err.Error())
		destReader = sor
	}

	var sorbody []byte
    // NOTE: 上面的 UTF-8 转换的原因在于 ioutil.ReadAll 的接口需要？
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		mlog.LogInst().LogError(err.Error())
		// For gb2312, an error will be returned.
		// Error like: simplifiedchinese: invalid GBK encoding
		// return ""
	}
	//e,name,certain := charset.DetermineEncoding(sorbody,contentTypeStr)
	bodystr := string(sorbody)

	return bodystr
}

// choose http GET/method to download
func connectByHttp(p *page.Page, req *request.Request) (*http.Response, error) {
    // NOTE: 这里有点意思，为何只关注 redirect 功能
    client := &http.Client{
        CheckRedirect: req.GetRedirectFunc(),
    }

    // 构建 HTTP request
    httpReq, err := http.NewRequest(req.GetMethod(), req.GetUrl(), strings.NewReader(req.GetPostdata()))
    if header := req.GetHeader(); header != nil {
        // NOTE: 这里不是应该直接使用 header 进行赋值么？
        httpReq.Header = req.GetHeader()
    }

    if cookies := req.GetCookies(); cookies != nil {
        for i := range cookies {
            httpReq.AddCookie(cookies[i])
        }
    }

    // 发起 HTTP request 获取 HTTP response
    var resp *http.Response
    if resp, err = client.Do(httpReq); err != nil {
        if e, ok := err.(*url.Error); ok && e.Err != nil && e.Err.Error() == "normal" {
            //  normal
        } else {
            mlog.LogInst().LogError(err.Error())
            p.SetStatus(true, err.Error())
            //fmt.Printf("client do error %v \r\n", err)
            return nil, err
        }
    }

    return resp, nil
}

// choose a proxy server to execute http GET/method to download
func connectByHttpProxy(p *page.Page, in_req *request.Request) (*http.Response, error) {
    request, _ := http.NewRequest("GET", in_req.GetUrl(), nil)
    proxy, err := url.Parse(in_req.GetProxyHost())
    if err != nil {
        return nil, err
    }
    client := &http.Client{
        Transport: &http.Transport{
            Proxy: http.ProxyURL(proxy),
        },
    }
    resp, err := client.Do(request)
    if err != nil {
        return nil, err
    }
    return resp, nil
}

// 该函数作为其它 downloadXXX 函数的基础
//
// 1. 发送 HTTP request 获取 HTTP response
// 2. 支持 HTTP Proxy
// 3. 支持 gzip 处理（自动）
// 4. 支持 page 字符集自动识别和转换（转 UTF-8）
func (this *HttpDownloader) downloadFile(p *page.Page, req *request.Request) (*page.Page, string) {
    var err error
    var urlstr string
    if urlstr = req.GetUrl(); len(urlstr) == 0 {
        mlog.LogInst().LogError("url is empty")
        p.SetStatus(true, "url is empty")
        return p, ""
    }

    var resp *http.Response

    // 发送 HTTP request 获取 HTTP response
    // 考虑是否存在 HTTP 代理的情况
    if proxystr := req.GetProxyHost(); len(proxystr) != 0 {
        // 基于 HTTP Proxy 进行下载
        //fmt.Print("HttpProxy Enter ",proxystr,"\n")
        resp, err = connectByHttpProxy(p, req)
    } else {
        // 直接 HTTP 下载
        //fmt.Print("Http Normal Enter \n",proxystr,"\n")
        resp, err = connectByHttp(p, req)
    }

    if err != nil {
        return p, ""
    }

    //b, _ := ioutil.ReadAll(resp.Body)
    //fmt.Printf("Resp body %v \r\n", string(b))

    p.SetRspHeader(resp.Header)
    p.SetCookies(resp.Cookies())

    // 进行 UTF-8 编码转换（考虑是否使用 gzip 的情况）
	var bodyStr string
	if resp.Header.Get("Content-Encoding") == "gzip" {
		bodyStr = this.changeCharsetEncodingAutoGzipSupport(resp.Header.Get("Content-Type"), resp.Body)
	} else {
		bodyStr = this.changeCharsetEncodingAuto(resp.Header.Get("Content-Type"), resp.Body)
	}
    //fmt.Printf("utf-8 body %v \r\n", bodyStr)
    defer resp.Body.Close()
    return p, bodyStr
}

func (this *HttpDownloader) downloadHtml(p *page.Page, req *request.Request) *page.Page {
    var err error
    // destbody 已经是 resp.Body 经过 UTF-8 编码转换后的 string 了
    p, destbody := this.downloadFile(p, req)
    //fmt.Printf("Destbody %v \r\n", destbody)
    if !p.IsSucc() {
        //fmt.Print("Page error \r\n")
        return p
    }

    // 解析获取到的 response body 内容
    bodyReader := bytes.NewReader([]byte(destbody))

    var doc *goquery.Document
    if doc, err = goquery.NewDocumentFromReader(bodyReader); err != nil {
        mlog.LogInst().LogError(err.Error())
        p.SetStatus(true, err.Error())
        return p
    }

    var body string
    if body, err = doc.Html(); err != nil {
        mlog.LogInst().LogError(err.Error())
        p.SetStatus(true, err.Error())
        return p
    }

    p.SetBodyStr(body).SetHtmlParser(doc).SetStatus(false, "")

    return p
}

func (this *HttpDownloader) downloadJson(p *page.Page, req *request.Request) *page.Page {
    var err error
    p, destbody := this.downloadFile(p, req)
    if !p.IsSucc() {
        return p
    }

    var body []byte
    body = []byte(destbody)
    mtype := req.GetResponseType()
    // 关于 jsonp 详见这里：http://www.cnblogs.com/dowinning/archive/2012/04/19/json-jsonp-jquery.html
    if mtype == "jsonp" {
        tmpstr := util.JsonpToJson(destbody)
        body = []byte(tmpstr)
    }

    var r *simplejson.Json
    if r, err = simplejson.NewJson(body); err != nil {
        mlog.LogInst().LogError(string(body) + "\t" + err.Error())
        p.SetStatus(true, err.Error())
        return p
    }

    // json result
    p.SetBodyStr(string(body)).SetJson(r).SetStatus(false, "")

    return p
}

func (this *HttpDownloader) downloadText(p *page.Page, req *request.Request) *page.Page {
    p, destbody := this.downloadFile(p, req)
    if !p.IsSucc() {
        return p
    }

    p.SetBodyStr(destbody).SetStatus(false, "")
    return p
}
