package bilibili

import (
	"github.com/hazuki-keatsu/bili-audio-parse-api/internal/core/audio"
	"github.com/hazuki-keatsu/bili-audio-parse-api/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// AudioParser B站音频解析器
type AudioParser struct {
	client     *http.Client
	wbiManager *WBIManager
	downloader *audio.Downloader
	userAgent  string
	referer    string
}

// NewAudioParser 创建音频解析器
func NewAudioParser(userAgent, referer, cacheDir string) *AudioParser {
	return &AudioParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		wbiManager: NewWBIManager(userAgent, referer),
		downloader: audio.NewDownloader(cacheDir, userAgent, referer),
		userAgent:  userAgent,
		referer:    referer,
	}
}

// ParseAudio 解析音频资源
func (p *AudioParser) ParseAudio(bvid string, quality int) (*models.AudioInfo, error) {
	// 1. 获取视频信息
	videoInfo, err := p.getVideoInfo(bvid)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// 2. 获取播放地址
	playURL, err := p.getPlayURL(videoInfo.Data.CID, bvid, quality)
	if err != nil {
		return nil, fmt.Errorf("failed to get play URL: %w", err)
	}

	// 3. 解析DASH音频
	dashInfo, err := p.extractAudioFromDASH(playURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract audio from DASH: %w", err)
	}

	// 4. 下载并转换音频
	audioInfo, err := p.downloader.DownloadAndConvert(
		bvid,
		quality,
		dashInfo.OriginalURL, // 使用原始B站URL
		dashInfo.Bitrate,
		dashInfo.Duration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to download and convert audio: %w", err)
	}

	return audioInfo, nil
}

// getVideoInfo 获取视频信息
func (p *AudioParser) getVideoInfo(bvid string) (*VideoInfoResponse, error) {
	params := map[string]string{
		"bvid": bvid,
	}

	query, err := p.wbiManager.SignParams(params)
	if err != nil {
		return nil, fmt.Errorf("failed to sign params: %w", err)
	}

	url := "https://api.bilibili.com/x/web-interface/view?" + query

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Referer", p.referer)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var videoInfo VideoInfoResponse
	if err := json.Unmarshal(body, &videoInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal video info response: %w", err)
	}

	if videoInfo.Code != 0 {
		return nil, fmt.Errorf("video info API returned error: code=%d, message=%s", videoInfo.Code, videoInfo.Message)
	}

	return &videoInfo, nil
}

// getPlayURL 获取播放地址
func (p *AudioParser) getPlayURL(cid int64, bvid string, quality int) (*PlayURLResponse, error) {
	params := map[string]string{
		"bvid":  bvid,
		"cid":   strconv.FormatInt(cid, 10),
		"fnval": "4048", // 获取所有DASH格式
		"fnver": "0",
		"fourk": "1",
	}

	if quality > 0 {
		params["qn"] = strconv.Itoa(quality)
	}

	query, err := p.wbiManager.SignParams(params)
	if err != nil {
		return nil, fmt.Errorf("failed to sign params: %w", err)
	}

	url := "https://api.bilibili.com/x/player/wbi/playurl?" + query

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Referer", p.referer)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get playurl response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var playURL PlayURLResponse
	if err := json.Unmarshal(body, &playURL); err != nil {
		return nil, fmt.Errorf("failed to unmarshal playurl response: %w", err)
	}

	if playURL.Code != 0 {
		return nil, fmt.Errorf("playurl API returned error: code=%d, message=%s", playURL.Code, playURL.Message)
	}

	return &playURL, nil
}

// extractAudioFromDASH 从DASH中提取音频信息
func (p *AudioParser) extractAudioFromDASH(playURL *PlayURLResponse) (*models.AudioInfo, error) {
	if playURL.Data.Dash == nil || len(playURL.Data.Dash.Audio) == 0 {
		return nil, fmt.Errorf("no audio streams found")
	}

	// 选择最高质量的音频流
	var bestAudio *struct {
		ID        int      `json:"id"`
		BaseURL   string   `json:"baseUrl"`
		BackupURL []string `json:"backupUrl"`
		Bandwidth int      `json:"bandwidth"`
		MimeType  string   `json:"mimeType"`
		Codecs    string   `json:"codecs"`
		Width     int      `json:"width,omitempty"`
		Height    int      `json:"height,omitempty"`
		FrameRate string   `json:"frameRate,omitempty"`
	}

	for i := range playURL.Data.Dash.Audio {
		audio := &playURL.Data.Dash.Audio[i]
		if bestAudio == nil || audio.Bandwidth > bestAudio.Bandwidth {
			bestAudio = audio
		}
	}

	if bestAudio == nil {
		return nil, fmt.Errorf("no suitable audio stream found")
	}

	// 计算比特率 (从带宽估算)
	bitrate := bestAudio.Bandwidth / 1000 // 转换为kbps

	audioInfo := &models.AudioInfo{
		URL:         "", // 将在下载转换后设置
		OriginalURL: bestAudio.BaseURL,
		Format:      "m4s", // 原始格式
		Bitrate:     bitrate,
		Duration:    playURL.Data.Dash.Duration,
		Quality:     bestAudio.ID,
		Size:        0,  // 将在下载后设置
		FileName:    "", // 将在下载后设置
	}

	return audioInfo, nil
}
