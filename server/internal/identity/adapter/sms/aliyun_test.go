package sms

import (
	"context"
	"os"
	"testing"

	"ego-server/internal/config"
)

func init() {
	config.Load() // loads .env into os.Environ
}

func newTestService(t *testing.T) *AliyunSmsService {
	t.Helper()

	id := os.Getenv("ALIYUN_ACCESS_KEY_ID")
	secret := os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	sign := getEnvDefault("ALIYUN_SMS_SIGN_NAME", "速通互联验证码")
	tpl := getEnvDefault("ALIYUN_SMS_TEMPLATE_CODE", "100001")
	codeLen := getEnvDefault("ALIYUN_SMS_CODE_LENGTH", "6")
	validTime := getEnvDefault("ALIYUN_SMS_VALID_TIME", "300")
	interval := getEnvDefault("ALIYUN_SMS_INTERVAL", "60")

	if id == "" || secret == "" {
		t.Skip("ALIYUN_ACCESS_KEY_ID and ALIYUN_ACCESS_KEY_SECRET required")
	}

	svc, err := NewAliyunSmsService(id, secret, sign, tpl, codeLen, validTime, interval)
	if err != nil {
		t.Fatalf("NewAliyunSmsService: %v", err)
	}
	return svc
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestSendSmsVerifyCode(t *testing.T) {
	svc := newTestService(t)
	phone := os.Getenv("TEST_PHONE_NUMBER")
	if phone == "" {
		t.Skip("TEST_PHONE_NUMBER not set")
	}

	err := svc.Send(context.Background(), phone)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	t.Log("SMS sent successfully")
}

func TestCheckSmsVerifyCode(t *testing.T) {
	svc := newTestService(t)
	phone := os.Getenv("TEST_PHONE_NUMBER")
	code := os.Getenv("TEST_SMS_CODE")
	if phone == "" || code == "" {
		t.Skip("TEST_PHONE_NUMBER and TEST_SMS_CODE required")
	}

	ok, err := svc.Verify(context.Background(), phone, code)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("verification failed: code not accepted")
	}
	t.Log("SMS code verified successfully")
}
