package bilibili

// WBI签名相关的模型结构

// NavResponse 导航接口响应
type NavResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		WbiImg struct {
			ImgURL string `json:"img_url"`
			SubURL string `json:"sub_url"`
		} `json:"wbi_img"`
	} `json:"data"`
}

// VideoInfoResponse 视频信息响应
type VideoInfoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		BVID  string `json:"bvid"`
		AID   int64  `json:"aid"`
		Title string `json:"title"`
		CID   int64  `json:"cid"`
		Pages []struct {
			CID   int64  `json:"cid"`
			Page  int    `json:"page"`
			From  string `json:"from"`
			Part  string `json:"part"`
			Index int    `json:"index"`
		} `json:"pages"`
	} `json:"data"`
}

// PlayURLResponse 播放地址响应
type PlayURLResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Quality     int    `json:"quality"`
		Format      string `json:"format"`
		Timelength  int    `json:"timelength"`
		VideoCodeid int    `json:"video_codecid"`
		Dash        *struct {
			Duration int `json:"duration"`
			Audio    []struct {
				ID        int      `json:"id"`
				BaseURL   string   `json:"baseUrl"`
				BackupURL []string `json:"backupUrl"`
				Bandwidth int      `json:"bandwidth"`
				MimeType  string   `json:"mimeType"`
				Codecs    string   `json:"codecs"`
				Width     int      `json:"width,omitempty"`
				Height    int      `json:"height,omitempty"`
				FrameRate string   `json:"frameRate,omitempty"`
			} `json:"audio"`
			Video []struct {
				ID        int      `json:"id"`
				BaseURL   string   `json:"baseUrl"`
				BackupURL []string `json:"backupUrl"`
				Bandwidth int      `json:"bandwidth"`
				MimeType  string   `json:"mimeType"`
				Codecs    string   `json:"codecs"`
				Width     int      `json:"width"`
				Height    int      `json:"height"`
				FrameRate string   `json:"frameRate"`
			} `json:"video"`
		} `json:"dash"`
	} `json:"data"`
}
