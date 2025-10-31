package handler

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
)

var (
	request     *Request
	onceRequest sync.Once
)

func GetRequest() *Request {
	if request == nil {
		onceRequest.Do(func() {
			requestId := fmt.Sprintf("bdpan%d", time.Now().UnixMicro())
			envID := os.Getenv("WGO_TEST_REQUEST_ID")
			if envID != "" {
				requestId = envID
			}
			request = &Request{
				ID: requestId,
			}
		})
	}
	return request
}

type Request struct {
	ID        string
	GlobalReq dto.GlobalReq
}

func (r Request) GetConfigPath() string {
	if r.GlobalReq.Config != "" {
		return r.GlobalReq.Config
	} else {
		configPath, err := config.GetDefaultConfigPath()
		if err != nil {
			panic("get config path error: " + err.Error())
		}
		return configPath
	}
}
