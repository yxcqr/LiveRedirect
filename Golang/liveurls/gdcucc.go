package liveurls

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Gdcucc struct{}

type GdcuccCacheItem struct {
	Value      string
	Expiration int64
}

var gdcuccCache sync.Map

func generateRandomMACAddress() string {
	mac := make([]string, 6)
	for i := 0; i < 6; i++ {
		mac[i] = fmt.Sprintf("%02X", rand.Intn(256))
	}
	return strings.Join(mac, ":")
}

func getGdcuccCache(key string) (string, bool) {
	if item, found := gdcuccCache.Load(key); found {
		cacheItem, ok := item.(GdcuccCacheItem)
		if ok && time.Now().Unix() < cacheItem.Expiration {
			return cacheItem.Value, true
		}
	}
	return "", false
}

func setGdcuccCache(key, value string) {
	gdcuccCache.Store(key, GdcuccCacheItem{
		Value:      value,
		Expiration: time.Now().Unix() + 3600, // 缓存固定1小时
	})
}

func getFinalM3U8(id string) (string, error) {
	// 从缓存中读取
	if cachedUrl, found := getGdcuccCache(id); found {
		return cachedUrl, nil
	}

	mac := generateRandomMACAddress()
	api := fmt.Sprintf("http://gdcucc-livod.dispatcher.gitv.tv/gitv_live/%s/%s.m3u8?p=GITV&area=GD_CUCC&partnerCode=GD_CUCC&token=&version=0.0.0.0&apkVersion=4.2.33&mac=%s", id, id, mac)

	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "okhttp/3.8.1")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var builder strings.Builder
	if _, err := io.Copy(&builder, resp.Body); err != nil {
		return "", err
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(builder.String()), &result); err != nil {
		return "", err
	}

	// 获取 m3u8 URL
	dataArray, ok := result["data"].([]any)
	if !ok || len(dataArray) == 0 {
		return "", fmt.Errorf("no data available")
	}

	data := dataArray[0].(map[string]any)
	m3u8url, ok := data["url"].(string)
	if !ok {
		return "", fmt.Errorf("m3u8 url not found")
	}

	// 缓存结果
	setGdcuccCache(id, m3u8url)

	return m3u8url, nil
}

func (g *Gdcucc) HandleGdcuccMainRequest(c *gin.Context, rid string) {
	// 处理 rid，去掉 .m3u8 后缀
	id := strings.TrimSuffix(rid, ".m3u8")

	// 获取最终的M3U8 URL
	m3u8url, err := getFinalM3U8(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 继续请求该 m3u8 URL
	req, err := http.NewRequest("GET", m3u8url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("User-Agent", "okhttp/3.8.1")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	var builder strings.Builder
	if _, err := io.Copy(&builder, resp.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	m3u8Content := builder.String()

	// 替换M3U8内容中的&为$
	m3u8Content = strings.ReplaceAll(m3u8Content, "&", "$")

	// 替换TS文件链接
	redirectPrefix := fmt.Sprintf("http://%s%s?ts=", c.Request.Host, c.Request.URL.Path)
	re := regexp.MustCompile(`((?i).*?\.ts)`)
	m3u8Content = re.ReplaceAllStringFunc(m3u8Content, func(match string) string {
		return redirectPrefix + match
	})

	// 输出处理后的M3U8内容
	c.String(http.StatusOK, m3u8Content)
}

func (g *Gdcucc) HandleGdcuccTsRequest(c *gin.Context, ts string) {
	// 将$替换回&
	ts = strings.ReplaceAll(ts, "$", "&")

	req, err := http.NewRequest("GET", ts, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	req.Header.Set("User-Agent", "okhttp/3.8.1")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	var builder strings.Builder
	if _, err := io.Copy(&builder, resp.Body); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Content-Type", "video/MP2T")
	c.String(http.StatusOK, builder.String())
}
