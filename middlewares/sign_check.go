package middlewares

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	"os"
	"short_link/libraries"
	"short_link/models"
	"sort"
	"strconv"
	"strings"
)

func SignCheck(c *gin.Context) {

	var params map[string]interface{}

	// 拿全部 JSON 参数
	if err := c.ShouldBindBodyWithJSON(&params); err != nil {
		libraries.Err(c, &libraries.Response{Message: "签名不正确，参数缺失。"})
		return
	}

	// 传进来的签名
	key, kOk := params["key"]
	sign, sOk := params["sign"]

	if !kOk || !sOk {
		libraries.Err(c, &libraries.Response{Message: "缺少必要的验签参数"})
		c.Abort()
		return
	}

	// 查询密钥
	secret, appId := (&models.Application{}).GetSecretByKey(key.(string))

	// 应用 id 放入 header
	c.Writer.Header().Set("X-APP-ID", fmt.Sprintf("%d", appId))

	// debug 环境不验签
	isDebug, parErr := strconv.ParseBool(os.Getenv("DEBUG"))
	if parErr == nil && isDebug {
		c.Next()
		return
	}

	// 从全部参数中删除 key / sign
	delete(params, "key")
	delete(params, "sign")

	// 升序排
	paramKeys := make([]string, 0, len(params))
	for k, v := range params {
		query := strings.Join([]string{k, v.(string)}, "=")
		paramKeys = append(paramKeys, query)
	}
	sort.Strings(paramKeys)

	// 拼接
	queryStr := strings.Join(paramKeys, "&")

	// SHA256 加密
	sha256Str := libraries.Sha256(queryStr, secret)

	// base64 编码
	base64Str := base64.StdEncoding.EncodeToString(sha256Str)

	// 最后 url encode 得到签名
	signature := url.QueryEscape(base64Str)

	// 验证签名是否一致
	if sign != signature {
		libraries.Err(c, &libraries.Response{Message: "签名验证不通过"})
		c.Abort()
		return
	}

	c.Next()
}
