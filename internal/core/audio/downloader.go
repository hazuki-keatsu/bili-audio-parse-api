package audio

import (
	"bili-parse-api/internal/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Downloader 音频下载器
type Downloader struct {
	cacheDir  string
	userAgent string
	referer   string
	client    *http.Client
}

// NewDownloader 创建音频下载器
func NewDownloader(cacheDir, userAgent, referer string) *Downloader {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create cache directory %s: %v\n", cacheDir, err)
	}

	return &Downloader{
		cacheDir:  cacheDir,
		userAgent: userAgent,
		referer:   referer,
		client: &http.Client{
			Timeout: 5 * time.Minute, // 增加超时时间以支持大文件下载
		},
	}
}

// DownloadAndConvert 下载音频并转换为MP3
func (d *Downloader) DownloadAndConvert(bvid string, quality int, dashURL string, bitrate int, duration int) (*models.AudioInfo, error) {
	// 生成文件名
	fileName := fmt.Sprintf("%s_%d_%d.mp3", bvid, quality, time.Now().Unix())
	mp3Path := filepath.Join(d.cacheDir, fileName)

	// 检查文件是否已存在
	if _, err := os.Stat(mp3Path); err == nil {
		// 文件已存在，返回现有文件信息
		stat, _ := os.Stat(mp3Path)
		return &models.AudioInfo{
			URL:         "/static/" + fileName,
			OriginalURL: dashURL,
			Format:      "mp3",
			Bitrate:     bitrate,
			Duration:    duration,
			Quality:     quality,
			Size:        stat.Size(),
			FileName:    fileName,
		}, nil
	}

	// 1. 下载m4s文件
	m4sPath := strings.Replace(mp3Path, ".mp3", ".m4s", 1)
	if err := d.downloadFile(dashURL, m4sPath); err != nil {
		return nil, fmt.Errorf("failed to download m4s file: %w", err)
	}

	// 2. 转换为MP3
	if err := d.convertToMP3(m4sPath, mp3Path, bitrate); err != nil {
		// 清理临时文件
		os.Remove(m4sPath)
		return nil, fmt.Errorf("failed to convert to mp3: %w", err)
	}

	// 3. 清理临时m4s文件
	os.Remove(m4sPath)

	// 4. 获取转换后的文件信息
	stat, err := os.Stat(mp3Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get mp3 file info: %w", err)
	}

	return &models.AudioInfo{
		URL:         "/static/" + fileName,
		OriginalURL: dashURL,
		Format:      "mp3",
		Bitrate:     bitrate,
		Duration:    duration,
		Quality:     quality,
		Size:        stat.Size(),
		FileName:    fileName,
	}, nil
}

// downloadFile 下载文件
func (d *Downloader) downloadFile(url, filePath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置必要的请求头以绕过防盗链
	req.Header.Set("User-Agent", d.userAgent)
	req.Header.Set("Referer", d.referer)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 创建文件
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// 复制数据
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	return nil
}

// convertToMP3 使用ffmpeg转换为MP3
func (d *Downloader) convertToMP3(inputPath, outputPath string, bitrate int) error {
	// 检查ffmpeg是否可用
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}

	// 构建ffmpeg命令
	// -i: 输入文件
	// -acodec libmp3lame: 使用MP3编码器
	// -ab: 音频比特率
	// -y: 覆盖输出文件
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-acodec", "libmp3lame",
		"-ab", fmt.Sprintf("%dk", bitrate),
		"-y",
		outputPath,
	)

	// 执行转换
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GetFileSize 获取文件大小
func (d *Downloader) GetFileSize(filePath string) (int64, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// FileExists 检查文件是否存在
func (d *Downloader) FileExists(fileName string) bool {
	filePath := filepath.Join(d.cacheDir, fileName)
	_, err := os.Stat(filePath)
	return err == nil
}

// CleanupOldFiles 清理过期文件
func (d *Downloader) CleanupOldFiles(maxAge time.Duration) error {
	return filepath.Walk(d.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件是否过期
		if time.Since(info.ModTime()) > maxAge {
			if err := os.Remove(path); err != nil {
				fmt.Printf("Warning: failed to remove old file %s: %v\n", path, err)
			}
		}

		return nil
	})
}
