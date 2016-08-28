package downloader

import (
	"bytes"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/bitly/go-simplejson"
	"github.com/fatih/color"
	//    iconv "github.com/djimenez/iconv-go"
	"github.com/hu17889/go_spider/core/common/mlog"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/common/request"
	"github.com/hu17889/go_spider/core/common/util"
	"github.com/hu17889/go_spider/core/downloader/phantom"
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
	"compress/gzip"
	"strings"
)

// HTTPDownloader download page by package net/http.
// The "html" content is contained in dom parser of package goquery.
// The "json" content is saved.
// The "jsonp" content is modified to json.
// The "text" content will save body plain text only.
// The page result is saved in Page.
type HTTPDownloader struct {
}

var (
	tempJsDir     = os.Getenv("GOPATH") + `/src/github.com/hu17889/go_spider/core/downloader/phantom/`
	phantomjsFile = os.Getenv("GOPATH") + `/src/github.com/hu17889/go_spider/core/downloader/phantom/phantomjs/bin/phantomjs`
)

//NewHTTPDownloader  Construction
func NewHTTPDownloader() *HTTPDownloader {
	return &HTTPDownloader{}
}

//Download implement the Download method
func (loader *HTTPDownloader) Download(req *request.Request) *page.Page {

	var mtype string

	//return pages
	var p = page.NewPage(req)
	mtype = req.GetResponceType()
	switch mtype {
	case "html":
		return loader.downloadHTML(p, req)
	case "json":
		fallthrough
	case "jsonp":
		return loader.downloadJSON(p, req)
	case "text":
		return loader.downloadText(p, req)
	default:
		mlog.LogInst().LogError("error request type:" + mtype)
	}
	return p
}

/*
// The acceptableCharset is test for whether Content-Type is UTF-8 or not
func (this *HTTPDownloader) acceptableCharset(contentTypes []string) bool {
    // each type is like [text/html; charset=UTF-8]
    // we want the UTF-8 only
    for _, cType := range contentTypes {
        if strings.Index(cType, "UTF-8") != -1 || strings.Index(cType, "utf-8") != -1 {
            return true
        }
    }
    return false
}

*/
/*
//getCharset used for parsing the header["Content-Type"] string to get charset of the page.
func (this *HTTPDownloader) getCharset(header http.Header) string {
	reg, err := regexp.Compile("charset=(.*)$")
	if err != nil {
		mlog.LogInst().LogError(err.Error())
		return ""
	}

	var charset string ="utf-8"
	for _, cType := range header["Content-Type"] {
		substrings := reg.FindStringSubmatch(cType)
		if len(substrings) == 2 {
			charset = substrings[1]
		}
	}

	return charset
}
*/
/*
// Use golang.org/x/text/encoding. Get page body and change it to utf-8
func (this *HTTPDownloader) changeCharsetEncoding(charset string, sor io.ReadCloser) string {
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

func (this *HTTPDownloader) changeCharsetGoIconv(charset string, sor io.ReadCloser) string {
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
func (loader *HTTPDownloader) changeCharsetEncodingAuto(contentTypeStr string, sor io.ReadCloser) string {
	var err error
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

func (loader *HTTPDownloader) changeCharsetEncodingAutoGzipSupport(contentTypeStr string, sor io.ReadCloser) string {
	var err error
	gzipReader, err := gzip.NewReader(sor)
	if err != nil {
		mlog.LogInst().LogError(err.Error())
		return ""
	}
	defer gzipReader.Close()
	destReader, err := charset.NewReader(gzipReader, contentTypeStr)

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

// choose http GET/method to download
func connectByHTTP(p *page.Page, req *request.Request) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: req.GetRedirectFunc(),
	}

	httpreq, err := http.NewRequest(req.GetMethod(), req.GetUrl(), strings.NewReader(req.GetPostdata()))
	if header := req.GetHeader(); header != nil {
		httpreq.Header = req.GetHeader()
	}

	if cookies := req.GetCookies(); cookies != nil {
		for i := range cookies {
			httpreq.AddCookie(cookies[i])
		}
	}

	var resp *http.Response
	if resp, err = client.Do(httpreq); err != nil {
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

func connectByHTTPViaPhantom(p *page.Page, req *request.Request) (resp *http.Response, err error) {
	var phan *phantom.Phantom
	phan = phantom.Newphantom(phantomjsFile, tempJsDir)

	resp, err = phan.DownloadViaphantom(req)
	if err != nil {
		p.SetStatus(true, err.Error())
		return nil, err
	}
	// body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))
	//color.Blue("[DUMP]" + string(body))
	return
}

// choose a proxy server to excute http GET/method to download
func connectByHTTPProxy(p *page.Page, inReq *request.Request) (*http.Response, error) {
	request, _ := http.NewRequest("GET", inReq.GetUrl(), nil)
	proxy, err := url.Parse(inReq.GetProxyHost())
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

// Download file and change the charset of page charset.
func (loader *HTTPDownloader) downloadFile(p *page.Page, req *request.Request) (*page.Page, string) {
	var err error
	var urlstr string
	if urlstr = req.GetUrl(); len(urlstr) == 0 {
		mlog.LogInst().LogError("url is empty")
		p.SetStatus(true, "url is empty")
		return p, ""
	}

	var resp *http.Response

	if proxystr := req.GetProxyHost(); len(proxystr) != 0 {
		//using http proxy
		//fmt.Print("HttpProxy Enter ",proxystr,"\n")
		resp, err = connectByHTTPProxy(p, req)
	} else {
		//normal http download
		//fmt.Print("Http Normal Enter \n",proxystr,"\n")
		// resp, err = connectByHTTP(p, req)
		resp, err = connectByHTTPViaPhantom(p, req)
	}
	//	color.Green("[DEBUG]OUT OF THE phantomJS")
	if err != nil {
		color.Red("[ERR]" + err.Error())
		// panic(err)
		return p, ""

	}

	//b, _ := ioutil.ReadAll(resp.Body)
	//fmt.Printf("Resp body %v \r\n", string(b))

	p.SetHeader(resp.Header)
	p.SetCookies(resp.Cookies())

	// get converter to utf-8
	var bodyStr string
	if resp.Header.Get("Content-Encoding") == "gzip" {
		bodyStr = loader.changeCharsetEncodingAutoGzipSupport(resp.Header.Get("Content-Type"), resp.Body)
	} else {
		bodyStr = loader.changeCharsetEncodingAuto(resp.Header.Get("Content-Type"), resp.Body)
	}
	//fmt.Printf("utf-8 body %v \r\n", bodyStr)
	defer resp.Body.Close()
	return p, bodyStr
}

func (loader *HTTPDownloader) downloadHTML(p *page.Page, req *request.Request) *page.Page {
	var err error
	p, destbody := loader.downloadFile(p, req)
	//fmt.Printf("Destbody %v \r\n", destbody)
	if !p.IsSucc() {
		//fmt.Print("Page error \r\n")
		return p
	}
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

func (loader *HTTPDownloader) downloadJSON(p *page.Page, req *request.Request) *page.Page {
	var err error
	p, destbody := loader.downloadFile(p, req)
	if !p.IsSucc() {
		return p
	}

	var body []byte
	body = []byte(destbody)
	mtype := req.GetResponceType()
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

func (loader *HTTPDownloader) downloadText(p *page.Page, req *request.Request) *page.Page {
	p, destbody := loader.downloadFile(p, req)
	if !p.IsSucc() {
		return p
	}

	p.SetBodyStr(destbody).SetStatus(false, "")
	return p
}
