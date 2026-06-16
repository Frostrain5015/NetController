package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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
	Now  string   `json:"now"`
}

type clashResponse struct {
	Proxies map[string]clashProxy `json:"proxies"`
}

type clashDelayResponse struct {
	Delay int `json:"delay"`
}

type clashConnectionsResponse struct {
	Connections []json.RawMessage `json:"connections"`
}

type proxyNodeInfo struct {
	Name        string
	DisplayName string
	Group       string
	GroupType   string
	Type        string
	Country     string
	Location    []float64
	Selected    bool
}

type subscriptionUsage struct {
	RemainingGB *float64
	UsedGB      *float64
	TotalGB     *float64
	Expiry      string
}

// --- Clash API client ---

func fetchClashProxies(apiPort int) ([]proxyNodeInfo, subscriptionUsage) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/proxies", apiPort))
	if err != nil {
		return nil, subscriptionUsage{}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data clashResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, subscriptionUsage{}
	}

	seed := make(map[string]*proxyNodeInfo)
	textUsage := subscriptionUsage{}

	for key, obj := range data.Proxies {
		t := strings.ToLower(obj.Type)
		if groupTypes[t] {
			for _, name := range obj.All {
				if info, exists := seed[name]; !exists {
					seed[name] = &proxyNodeInfo{Name: name, Group: key, GroupType: t, Selected: obj.Now == name}
				} else {
					if preferGroup(info.GroupType, t) {
						info.Group = key
						info.GroupType = t
					}
					info.Selected = info.Selected || obj.Now == name
				}
				textUsage.mergeText(name)
			}
		}
		if _, exists := seed[key]; !exists && !groupTypes[t] {
			seed[key] = &proxyNodeInfo{Name: key, Type: t}
		} else if info, exists := seed[key]; exists && info.Type == "" {
			info.Type = t
		}
		textUsage.mergeText(key)
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
	usage := fetchProviderSubscriptionUsage(apiPort)
	usage.fillMissing(textUsage)
	return nodes, usage
}

func preferGroup(current string, next string) bool {
	if current == "" {
		return true
	}
	if current != "selector" && next == "selector" {
		return true
	}
	return false
}

func fetchClashConnectionCount(apiPort int) int {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/connections", apiPort))
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	var data clashConnectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0
	}
	return len(data.Connections)
}

func selectProxyNode(apiPort int, group string, name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("empty proxy node name")
	}
	var errs []string
	groups := []string{}
	if strings.TrimSpace(group) != "" {
		groups = append(groups, group)
	}
	for _, candidate := range findSelectableProxyGroups(apiPort, name) {
		if candidate != group {
			groups = append(groups, candidate)
		}
	}
	if len(groups) == 0 {
		return fmt.Errorf("no selectable mihomo group contains %q", name)
	}
	for _, g := range groups {
		if err := putProxySelection(apiPort, g, name); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v", g, err))
		}
	}
	return fmt.Errorf("switch mihomo proxy failed: %s", strings.Join(errs, "; "))
}

func findSelectableProxyGroups(apiPort int, name string) []string {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/proxies", apiPort))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var data clashResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	var groups []string
	for key, obj := range data.Proxies {
		if strings.ToLower(obj.Type) != "selector" {
			continue
		}
		for _, item := range obj.All {
			if item == name {
				groups = append(groups, key)
				break
			}
		}
	}
	return groups
}

func putProxySelection(apiPort int, group string, name string) error {
	payload, _ := json.Marshal(map[string]string{"name": name})
	u := fmt.Sprintf("http://127.0.0.1:%d/proxies/%s", apiPort, url.PathEscape(group))
	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
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

var trafficRe = regexp.MustCompile(`(?i)(?:剩余流量|流量剩余|traffic\s*left|remaining)[：:\s]*([\d.]+)\s*(TB|GB|MB)`)
var expiryRe = regexp.MustCompile(`(?i)(?:套餐到期|到期时间|expire|expires)[：:\s]*(\d{4}[-/]\d{1,2}[-/]\d{1,2})`)

func (u *subscriptionUsage) fillMissing(other subscriptionUsage) {
	if u.RemainingGB == nil {
		u.RemainingGB = other.RemainingGB
	}
	if u.UsedGB == nil {
		u.UsedGB = other.UsedGB
	}
	if u.TotalGB == nil {
		u.TotalGB = other.TotalGB
	}
	if u.Expiry == "" {
		u.Expiry = other.Expiry
	}
}

func (u *subscriptionUsage) mergeText(text string) {
	if u.RemainingGB == nil {
		u.RemainingGB = parseTrafficRemaining(text)
	}
	if u.Expiry == "" {
		u.Expiry = parseExpiry(text)
	}
}

func parseTrafficRemaining(name string) *float64 {
	m := trafficRe.FindStringSubmatch(name)
	if len(m) >= 3 {
		v, _ := strconv.ParseFloat(m[1], 64)
		return floatPtr(toGB(v, m[2]))
	}
	return nil
}

func parseExpiry(name string) string {
	m := expiryRe.FindStringSubmatch(name)
	if len(m) >= 2 {
		return strings.ReplaceAll(m[1], "/", "-")
	}
	return ""
}

func fetchProviderSubscriptionUsage(apiPort int) subscriptionUsage {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/providers/proxies", apiPort))
	if err != nil {
		return subscriptionUsage{}
	}
	defer resp.Body.Close()
	var root map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&root); err != nil {
		return subscriptionUsage{}
	}
	providers, ok := root["providers"].(map[string]interface{})
	if !ok {
		return subscriptionUsage{}
	}
	usage := subscriptionUsage{}
	for _, rawProvider := range providers {
		provider, ok := rawProvider.(map[string]interface{})
		if !ok {
			continue
		}
		usage.mergeText(stringValue(provider, "name"))
		for _, key := range []string{"subscriptionInfo", "subscription-info", "subInfo", "subscription"} {
			rawInfo, ok := lookupAny(provider, key)
			if !ok {
				continue
			}
			info, ok := rawInfo.(map[string]interface{})
			if !ok {
				continue
			}
			usage.mergeSubscriptionInfo(info)
		}
	}
	return usage
}

func (u *subscriptionUsage) mergeSubscriptionInfo(info map[string]interface{}) {
	upload, hasUpload := lookupNumber(info, "upload")
	download, hasDownload := lookupNumber(info, "download")
	total, hasTotal := lookupNumber(info, "total")
	expire, hasExpire := lookupNumber(info, "expire")

	if hasUpload || hasDownload {
		used := bytesToGB(upload + download)
		u.UsedGB = floatPtr(used)
	}
	if hasTotal && total > 0 {
		totalGB := bytesToGB(total)
		u.TotalGB = floatPtr(totalGB)
		if hasUpload || hasDownload {
			remaining := math.Max(0, totalGB-bytesToGB(upload+download))
			u.RemainingGB = floatPtr(remaining)
		}
	}
	if hasExpire && expire > 0 {
		u.Expiry = time.Unix(int64(expire), 0).Format("2006-01-02")
	}
}

func lookupAny(m map[string]interface{}, key string) (interface{}, bool) {
	want := normalizeKey(key)
	for k, v := range m {
		if normalizeKey(k) == want {
			return v, true
		}
	}
	return nil, false
}

func lookupNumber(m map[string]interface{}, key string) (float64, bool) {
	v, ok := lookupAny(m, key)
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func stringValue(m map[string]interface{}, key string) string {
	v, ok := lookupAny(m, key)
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func normalizeKey(s string) string {
	return strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(s))
}

func bytesToGB(v float64) float64 {
	return v / 1024 / 1024 / 1024
}

func floatPtr(v float64) *float64 {
	return &v
}

func toGB(v float64, unit string) float64 {
	switch strings.ToUpper(unit) {
	case "TB":
		return v * 1024
	case "MB":
		return v / 1024
	default:
		return v
	}
}
