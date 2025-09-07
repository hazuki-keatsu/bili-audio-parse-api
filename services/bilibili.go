package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hazuki-keatsu/bili-parse-api/utils"
)

type BilibiliService struct {
	client *http.Client
}

type AudioInfo struct {
	URL      string `json:"url"`
	Format   string `json:"format"`
	Bitrate  int    `json:"bitrate"`
	Duration int    `json:"duration"`
	Quality  string `json:"quality"`
}

type BilibiliResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Dash struct {
			Audio []struct {
				ID       int    `json:"id"`
				BaseURL  string `json:"base_url"`
				Bitrate  int    `json:"bandwidth"`
				MimeType string `json:"mime_type"`
				Codecs   string `json:"codecs"`
			} `json:"audio"`
		} `json:"dash"`
		Duration int `json:"duration"`
	} `json:"data"`
}

func NewBilibiliService() *BilibiliService {
	return &BilibiliService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *BilibiliService) ParseAudio(bv string, quality int, cookies string) (*AudioInfo, error) {
	// 首先获取视频基本信息
	cid, err := s.getCID(bv)
	if err != nil {
		return nil, err
	}

	// 构建播放地址请求
	playURL := "https://api.bilibili.com/x/player/wbi/playurl"
	params := url.Values{
		"bvid":  {bv},
		"cid":   {strconv.Itoa(cid)},
		"fnval": {"4048"}, // DASH格式
		"fnver": {"0"},
		"fourk": {"1"},
	}

	// 添加WBI签名
	signedParams, err := utils.SignWBI(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", playURL+"?"+signedParams, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	req.Header.Set("Cookie", cookies)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var biliResp BilibiliResponse
	if err := json.NewDecoder(resp.Body).Decode(&biliResp); err != nil {
		return nil, err
	}

	if biliResp.Code != 0 {
		return nil, fmt.Errorf("bilibili API error: %s", biliResp.Message)
	}

	// 选择合适的音频质量
	if len(biliResp.Data.Dash.Audio) == 0 {
		return nil, fmt.Errorf("no audio stream found")
	}

	audio := s.selectAudioByQuality(biliResp.Data.Dash.Audio, quality)

	return &AudioInfo{
		URL:      audio.BaseURL,
		Format:   s.parseFormat(audio.MimeType),
		Bitrate:  audio.Bitrate,
		Duration: biliResp.Data.Duration,
		Quality:  s.getQualityName(audio.ID),
	}, nil
}

func (s *BilibiliService) getCID(bv string) (int, error) {
	// 获取视频CID的实现
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bv)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			CID int `json:"cid"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if result.Code != 0 {
		return 0, fmt.Errorf("failed to get CID")
	}

	return result.Data.CID, nil
}

func (s *BilibiliService) selectAudioByQuality(audios []struct {
	ID       int    `json:"id"`
	BaseURL  string `json:"base_url"`
	Bitrate  int    `json:"bandwidth"`
	MimeType string `json:"mime_type"`
	Codecs   string `json:"codecs"`
}, quality int) struct {
	ID       int    `json:"id"`
	BaseURL  string `json:"base_url"`
	Bitrate  int    `json:"bandwidth"`
	MimeType string `json:"mime_type"`
	Codecs   string `json:"codecs"`
} {
	// 根据质量参数选择合适的音频流
	// 可以根据ID或bitrate进行选择
	if quality > 0 && quality < len(audios) {
		return audios[quality]
	}
	// 默认返回最高质量
	return audios[0]
}

func (s *BilibiliService) parseFormat(mimeType string) string {
	if strings.Contains(mimeType, "mp4") {
		return "m4s"
	}
	return "unknown"
}

func (s *BilibiliService) getQualityName(id int) string {
	qualityMap := map[int]string{
		30216: "64K",
		30232: "132K",
		30280: "192K",
		30250: "Hi-Res无损",
	}
	if name, ok := qualityMap[id]; ok {
		return name
	}
	return "标准音质"
}
