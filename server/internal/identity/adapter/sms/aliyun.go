package sms

import (
	"context"
	"fmt"
	"strconv"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
)

// AliyunSmsService 通过阿里云号码认证服务发送和校验短信验证码。
// 阿里云服务内部管理验证码的生成、生命周期和安全校验，
// 本 adapter 仅做转发。
type AliyunSmsService struct {
	client       *dypnsapi.Client
	signName     string
	templateCode string
	codeLength   int
	validTime    int
	interval     int
}

func NewAliyunSmsService(accessKeyID, accessKeySecret, signName, templateCode, codeLength, validTime, interval string) (*AliyunSmsService, error) {
	cfg := &openapi.Config{
		AccessKeyId:     &accessKeyID,
		AccessKeySecret: &accessKeySecret,
	}
	cfg.Endpoint = stringPtr("dypnsapi.aliyuncs.com")

	client, err := dypnsapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create aliyun sms client: %w", err)
	}

	cl, _ := strconv.Atoi(codeLength)
	vt, _ := strconv.Atoi(validTime)
	iv, _ := strconv.Atoi(interval)

	if cl == 0 {
		cl = 6
	}
	if vt == 0 {
		vt = 300
	}
	if iv == 0 {
		iv = 60
	}

	return &AliyunSmsService{
		client:       client,
		signName:     signName,
		templateCode: templateCode,
		codeLength:   cl,
		validTime:    vt,
		interval:     iv,
	}, nil
}

// Send 调用阿里云 SendSmsVerifyCode API 发送短信验证码。
func (s *AliyunSmsService) Send(ctx context.Context, phone string) error {
	countryCode := "86"
	req := &dypnsapi.SendSmsVerifyCodeRequest{
		PhoneNumber:   &phone,
		CountryCode:   &countryCode,
		SignName:      &s.signName,
		TemplateCode:  &s.templateCode,
		TemplateParam: stringPtr(`{"code":"##code##","min":"5"}`),
		CodeLength:    int64Ptr(int64(s.codeLength)),
		ValidTime:     int64Ptr(int64(s.validTime)),
		Interval:      int64Ptr(int64(s.interval)),
	}

	resp, err := s.client.SendSmsVerifyCode(req)
	if err != nil {
		return fmt.Errorf("send sms verify code: %w", err)
	}
	if resp.Body == nil {
		return fmt.Errorf("send sms verify code: empty response")
	}
	if resp.Body.Code != nil && *resp.Body.Code != "OK" {
		msg := ""
		if resp.Body.Message != nil {
			msg = *resp.Body.Message
		}
		return fmt.Errorf("send sms verify code: %s - %s", *resp.Body.Code, msg)
	}

	return nil
}

// Verify 调用阿里云 CheckSmsVerifyCode API 校验短信验证码。
func (s *AliyunSmsService) Verify(ctx context.Context, phone, code string) (bool, error) {
	req := &dypnsapi.CheckSmsVerifyCodeRequest{
		PhoneNumber: &phone,
		VerifyCode:  &code,
	}

	resp, err := s.client.CheckSmsVerifyCode(req)
	if err != nil {
		return false, fmt.Errorf("check sms verify code: %w", err)
	}
	if resp.Body == nil {
		return false, fmt.Errorf("check sms verify code: empty response")
	}
	if resp.Body.Code != nil && *resp.Body.Code != "OK" {
		msg := ""
		if resp.Body.Message != nil {
			msg = *resp.Body.Message
		}
		return false, fmt.Errorf("check sms verify code: %s - %s", *resp.Body.Code, msg)
	}

	// Model.VerifyResult: "PASS" = success, "UNKNOWN" = failed
	if resp.Body.Model != nil && resp.Body.Model.VerifyResult != nil {
		return *resp.Body.Model.VerifyResult == "PASS", nil
	}

	return false, nil
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
