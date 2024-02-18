// Package liveurls
// @Time:2024/01/12 22:00
// @File:huya.go
// @SoftWare:Goland
// @Author:feiyang
// @Contact:TG@feiyangdigital

package liveurls

import (
	"Golang/utils"
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Huya struct {
	Rid     string
	Cdn     string
	CdnType string
}

func getJS() string {
	filePath := "res/huya.js"
	file, _ := os.Open(filePath)
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	jstr := strings.Join(lines, "\n")
	return jstr
}

func parseAntiCode(sStreamName, sFlvAntiCode string) string {
	var jsUtil = &utils.JsUtil{}
	c := strings.Split(sFlvAntiCode, "&")
	n := make(map[string]string)
	for _, str := range c {
		temp := strings.Split(str, "=")
		if len(temp) > 1 && temp[1] != "" {
			n[temp[0]] = temp[1]
		}
	}

	// Randomly generating uid
	uid := int64(1462220000000) + rand.Int63n(1145142333)
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	wsTime := strconv.FormatInt(currentTime/1000, 16)
	seqid := uid + currentTime + 216000000 // 216000000 = 30*1000*60*60 = 30h

	// Generating wsSecret
	fm, _ := n["fm"]
	fmDecoded, _ := url.QueryUnescape(fm)
	fmBase64Decoded, _ := base64.StdEncoding.DecodeString(fmDecoded)
	ctype := n["ctype"]
	if ctype == "" {
		ctype = "huya_live"
	}
	var funcContent []string
	funcContent = append(append(funcContent, getJS()), "Oe")
	oe := jsUtil.JsRun(funcContent, strings.Join([]string{strconv.FormatInt(seqid, 10), ctype, "100"}, "|"))
	r := strings.ReplaceAll(string(fmBase64Decoded), "$0", strconv.FormatInt(uid, 10))
	r = strings.ReplaceAll(r, "$1", sStreamName)
	r = strings.ReplaceAll(r, "$2", fmt.Sprintf("%s", oe))
	r = strings.ReplaceAll(r, "$3", wsTime)
	wsSecret := fmt.Sprintf("%s", jsUtil.JsRun(funcContent, r))

	var sb strings.Builder
	sb.WriteString("wsSecret=" + wsSecret + "&wsTime=" + wsTime +
		"&seqid=" + strconv.FormatInt(seqid, 10) +
		"&ctype=" + ctype +
		"&ver=1&fs=" + n["fs"] +
		"&sphdcdn=" + n["sphdcdn"] +
		"&sphdDC=" + n["sphdDC"] +
		"&sphd=" + n["sphd"] +
		"&exsphd=" + n["exsphd"] +
		"&dMod=mseh-32&sdkPcdn=1_1&u=" + strconv.FormatInt(uid, 10) +
		"&t=100&sv=2401190627&sdk_sid=" + strconv.FormatInt(currentTime, 10) + "&ratio=0")
	return sb.String()
}

func (h *Huya) GetLiveUrl() any {
	liveurl := "https://m.huya.com/" + h.Rid
	client := &http.Client{}
	r, _ := http.NewRequest("GET", liveurl, nil)
	r.Header.Add("user-agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.3 Mobile/15E148 Safari/604.1")
	r.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, _ := client.Do(r)
	defer resp.Body.Close()
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	reg := regexp.MustCompile("<script> window.HNF_GLOBAL_INIT = (.*)</script>")
	matches := reg.FindStringSubmatch(string(result))
	if matches == nil || len(matches) < 2 {
		return nil
	}
	return h.extractInfo(matches[1])
}

func (h *Huya) extractInfo(content string) any {
	parse := gjson.Parse(content)
	streamInfo := parse.Get("roomInfo.tLiveInfo.tLiveStreamInfo.vStreamInfo.value")
	if len(streamInfo.Array()) == 0 {
		return nil
	}
	var cdnSlice []string
	var finalurl string
	streamInfo.ForEach(func(key, value gjson.Result) bool {
		var cdnType = gjson.Get(value.String(), "sCdnType").String()
		cdnSlice = append(cdnSlice, cdnType)
		if cdnType == h.Cdn {
			urlStr := fmt.Sprintf("%s/%s.%s?%s",
				value.Get("sFlvUrl").String(),
				value.Get("sStreamName").String(),
				value.Get("sFlvUrlSuffix").String(),
				parseAntiCode(value.Get("sStreamName").String(), value.Get("sFlvAntiCode").String()))
			finalurl = strings.Replace(urlStr, "http://", "https://", 1)
		}
		return true
	})
	if h.CdnType == "display" {
		return cdnSlice
	}
	return finalurl
}
