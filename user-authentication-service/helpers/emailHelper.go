package helper

import (
	"crypto/tls"
	"fmt"
	"log"

	gomail "gopkg.in/mail.v2"
)

func SendSignUpEmail(otp string, email string, first_name string, last_name string) {
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", "hedgina@hedgina.com")

	// Set E-Mail receivers
	m.SetHeader("To", email)

	// Set E-Mail subject
	m.SetHeader("Subject", "New Account Signup at Hedgina")

	// Set E-Mail body. You can set plain text or html with text/html
	emailBody := fmt.Sprintf("<div><p>Dear %s %s,</p><p>Greetings!</p><p>You are just a step away from creating your Hedgina account.</p><p>We are sharing a verification code to create your account. The code is valid for 10 minutes and usable only once.</p><p>Once you have verified the code, you will be redirected to the login page. This is to ensure that only you have access to your account.</p><pre><table><tbody><tr><td><b>One time password:</b></td><td>%s</td></tr><tr><td><b>Expires in:</b></td><td>10 minutes</td></tr></tbody></table></pre><p>Best Regards,<br />Team Hedgina</p></div>", first_name, last_name, otp)
	m.SetBody("text/html", emailBody)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "vikjdk7@gmail.com", "pspdumazkddtzyop")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		log.Panic(err)
		return
	}

	return
}

func SendForgotPasswordEmail(otp string, email string) {
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", "hedgina@hedgina.com")

	// Set E-Mail receivers
	m.SetHeader("To", email)

	// Set E-Mail subject
	m.SetHeader("Subject", "Hedgina's Forgot Password Email")

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", "One Time Password: "+otp)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "vikjdk7@gmail.com", "pspdumazkddtzyop")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		log.Panic(err)
		return
	}

	return

}
