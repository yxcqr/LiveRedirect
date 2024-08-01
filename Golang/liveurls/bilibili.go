// Package liveurls
// @Time:2023/02/10 01:03
// @File:bilibili.go
// @SoftWare:Goland
// @Author:feiyang
// @Contact:TG@feiyangdigital

package liveurls

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
)

type BiliBili struct {
	Rid      string
	Line     string
	Quality  string
	Platform string
}

func (b *BiliBili) GetRealRoomID() any {
	var firstmap = make(map[string]any)
	var realroomid string
	apiurl := "https://api.live.bilibili.com/room/v1/Room/room_init?id=" + b.Rid
	client := &http.Client{}
	r, _ := http.NewRequest("GET", apiurl, nil)
	r.Header.Add("user-agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.3 Mobile/15E148 Safari/604.1")
	resp, _ := client.Do(r)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &firstmap)
	if firstmap["msg"] == "直播间不存在" {
		return nil
	}
	if newmap, ok := firstmap["data"].(map[string]any); ok {
		if newmap["live_status"] != float64(1) {
			return nil
		} else {
			if flt, ok := newmap["room_id"].(float64); ok {
				realroomid = fmt.Sprintf("%v", int(flt))
			}
		}

	}
	return realroomid
}

func (b *BiliBili) GetPlayUrl() any {
	var roomid string
	var realurl string
	if str, ok := b.GetRealRoomID().(string); ok {
		roomid = str
	} else {
		return nil
	}
	client := &http.Client{}
	r, _ := http.NewRequest("GET", "https://api.live.bilibili.com/room/v1/Room/playUrl?cid="+roomid+"&platform="+b.Platform+"&otype=json&quality="+b.Quality, nil)
	resp, _ := client.Do(r)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var jsonStr = string(body)
	if gjson.Get(jsonStr, "code").Int() != 0 {
		return nil
	}
	durls := gjson.Get(jsonStr, "data.durl").Array()

	for i, durl := range durls {
		switch b.Line {
		case "first":
			if i == 0 {
				realurl = durl.Get("url").String()
			}
		case "second":
			if i == 1 {
				realurl = durl.Get("url").String()
			}
		}
	}
	return realurl
}
