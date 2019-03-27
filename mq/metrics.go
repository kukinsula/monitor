package mq

import "encoding/json"

type Metrics struct {
	CPU map[string]interface{} `json:"cpu"`
	RAM map[string]interface{} `json:"ram"`
}

func (metrics *Metrics) String() string {
	data, err := json.Marshal(metrics)
	if err != nil {
		return ""
	}

	return string(data)
}
