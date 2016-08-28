package phantom

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/hu17889/go_spider/core/common/request"
)

//Phantom descripe the  structure of PhantomJS Engine
type Phantom struct {
	PhantomJsFilePath string
	TempJsDir         string
	JsEFilemap        map[string]string //e.g. "GET"=>{JSFILE}
	UA                string            //userAgent
	EN                string            //page encode
	Param                               //browser   params

}

//Param descripe the structure of params of browser
type Param struct {
	method   string
	url      string
	header   http.Header
	cookie   string
	postBody string
	// dialTimeout time.Duration
	// connTimeout time.Duration
	// tryTimes    int
	// retryPause  time.Duration
}

//Response descripe the Return of the phantomJS
type Response struct {
	Body   string
	Cookie string
}

const getjs string = `
var system = require('system');
var page = require('webpage').create();
var url = system.args[1];
var cookie = system.args[2];
var pageEncode = system.args[3];
var userAgent = system.args[4];
page.onResourceRequested = function(requestData, request) {
    request.setHeader('Cookie', cookie)
};
phantom.outputEncoding = pageEncode;
page.settings.userAgent = userAgent;
page.open(url, function(status) {
    if (status !== 'success') {
        console.log('Unable to access network');
    } else {
       	var cookie = page.evaluate(function(s) {
            return document.cookie;
        });
        var resp = {
            "Cookie": cookie,
            "Body": page.content
        };
        console.log(JSON.stringify(resp));
    }
    phantom.exit();
});
`
const postjs string = `
var system = require('system');
var page = require('webpage').create();
var url = system.args[1];
var cookie = system.args[2];
var pageEncode = system.args[3];
var userAgent = system.args[4];
var postdata = system.args[5];
page.onResourceRequested = function(requestData, request) {
    request.setHeader('Cookie', cookie)
};
phantom.outputEncoding = pageEncode;
page.settings.userAgent = userAgent;
page.open(url, 'post', postdata, function(status) {
    if (status !== 'success') {
        console.log('Unable to access network');
    } else {
        var cookie = page.evaluate(function(s) {
            return document.cookie;
        });
        var resp = {
            "Cookie": cookie,
            "Body": page.content
        };
        console.log(JSON.stringify(resp));
    }
    phantom.exit();
});
`

//Newphantom Construct the default Phantom
func Newphantom(PhanjsFilePath string, tempjsDir string) *Phantom {

	phantom := new(Phantom)
	phantom.PhantomJsFilePath = PhanjsFilePath
	phantom.TempJsDir = tempjsDir
	phantom.JsEFilemap = make(map[string]string)
	//default settings
	phantom.UA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.84 Safari/537.36"
	phantom.EN = "utf-8"
	//check phantomJs path
	if !filepath.IsAbs(phantom.PhantomJsFilePath) {
		phantom.PhantomJsFilePath, _ = filepath.Abs(phantom.PhantomJsFilePath)
	}
	if !filepath.IsAbs(phantom.TempJsDir) {
		phantom.TempJsDir, _ = filepath.Abs(phantom.TempJsDir)
	}
	// err := os.MkdirAll(phantom.TempJsDir, 0777)
	// if err != nil {
	// 	color.Red("Create file failed........ maybe U need more permissions :)")
	// }
	phantom.createPhantom("get", getjs)
	//	color.Green("[DEBUG]:GET file function call ended")
	phantom.createPhantom("post", postjs)
	//	color.Green("[DEBUG]:POST file function call ended")
	return phantom
}

//DownloadViaphantom start a download task via phantom
func (ph *Phantom) DownloadViaphantom(req *request.Request) (*http.Response, error) {
	ph.method = strings.ToUpper(req.GetMethod())
	ph.url = req.GetUrl()
	ph.header = req.GetHeader()
	ph.postBody = req.GetPostdata()
	for _, v := range req.GetCookies() {
		ph.cookie = v.String()
	}

	var argt []string
	//get request method
	method := ph.method
	//	color.Blue("[DUMP]" + method)
	switch method {
	case "GET":
		argt = []string{
			ph.JsEFilemap["get"],
			ph.url,
			ph.cookie,
			ph.EN,
			ph.UA,
		}
		break
	case "POST":
		argt = []string{
			ph.JsEFilemap["post"],
			ph.url,
			ph.cookie,
			ph.EN,
			ph.UA,
			ph.postBody,
		}
		break
	default:
		break

	}
	//DEBUG
	//	color.Green("[DEBUG]" + "Connect PASS")
	//	color.Blue("[DUMP]" + ph.PhantomJsFilePath)
	// for _, v := range argt {
	// 	color.Blue("[DUMP]" + v)
	// }

	//Response BODY
	cmd := exec.Command(ph.PhantomJsFilePath, argt...)
	//	color.Green("[DEBUG]" + "phantomJS exec start!")
	// resp.Body, err = cmd.StdoutPipe()
	body, err := cmd.StdoutPipe()
	if err != nil {
		// log.Fatalln(err)
		return nil, err
	}
	//debug
	//	color.Green("[DEBUG]" + "phantomJS exec successfully!")
	if err = cmd.Start(); err != nil || body == nil {
		log.Fatalln("empty body....... Retry ?")
		color.Red("[ERR]empty body")
		return nil, err
	}
	var b []byte
	b, err = ioutil.ReadAll(body)
	if err != nil {
		color.Red("[ERR]" + err.Error())
	}
	//RESPONSE is JSON ,So we need Unmarshal()
	phresp := Response{}
	err = json.Unmarshal(b, &phresp)
	if err != nil {
		color.Red("[ERR]" + "站点解析失败了.... 或许你需要PING一下 :)")
	}
	resp := new(http.Response)
	resp.Status = "200 OK"
	resp.StatusCode = 200
	// resp.Header.Set("Set-Cookie", phresp.Cookie)
	resp.Body = ioutil.NopCloser(strings.NewReader(phresp.Body))

	// bd, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	color.Red("[ERR]" + err.Error())
	// 	return nil, err
	// }
	// // fmt.Println(string(body))
	// // color.Blue("[DUMP]" + string(bd))
	// // defer resp.Body.Close()
	return resp, nil
}

func (ph *Phantom) createPhantom(method string, phantomstr string) {
	//phantomJS全路径(从根目录开始索引)
	jsfullpath := filepath.Join(ph.TempJsDir, method)
	file, err := os.Create(jsfullpath)
	if err != nil {
		color.Red("[ERR]" + err.Error())
	}
	if _, err := file.Write([]byte(phantomstr)); err != nil {
		color.Red("[ERR]" + err.Error())
	}
	//close file handle
	defer file.Close()
	ph.JsEFilemap[method] = jsfullpath
}

//SetUserAgent ...
func (ph *Phantom) SetUserAgent(userAgent string) {
	ph.UA = userAgent
}

//SetPageEncode ...
func (ph *Phantom) SetPageEncode(pageEncode string) {
	ph.EN = pageEncode
}
