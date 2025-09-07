package utils

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WBI签名相关常量
const (
	mixinKeyEncTab = "fgjklmnopqrstuvwxyzabcdehi6789+/="
	wbiKey         = "7cd084941338484aae1ad9425b84077c4932caff0ff746eab6f01bf08b70ac45"
)

func SignWBI(params url.Values) (string, error) {
	// 添加时间戳
	params.Set("wts", strconv.FormatInt(time.Now().Unix(), 10))

	// 对参数进行排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建查询字符串
	var query strings.Builder
	for i, k := range keys {
		if i > 0 {
			query.WriteByte('&')
		}
		query.WriteString(url.QueryEscape(k))
		query.WriteByte('=')
		query.WriteString(url.QueryEscape(params.Get(k)))
	}

	// 计算签名
	hash := md5.Sum([]byte(query.String() + wbiKey))
	params.Set("w_rid", hex.EncodeToString(hash[:]))

	return params.Encode(), nil
}
