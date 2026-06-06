package sms

import (
	"context"
	"fmt"
)

// AliyunSmsService 通过阿里云短信服务发送和校验验证码。
// 阿里云短信服务内部管理验证码的生成、生命周期和安全校验，
// 本 adapter 仅做转发。
// TODO: 集成阿里云 SDK
type AliyunSmsService struct {
	// accessKeyID     string
	// accessKeySecret string
	// signName        string
	// templateCode    string
}

func NewAliyunSmsService() *AliyunSmsService {
	return &AliyunSmsService{}
}

// Send 调用阿里云 API 发送短信验证码。
func (s *AliyunSmsService) Send(ctx context.Context, phone string) error {
	// TODO: 集成阿里云 SDK
	// client, _ := dysmsapi.NewClientWithAccessKey("cn-hangzhou", s.accessKeyID, s.accessKeySecret)
	// req := dysmsapi.CreateSendSmsRequest()
	// req.PhoneNumbers = phone
	// req.SignName = s.signName
	// req.TemplateCode = s.templateCode
	// req.TemplateParam = `{"code":"123456"}`
	// _, err := client.SendSms(req)
	// return err

	return fmt.Errorf("aliyun SMS not configured")
}

// Verify 调用阿里云 API 校验验证码。
func (s *AliyunSmsService) Verify(ctx context.Context, phone, code string) (bool, error) {
	// TODO: 集成阿里云 SDK
	// req := dysmsapi.CreateCheckSmsCodeRequest()
	// req.PhoneNumbers = phone
	// req.SmsCode = code
	// res, err := client.CheckSmsCode(req)
	// if err != nil { return false, err }
	// return res.Success, nil

	return false, fmt.Errorf("aliyun SMS not configured")
}
