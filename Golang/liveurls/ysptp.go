package liveurls

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Ysptp struct{}

var cache sync.Map

type CacheItem struct {
	Value      string
	Expiration int64
}

var cctvList = map[string]string{
	"cctv1.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv1.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv2.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv2.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv3.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv3.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv4.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv4.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv5.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv5.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv5p.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv5p.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv6.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv6.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv7.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv7.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv8.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv8.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv9.m3u8":        "http://liveali-tpgq.cctv.cn/live/cctv9.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv10.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv10.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv11.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv11.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv12.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv12.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv13.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv13.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv14.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv14.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv15.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv15.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv16.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv16.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv17.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv17.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cgtnar.m3u8":       "http://liveali-tpgq.cctv.cn/live/cgtnar.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cgtndoc.m3u8":      "http://liveali-tpgq.cctv.cn/live/cgtndoc.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cgtnen.m3u8":       "http://liveali-tpgq.cctv.cn/live/cgtnen.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cgtnfr.m3u8":       "http://liveali-tpgq.cctv.cn/live/cgtnfr.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cgtnru.m3u8":       "http://liveali-tpgq.cctv.cn/live/cgtnru.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cgtnsp.m3u8":       "http://liveali-tpgq.cctv.cn/live/cgtnsp.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv4k.m3u8":       "http://liveali-tpgq.cctv.cn/live/cctv4k.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv4k_10m.m3u8":   "http://liveali-tpgq.cctv.cn/live/cctv4k10m.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv4k16.m3u8":     "http://liveali-tpgq.cctv.cn/live/cctv4k16.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv4k16_10m.m3u8": "http://liveali-tpgq.cctv.cn/live/cctv4k1610m.m3u8,http://liveali-tpgq.cctv.cn/live/",
	"cctv8k_36m.m3u8":   "http://liveali-tp4k.cctv.cn/live/4K36M/playlist.m3u8,http://liveali-tp4k.cctv.cn/live/4K36M/",
	"cctv8k_120m.m3u8":  "http://liveali-tp4k.cctv.cn/live/8K120M/playlist.m3u8,http://liveali-tp4k.cctv.cn/live/8K120M/",
}

func (y *Ysptp) HandleMainRequest(c *gin.Context, id string) {
	uid := c.DefaultQuery("uid", "1234123122")

	if _, ok := cctvList[id]; !ok {
		c.String(http.StatusNotFound, "id not found!")
		return
	}

	urls := strings.Split(cctvList[id], ",")
	data := getURL(id, urls[0], uid, urls[1])
	golang := "http://" + c.Request.Host + c.Request.URL.Path
	re := regexp.MustCompile(`((?i).*?\.ts)`)
	data = re.ReplaceAllString(data, golang+"?ts="+urls[1]+"$1")

	c.Header("Content-Disposition", "attachment;filename="+id)
	c.String(http.StatusOK, data)
}

func (y *Ysptp) HandleTsRequest(c *gin.Context, ts, wsTime string) {
	data := ts + "&wsTime=" + wsTime
	c.Header("Content-Type", "video/MP2T")
	c.String(http.StatusOK, getTs(data))
}

func getURL(id, url, uid, path string) string {
	cacheKey := id + uid
	if playURL, found := getCache(cacheKey); found {
		return fetchData(playURL, path, uid)
	}

	bstrURL := "https://ytpvdn.cctv.cn/cctvmobileinf/rest/cctv/videoliveUrl/getstream"
	postData := `appcommon={"ap":"cctv_app_tv","an":"央视投屏助手","adid":" ` + uid + `","av":"1.1.7"}&url=` + url

	req, err := http.NewRequest("POST", bstrURL, strings.NewReader(postData))
	if err != nil {
		// 处理请求创建错误
		return ""
	}
	req.Header.Set("User-Agent", "cctv_app_tv")
	req.Header.Set("Referer", "api.cctv.cn")
	req.Header.Set("UID", uid)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// 处理请求错误
		return ""
	}
	defer resp.Body.Close()

	var body strings.Builder
	_, err = io.Copy(&body, resp.Body)
	if err != nil {
		// 处理读取响应体错误
		return ""
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(body.String()), &result)
	if err != nil {
		// 处理 JSON 解析错误
		return ""
	}

	playURL, ok := result["url"].(string)
	if !ok || playURL == "" {
		// 处理类型断言或 URL 为空的情况
		return ""
	}

	setCache(cacheKey, playURL)

	return fetchData(playURL, path, uid)
}

func fetchData(playURL, path, uid string) string {
	client := &http.Client{}
	for {
		req, err := http.NewRequest("GET", playURL, nil)
		if err != nil {
			// 处理请求创建错误
			return ""
		}
		req.Header.Set("User-Agent", "cctv_app_tv")
		req.Header.Set("Referer", "api.cctv.cn")
		req.Header.Set("UID", uid)

		resp, err := client.Do(req)
		if err != nil {
			// 处理请求错误
			return ""
		}
		defer resp.Body.Close()

		var body strings.Builder
		_, err = io.Copy(&body, resp.Body)
		if err != nil {
			// 处理读取响应体错误
			return ""
		}

		data := body.String()
		re := regexp.MustCompile(`(.*\.m3u8\?.*)`)
		matches := re.FindStringSubmatch(data)
		if len(matches) > 0 {
			playURL = path + matches[0]
		} else {
			return data
		}
	}
}

func getTs(url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// 处理请求创建错误
		return ""
	}
	req.Header.Set("User-Agent", "cctv_app_tv")
	req.Header.Set("Referer", "https://api.cctv.cn/")
	req.Header.Set("UID", "1234123122")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-encoding", "gzip, deflate")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// 处理请求错误
		return ""
	}
	defer resp.Body.Close()

	var body strings.Builder
	_, err = io.Copy(&body, resp.Body)
	if err != nil {
		// 处理读取响应体错误
		return ""
	}

	return body.String()
}

func getCache(key string) (string, bool) {
	if item, found := cache.Load(key); found {
		cacheItem, ok := item.(CacheItem)
		if ok && time.Now().Unix() < cacheItem.Expiration {
			return cacheItem.Value, true
		}
	}
	return "", false
}

func setCache(key, value string) {
	cache.Store(key, CacheItem{
		Value:      value,
		Expiration: time.Now().Unix() + 3600,
	})
}
