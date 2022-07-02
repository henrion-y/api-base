package censor

const (
	CensorTypeText = iota + 1
	CensorTypeImage
	CensorTypeVideo
	CensorTypeVoice
)

type Client interface {
	CensorText(text string) (*CensorResult, error)
	CensorImage(imgUrl string) (*CensorResult, error)
	// CensorVideo(videoUrl string) (*CensorResult, error)
	// CensorVoice(voiceUrl string) (*CensorResult, error)
}

type ReviewLabel struct {
	ReviewContent string  `json:"review_content"` // 审核中疑似内容
	Label         string  `json:"label"`          // 检测分类
	Rate          float64 `json:"rate"`           // 相似度
}

// CensorResult 检测涉政、涉黄、涉恐
type CensorResult struct {
	CensorContent   string        `json:"scan_content"`     // 审核的内容
	CensorType      int32         `json:"scan_type"`        // 审核的类型 1、文字；2、图片; 3、视频; 4、语音
	ReviewLabel     []ReviewLabel `json:"review_label"`     // 疑似内容列表
	Suggestion      string        `json:"suggestion"`       // 建议
	InterceptStatus bool          `json:"intercept_status"` // 是否被拦截
}
