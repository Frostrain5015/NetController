package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// --- Geo-location ---

type geoEntry struct {
	keywords []string
	country  string
	lng      float64
	lat      float64
}

var geoKeywords = []geoEntry{
	{[]string{"hong kong", "hongkong", "hk", "香港"}, "HK", 114.17, 22.30},
	{[]string{"tokyo", "japan", "jp", "日本", "东京"}, "JP", 139.69, 35.69},
	{[]string{"singapore", "sg", "新加坡"}, "SG", 103.82, 1.35},
	{[]string{"los angeles", "la", "san jose", "seattle", "fremont", "硅谷"}, "US", -118.24, 34.05},
	{[]string{"new york", "ny", "纽约"}, "US", -74.01, 40.71},
	{[]string{"usa", "united states", "美国", "us"}, "US", -122.42, 37.77},
	{[]string{"taipei", "taiwan", "台湾", "台北", "tw"}, "TW", 121.56, 25.03},
	{[]string{"seoul", "korea", "韩国", "首尔", "kr"}, "KR", 126.98, 37.57},
	{[]string{"frankfurt", "germany", "德国", "法兰克福", "de"}, "DE", 8.68, 50.11},
	{[]string{"london", "uk", "england", "united kingdom", "英国", "伦敦", "gb"}, "GB", -0.13, 51.51},
	{[]string{"paris", "france", "法国", "巴黎", "fr"}, "FR", 2.35, 48.86},
	{[]string{"mumbai", "india", "印度", "孟买", "in"}, "IN", 72.88, 19.08},
	{[]string{"toronto", "vancouver", "canada", "加拿大", "多伦多", "温哥华", "ca"}, "CA", -79.38, 43.65},
	{[]string{"sydney", "australia", "澳大利亚", "悉尼", "au"}, "AU", 151.21, -33.87},
	{[]string{"amsterdam", "netherlands", "荷兰", "阿姆斯特丹", "nl"}, "NL", 4.90, 52.37},
	{[]string{"moscow", "russia", "俄罗斯", "莫斯科", "ru"}, "RU", 37.62, 55.76},
	{[]string{"sao paulo", "brazil", "巴西", "圣保罗", "br"}, "BR", -46.63, -23.55},
	{[]string{"vietnam", "越南", "胡志明", "vn"}, "VN", 106.63, 10.82},
	{[]string{"bangkok", "thailand", "泰国", "曼谷", "th"}, "TH", 100.50, 13.76},
	{[]string{"manila", "philippines", "菲律宾", "马尼拉", "ph"}, "PH", 120.98, 14.60},
	{[]string{"kuala lumpur", "malaysia", "马来西亚", "吉隆坡", "my"}, "MY", 101.69, 3.14},
	{[]string{"jakarta", "indonesia", "印尼", "印度尼西亚", "id"}, "ID", 106.85, -6.21},
	{[]string{"istanbul", "turkey", "土耳其", "tr"}, "TR", 28.98, 41.01},
	{[]string{"dubai", "uae", "阿联酋", "迪拜", "ae"}, "AE", 55.27, 25.20},
	{[]string{"south africa", "南非", "za"}, "ZA", 28.05, -26.20},
	{[]string{"argentina", "阿根廷", "ar"}, "AR", -58.38, -34.60},
	{[]string{"milan", "italy", "意大利", "米兰", "it"}, "IT", 9.19, 45.46},
	{[]string{"madrid", "spain", "西班牙", "马德里", "es"}, "ES", -3.70, 40.42},
	{[]string{"warsaw", "poland", "波兰", "pl"}, "PL", 21.01, 52.23},
	{[]string{"stockholm", "sweden", "瑞典", "se"}, "SE", 18.07, 59.33},
	{[]string{"zurich", "switzerland", "瑞士", "ch"}, "CH", 8.54, 47.38},
	{[]string{"helsinki", "finland", "芬兰", "fi"}, "FI", 24.94, 60.17},
}

var countryNames = map[string]string{
	"HK": "香港", "JP": "日本", "SG": "新加坡", "US": "美国", "TW": "台湾", "KR": "韩国",
	"DE": "德国", "GB": "英国", "FR": "法国", "IN": "印度", "CA": "加拿大", "AU": "澳大利亚",
	"NL": "荷兰", "RU": "俄罗斯", "BR": "巴西", "VN": "越南", "TH": "泰国", "PH": "菲律宾",
	"MY": "马来西亚", "ID": "印尼", "TR": "土耳其", "AE": "阿联酋", "ZA": "南非", "AR": "阿根廷",
	"IT": "意大利", "ES": "西班牙", "PL": "波兰", "SE": "瑞典", "CH": "瑞士", "FI": "芬兰",
}

var leafTypes = map[string]bool{
	"ss": true, "vmess": true, "vless": true, "trojan": true, "shadowsocks": true,
	"socks5": true, "http": true, "snell": true, "hysteria": true, "hysteria2": true,
	"tuic": true, "wireguard": true, "ssh": true,
}

var groupTypes = map[string]bool{
	"selector": true, "urltest": true, "fallback": true, "loadbalance": true, "relay": true,
}

func geolocateProxyNode(name string) (lat, lng float64, country string, ok bool) {
	lower := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '-' || r == '_' || r == ' ' || r == '｜' || r == '|' {
			return ' '
		}
		return r
	}, strings.ToLower(name))
	lower = strings.Join(strings.Fields(lower), " ")
	for _, entry := range geoKeywords {
		for _, kw := range entry.keywords {
			if strings.Contains(lower, kw) {
				return entry.lat, entry.lng, entry.country, true
			}
		}
	}
	return 0, 0, "", false
}

// --- Clash API types ---

type clashProxy struct {
	Type string   `json:"type"`
	All  []string `json:"all"`
}

type clashResponse struct {
	Proxies map[string]clashProxy `json:"proxies"`
}

type clashDelayResponse struct {
	Delay int `json:"delay"`
}

type proxyNodeInfo struct {
	Name        string
	DisplayName string
	Group       string
	Type        string
	Country     string
	Location    []float64
}

// --- Clash API client ---

func fetchClashProxies(apiPort int) ([]proxyNodeInfo, float64, string) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/proxies", apiPort))
	if err != nil {
		return nil, 0, ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data clashResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, 0, ""
	}

	seed := make(map[string]*proxyNodeInfo)
	var trafficGB float64
	var expiry string

	for key, obj := range data.Proxies {
		t := strings.ToLower(obj.Type)
		if groupTypes[t] {
			for _, name := range obj.All {
				if _, exists := seed[name]; !exists {
					seed[name] = &proxyNodeInfo{Name: name, Group: key}
				}
				if trafficGB == 0 {
					trafficGB = parseTrafficRemaining(name)
				}
				if expiry == "" {
					expiry = parseExpiry(name)
				}
			}
		}
		if _, exists := seed[key]; !exists && !groupTypes[t] {
			seed[key] = &proxyNodeInfo{Name: key, Type: t}
		} else if info, exists := seed[key]; exists && info.Type == "" {
			info.Type = t
		}
		// Also try to parse from top-level keys
		if trafficGB == 0 {
			trafficGB = parseTrafficRemaining(key)
		}
		if expiry == "" {
			expiry = parseExpiry(key)
		}
	}

	var nodes []proxyNodeInfo
	for _, info := range seed {
		if !leafTypes[info.Type] {
			continue
		}
		lat, lng, country, ok := geolocateProxyNode(info.Name)
		if !ok {
			continue
		}
		info.Country = country
		info.Location = []float64{lng, lat}
		if dn, ok2 := countryNames[country]; ok2 {
			info.DisplayName = dn
		} else {
			info.DisplayName = info.Name
		}
		nodes = append(nodes, *info)
	}
	return nodes, trafficGB, expiry
}

func probeSingleDelay(apiPort int, name string, testURL string, timeoutSec int) int {
	encoded := url.PathEscape(name)
	u := fmt.Sprintf("http://127.0.0.1:%d/proxies/%s/delay?url=%s&timeout=%d",
		apiPort, encoded, url.QueryEscape(testURL), timeoutSec*1000)

	client := &http.Client{Timeout: time.Duration(timeoutSec+3) * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		log.Printf("    delay HTTP err: %v", err)
		return 0
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("    delay resp: status=%d body=%s", resp.StatusCode, string(body)[:min(100, len(body))])

	var dr clashDelayResponse
	if err := json.Unmarshal(body, &dr); err != nil {
		log.Printf("    delay parse err: %v", err)
		return 0
	}
	return dr.Delay
}

// --- Traffic & expiry parsing ---

var trafficRe = regexp.MustCompile(`剩余流量[：:]\s*([\d.]+)\s*GB`)
var expiryRe = regexp.MustCompile(`套餐到期[：:]\s*(\d{4}-\d{2}-\d{2})`)

func parseTrafficRemaining(name string) float64 {
	m := trafficRe.FindStringSubmatch(name)
	if len(m) >= 2 {
		v, _ := strconv.ParseFloat(m[1], 64)
		return v
	}
	return 0
}

func parseExpiry(name string) string {
	m := expiryRe.FindStringSubmatch(name)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
