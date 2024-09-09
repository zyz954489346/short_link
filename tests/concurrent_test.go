package tests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"short_link/libraries"
	"short_link/models"
	"short_link/routes"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

const MaxConcurrent = 100
const DOMAIN = "http://localhost:8080"
const (
	AppSecret string = "s6NmXR0E8pPd23KT"
	AppKey    string = "p79KKyJTgfG2snUs"
)

// 100个网址
var urls = [MaxConcurrent]string{
	"https://www.baidu.com",
	"https://www.qq.com",
	"https://www.sina.com.cn",
	"https://www.sohu.com",
	"https://www.163.com",
	"https://www.ifeng.com",
	"https://www.toutiao.com",
	"https://www.weibo.com",
	"https://www.zhihu.com",
	"https://www.bilibili.com",
	"https://www.taobao.com",
	"https://www.jd.com",
	"https://www.tmall.com",
	"https://www.pinduoduo.com",
	"https://www.xiaohongshu.com",
	"https://www.douyu.com",
	"https://www.huya.com",
	"https://www.iqiyi.com",
	"https://www.youku.com",
	"https://v.qq.com",
	"https://www.mgtv.com",
	"https://www.kuaishou.com",
	"https://www.meituan.com",
	"https://www.dianping.com",
	"https://www.ctrip.com",
	"https://www.fliggy.com",
	"https://www.qunar.com",
	"https://www.liren.com",
	"https://www.huawei.com",
	"https://www.mi.com",
	"https://www.oppo.com",
	"https://www.vivo.com.cn",
	"https://www.oneplus.com",
	"https://www.zol.com.cn",
	"https://www.pconline.com.cn",
	"https://www.mydrivers.com",
	"https://www.sogou.com",
	"https://www.so.com",
	"https://www.asus.com.cn",
	"https://www.lenovo.com.cn",
	"https://www.thinkpad.com.cn",
	"https://www.hp.com.cn",
	"https://www.dell.com.cn",
	"https://www.ibm.com/cn",
	"https://www.amazon.cn",
	"https://www.ele.me",
	"https://www.alipay.com",
	"https://pay.weixin.qq.com",
	"https://www.icbc.com.cn",
	"https://www.boc.cn",
	"https://www.abchina.com",
	"https://www.ccb.com",
	"https://www.bankcomm.com",
	"https://www.cmbchina.com",
	"https://bank.pingan.com",
	"https://www.cmbc.com.cn",
	"https://bank.ecitic.com",
	"https://www.spdb.com.cn",
	"https://www.cebbank.com",
	"https://www.cib.com.cn",
	"https://www.hxb.com.cn",
	"https://www.cgbchina.com.cn",
	"https://jr.jd.com",
	"https://www.antgroup.com",
	"https://www.eastmoney.com",
	"https://www.10jqka.com.cn",
	"https://www.stockstar.com",
	"https://finance.sina.com.cn",
	"https://www.hexun.com",
	"https://www.xdf.cn",
	"https://www.100tal.com",
	"https://www.xueersi.com",
	"https://www.vipkid.com.cn",
	"https://www.zybang.com",
	"https://www.dedao.cn",
	"https://www.douban.com",
	"https://music.163.com",
	"https://y.qq.com",
	"https://www.kugou.com",
	"https://www.kuwo.cn",
	"https://music.taihe.com",
	"https://www.ximalaya.com",
	"https://www.lizhi.fm",
	"https://www.qingting.fm",
	"https://kg.qq.com",
	"http://www.renren.com",
	"https://www.58.com",
	"https://www.anjuke.com",
	"https://www.lianjia.com",
	"https://www.ke.com",
	"https://www.fang.com",
	"https://www.ganji.com",
	"https://www.liepin.com",
	"https://www.zhaopin.com",
	"https://www.51job.com",
	"https://www.zhipin.com",
	"https://www.elitejob.com.cn",
	"https://mail.163.com",
	"https://mail.qq.com",
	"https://www.chinapost.com.cn",
}

// setup 初始化环境
func setup() *gin.Engine {

	r := gin.Default()
	// 加载 env
	_ = godotenv.Load("../.env")

	// 数据库连接
	models.Conn()

	// redis 连接
	libraries.RedisConn(nil)

	routes.Register(r)
	return r
}

// generateParams 补全签名参数
func generateParams(params map[string]string) map[string]string {
	// 倒排转字符串
	keys := make([]string, 0, len(params))
	for k, v := range params {
		keys = append(keys, strings.Join([]string{k, v}, "="))
	}
	sort.Strings(keys)
	queryStr := strings.Join(keys, "&")

	// SHA256 加密
	sha256Str := libraries.Sha256(queryStr, AppSecret)

	// base64 编码
	base64Str := base64.StdEncoding.EncodeToString(sha256Str)

	// 最后 url encode 得到签名
	params["sign"] = url.QueryEscape(base64Str)
	params["key"] = AppKey

	return params
}

// TestUrlShorten 并发测试 for POST /urls/shorten
func TestUrlShorten(t *testing.T) {
	router := setup()
	ch := make(chan string, MaxConcurrent)
	var wg sync.WaitGroup

	for k := range urls {
		wg.Add(1)
		go func(url string, index int) {
			w := httptest.NewRecorder()

			params := generateParams(map[string]string{"url": url})
			data, _ := json.Marshal(params)
			req, _ := http.NewRequest("POST", DOMAIN+"/urls/shorten", strings.NewReader(string(data)))

			router.ServeHTTP(w, req)

			//assert.Equal(t, 200, w.Code)

			resBody, _ := io.ReadAll(w.Body)
			var response map[string]any
			_ = json.Unmarshal(resBody, &response)

			output := strings.Join([]string{
				strconv.Itoa(index),
				url,
				response["data"].(string),
			}, " -> ")

			ch <- output

			wg.Done()

		}(urls[k], k)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for once := range ch {
		fmt.Println(once)
	}
}
