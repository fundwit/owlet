package meta

import (
	"encoding/json"
	"io/ioutil"

	"github.com/fundwit/go-commons/types"
)

type Build struct {
	Timestamp   types.Timestamp `json:"buildTime"`
	Release     string          `json:"version"`
	SourceCodes []SourceCode    `json:"sourceCodes"`
}

type SourceCode struct {
	Repository string `json:"repository"`
	Ref        string `json:"ref"`

	LastChange CodeReversion `json:"reversion"`
}

type CodeReversion struct {
	ID        string          `json:"id"`
	Timestamp types.Timestamp `json:"timestamp"`
	Author    string          `json:"author"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
}

var (
	buildInfoPath        = "/buildInfo.json"
	AcquireBuildInfoFunc = AcquireBuildInfo
)

func AcquireBuildInfo(path string) (*Build, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	b := Build{}
	if err := json.Unmarshal(content, &b); err != nil {
		return nil, err
	}

	return &b, nil
}
