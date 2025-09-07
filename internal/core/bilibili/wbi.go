package bilibili

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WBI签名管理器
type WBIManager struct {
	imgKey     string
	subKey     string
	mixinKey   string
	client     *http.Client
	lastUpdate time.Time
	userAgent  string
	referer    string
}

// NewWBIManager 创建新的WBI管理器
func NewWBIManager(userAgent, referer string) *WBIManager {
	return &WBIManager{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: userAgent,
		referer:   referer,
	}
}

// 混合密钥编码表
var mixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

// UpdateWBIKeys 更新WBI密钥
func (w *WBIManager) UpdateWBIKeys() error {
	// 如果密钥更新时间在1小时内，不需要重新获取
	if time.Since(w.lastUpdate) < time.Hour {
		return nil
	}

	req, err := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", w.userAgent)
	req.Header.Set("Referer", w.referer)

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get nav response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var navResp NavResponse
	if err := json.Unmarshal(body, &navResp); err != nil {
		return fmt.Errorf("failed to unmarshal nav response: %w", err)
	}

	if navResp.Code != 0 && navResp.Code != -101 {
		return fmt.Errorf("nav API returned error: code=%d, message=%s", navResp.Code, navResp.Message)
	}

	// 提取密钥
	w.imgKey = extractKeyFromURL(navResp.Data.WbiImg.ImgURL)
	w.subKey = extractKeyFromURL(navResp.Data.WbiImg.SubURL)

	// 生成混合密钥
	w.mixinKey = w.generateMixinKey()
	w.lastUpdate = time.Now()

	return nil
}

// extractKeyFromURL 从URL中提取密钥
func extractKeyFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	filename := parts[len(parts)-1]
	return strings.TrimSuffix(filename, ".png")
}

// generateMixinKey 生成混合密钥
func (w *WBIManager) generateMixinKey() string {
	rawKey := w.imgKey + w.subKey
	var mixinKey strings.Builder

	for i := 0; i < 32; i++ {
		if i < len(mixinKeyEncTab) && mixinKeyEncTab[i] < len(rawKey) {
			mixinKey.WriteByte(rawKey[mixinKeyEncTab[i]])
		}
	}

	return mixinKey.String()
}

// SignParams 对参数进行WBI签名
func (w *WBIManager) SignParams(params map[string]string) (string, error) {
	// 确保密钥是最新的
	if err := w.UpdateWBIKeys(); err != nil {
		return "", fmt.Errorf("failed to update WBI keys: %w", err)
	}

	// 添加时间戳
	params["wts"] = strconv.FormatInt(time.Now().Unix(), 10)

	// 过滤参数值中的特殊字符
	filteredParams := make(map[string]string)
	for k, v := range params {
		filteredValue := strings.ReplaceAll(v, "!", "")
		filteredValue = strings.ReplaceAll(filteredValue, "'", "")
		filteredValue = strings.ReplaceAll(filteredValue, "(", "")
		filteredValue = strings.ReplaceAll(filteredValue, ")", "")
		filteredValue = strings.ReplaceAll(filteredValue, "*", "")
		filteredParams[k] = filteredValue
	}

	// 按键名排序
	keys := make([]string, 0, len(filteredParams))
	for k := range filteredParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建查询字符串
	var queryParts []string
	for _, k := range keys {
		v := filteredParams[k]
		queryParts = append(queryParts, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	query := strings.Join(queryParts, "&")

	// 计算签名
	toSign := query + w.mixinKey
	hash := md5.Sum([]byte(toSign))
	wRid := fmt.Sprintf("%x", hash)

	// 添加签名参数
	params["w_rid"] = wRid

	return query + "&w_rid=" + wRid, nil
}
