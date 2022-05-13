package meta

import (
	"encoding/json"
	"io/ioutil"
)

type Brand struct {
	Name string `json:"name"`
	Logo string `json:"logo"`

	Copyright string `json:"copyright"`
	License   string `json:"license"`
}

var (
	brandPath            = "/etc/owlet/brand.json"
	AcquireBrandInfoFunc = AcquireBrandInfo
)

func AcquireBrandInfo(path string) (*Brand, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	b := Brand{}
	if err := json.Unmarshal(content, &b); err != nil {
		return nil, err
	}

	return &b, nil
}
