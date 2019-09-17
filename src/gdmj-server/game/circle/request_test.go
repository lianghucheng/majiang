package circle

import (
	"encoding/json"
	"testing"
)

func TestDoRequest(t *testing.T) {
	req := NewCircleCreateRedPacketRequest(&RedPacketInfo{
		UserID: 446610,
		Sum:    1.5,
		Desc:   "",
	})
	temp := &struct {
		Code string
		Data string
	}{}
	data := DoRequest(req.GatewayUrl, req.Method, req.Params)
	// {"id":null,"code":"0","model":null,"message":"ok","data":"SUCCESS"}
	err := json.Unmarshal(data, temp)
	t.Log(string(data), err)
}
