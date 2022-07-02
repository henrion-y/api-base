package baiduCensor

import (
	"encoding/json"

	"github.com/henrion-y/api-base/infra/censor"
)

type CensorTextResult struct {
	Conclusion string `json:"conclusion"`
	LogID      int64  `json:"log_id"`
	Data       []struct {
		Msg        string `json:"msg"`
		Conclusion string `json:"conclusion"`
		Hits       []struct {
			WordHitPositions []struct {
				Positions [][]int `json:"positions"`
				Label     string  `json:"label"`
				Keyword   string  `json:"keyword"`
			} `json:"wordHitPositions"`
			Probability       float64     `json:"probability"`
			DatasetName       string      `json:"datasetName"`
			Words             []string    `json:"words"`
			ModelHitPositions [][]float64 `json:"modelHitPositions"`
		} `json:"hits"`
		SubType        int `json:"subType"`
		ConclusionType int `json:"conclusionType"`
		Type           int `json:"type"`
	} `json:"data"`
	IsHitMd5       bool `json:"isHitMd5"`
	ConclusionType int  `json:"conclusionType"`
}

func (c *CensorClient) CensorText(text string) (*censor.CensorResult, error) {
	result := &CensorTextResult{}
	data := c.ContentCensorClient.TextCensor(text)
	err := json.Unmarshal([]byte(data), result)
	if err != nil {
		return nil, err
	}

	censorResult := &censor.CensorResult{
		CensorContent: text,
		CensorType:    censor.CensorTypeText,
		Suggestion:    result.Conclusion,
	}

	switch result.Conclusion {
	case "不合规":
		for i := range result.Data {
			for j := range result.Data[i].Hits {
				for k := range result.Data[i].Hits[j].WordHitPositions {
					reviewLabel := censor.ReviewLabel{
						ReviewContent: result.Data[i].Hits[j].WordHitPositions[k].Keyword,
						Label:         result.Data[i].Msg,
						Rate:          result.Data[i].Hits[j].Probability,
					}
					censorResult.ReviewLabel = append(censorResult.ReviewLabel, reviewLabel)
				}
			}
		}
		censorResult.InterceptStatus = true
	case "合规":
		censorResult.InterceptStatus = false
	}

	return censorResult, nil
}
