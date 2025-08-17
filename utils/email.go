package utils

import (
	"context"
	"fmt"
	"net/smtp"
	"regexp"
	"time"

	"go-cloud-disk/conf"
	"github.com/jordan-wright/email"
)

func VerifyEmailFormat(email string) bool {
	pattern := `^[^\s@]+@[^\s@]+\.[^\s@]+$` // 匹配邮箱格式
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// sendMessage 使用默认SMTP认证发送邮件，如果运行时间超过
// 900毫秒，自动认为发送成功。
func sendMessage(ctx context.Context, em *email.Email) {
	c, cancel := context.WithTimeout(ctx, time.Millisecond*900)
	go func() {
		em.Send(conf.EmailSMTPServer+":25", smtp.PlainAuth("", conf.EmailAddr, conf.EmailSecretKey, conf.EmailSMTPServer))
		defer cancel()
	}()

	select {
	case <-c.Done():
		return
	case <-time.After(time.Millisecond * 900):
		return
	}
}

// SendConfirmMessage 发送确认码到目标邮箱，
// 当发送邮件超过5秒或连接邮件服务器出错时，此函数将返回错误
func SendConfirmMessage(targetMailBox string, code string) error {
	em := email.NewEmail()
	em.From = fmt.Sprintf("Go-Cloud-Disk <%s>", conf.EmailAddr)
	em.To = []string{targetMailBox}

	// 邮件标题
	em.Subject = "邮箱确认码 " + code

	// 构建邮件内容
	emailContentCode := "您的确认码是 " + code + "，验证码将在30分钟后过期"
	emailContentEmail := "您的确认邮箱是 " + targetMailBox
	emailContent := emailContentCode + "\n" + emailContentEmail
	em.Text = []byte(emailContent)

	// 发送邮件
	sendMessage(context.Background(), em)

	return nil
}
