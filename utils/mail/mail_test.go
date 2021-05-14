package mail

import (
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"testing"
)

func TestSendMail(t *testing.T) {
	configs.Init("../../config.yml")
	logs.Init("debug",7,"1")

	err := SendMail([]string{"13624015331@163.com"}, "测试", "<h1>测试</h1>")
	if err != nil {
		t.Fatal(err)
	}
}
