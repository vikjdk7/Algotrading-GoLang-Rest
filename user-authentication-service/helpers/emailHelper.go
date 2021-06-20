package helper

import (
	"crypto/tls"
	"log"

	gomail "gopkg.in/mail.v2"
)

func SendEmail(otp string, email string) {
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
	d := gomail.NewDialer("smtp.gmail.com", 587, "neha190495@gmail.com", "fzhnpnkrlzksefti")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		log.Panic(err)
		return
	}

	return

}
