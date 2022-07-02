package baiduCensor

import (
	"encoding/json"

	"github.com/henrion-y/api-base/infra/censor"
)

type CensorImageResult struct {
	Conclusion string `json:"conclusion"`
	LogID      int64  `json:"log_id"`
	Data       []struct {
		Msg            string  `json:"msg"`
		Conclusion     string  `json:"conclusion"`
		Probability    float64 `json:"probability"`
		SubType        int     `json:"subType"`
		ConclusionType int     `json:"conclusionType"`
		Type           int     `json:"type"`
	} `json:"data"`
	IsHitMd5       bool `json:"isHitMd5"`
	ConclusionType int  `json:"conclusionType"`
}

func (c *CensorClient) CensorImage(imageUrl string) (*censor.CensorResult, error) {
	result := &CensorImageResult{}
	data := c.ContentCensorClient.ImgCensorUrl(imageUrl, nil)
	err := json.Unmarshal([]byte(data), result)
	if err != nil {
		return nil, err
	}

	censorResult := &censor.CensorResult{
		CensorContent: imageUrl,
		CensorType:    censor.CensorTypeImage,
		Suggestion:    result.Conclusion,
	}

	switch result.Conclusion {
	case "不合规":
		for i := range result.Data {
			reviewLabel := censor.ReviewLabel{
				ReviewContent: result.Data[i].Msg,
				Label:         result.Data[i].Msg,
				Rate:          result.Data[i].Probability,
			}
			censorResult.ReviewLabel = append(censorResult.ReviewLabel, reviewLabel)
		}
		censorResult.InterceptStatus = true
	case "合规":
		censorResult.InterceptStatus = false
	}

	return censorResult, nil
}
