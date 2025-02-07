package utils

import (
	"context"
	"fmt"
	"net/smtp"
	"os"

	"firebase.google.com/go/auth"
)

// SendEmail Function to send email
func SendEmail(to, subject, body string) error {
	// SMTP configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	from := smtpUsername

	// Constructing email headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""

	// Compose the email message
	var msg string
	for key, value := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	msg += "\r\n" + body

	// Connect to the SMTP server with TLS
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpHost, smtpPort), auth, from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func SendVerificationEmail(user *auth.UserRecord) error {
	// Generate email verification link with settings
	settings := &auth.ActionCodeSettings{
		URL:             "https://wegrowparenting.com",
		HandleCodeInApp: true,
	}
	// Send email with the verification link
	link, err := FirebaseAuth.EmailVerificationLinkWithSettings(context.Background(), user.Email, settings)
	if err != nil {
		return fmt.Errorf("error generating email verification link: %v", err)
	}
	// Log the verification link
	fmt.Printf("Verification link for user %s: %s\n", user.Email, link)

	// Construct the email body with the new content
	body := fmt.Sprintf(`
		<p>Dear Family Member,</p>
		<p>Congratulations on starting this beautiful journey of Parenthood! Remember, we're here to support you at every step of the way. From expert guidance to a warm community, you’ve got a partner in us. Verify your email below to start exploring.</p>
		<p><a href="%s">🔗 Verify Your Email</a></p>
		<p>💡 Follow us on <a href="https://www.instagram.com/wegrowparenting?igsh=MXhtMTFmcHh1NDI4YQ==">Instagram</a> and <a href="https://youtube.com/@wegrowparenting?si=0hZc-iH0msSHBblE">YouTube</a> for daily parenting tips & support!</p>
		<p>Happy Parenting! 🌿💚</p>
	`, link)

	// Call the sendEmail function to send the email
	err = SendEmail(user.Email, "Welcome to We Grow Family! 💚", body)
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	fmt.Println("Verification email sent successfully")
	return nil
}

func SendPasswordResetEmail(email string) error {
	// Generate password reset link
	settings := &auth.ActionCodeSettings{
		URL:             "https://wegrowparenting.com",
		HandleCodeInApp: true,
	}

	link, err := FirebaseAuth.PasswordResetLinkWithSettings(context.Background(), email, settings)
	if err != nil {
		return fmt.Errorf("error generating password reset link: %v", err)
	}
	// Log the password reset link
	fmt.Printf("Password reset link for user %s: %s\n", email, link)

	// Construct the email body with engaging content
	body := fmt.Sprintf(`
		<p>Dear We Grow Family Member,</p>
		<p>We understand that sometimes passwords slip our minds. No worries! You can reset your password quickly by clicking the link below.</p>
		<p><a href="%s">🔗 Reset Your Password</a></p>
		<p>Need further assistance? Feel free to reach out to our support team. We’re always here to help!</p>
		<p>💡 Stay connected with us on <a href="https://www.instagram.com/wegrowparenting?igsh=MXhtMTFmcHh1NDI4YQ==">Instagram</a> and <a href="https://youtube.com/@wegrowparenting?si=0hZc-iH0msSHBblE">YouTube</a> for daily parenting tips & support!</p>
		<p>Warm regards,<br/>We Grow Team 🌿💚</p>
	`, link)

	// Call the sendEmail function to send the email
	err = SendEmail(email, "Reset Your Password - We Grow", body)
	if err != nil {
		return fmt.Errorf("error sending password reset email: %v", err)
	}

	fmt.Println("Password reset email sent successfully")

	return nil
}

func ResendVerificationEmail(email string) error {
	_, err := FirebaseAuth.GetUserByEmail(context.Background(), email)
	if err != nil {
		return fmt.Errorf("error fetching user data: %v", err)
	}

	// Generate email verification link with settings
	settings := &auth.ActionCodeSettings{
		URL:             "https://wegrowparenting.com",
		HandleCodeInApp: true,
	}
	link, err := FirebaseAuth.EmailVerificationLinkWithSettings(context.Background(), email, settings)
	if err != nil {
		return fmt.Errorf("error generating email verification link: %v", err)
	}
	// Log the verification link
	fmt.Printf("Verification link for user %s: %s\n", email, link)

	// Construct the email body with the new content
	body := fmt.Sprintf(`
		<p>Dear Family Member,</p>
		<p>Congratulations on starting this beautiful journey of Parenthood! Remember, we're here to support you at every step of the way. From expert guidance to a warm community, you’ve got a partner in us. Verify your email below to start exploring.</p>
		<p><a href="%s">🔗 Verify Your Email</a></p>
		<p>💡 Follow us on <a href="https://www.instagram.com/wegrowparenting?igsh=MXhtMTFmcHh1NDI4YQ==">Instagram</a> and <a href="https://youtube.com/@wegrowparenting?si=0hZc-iH0msSHBblE">YouTube</a> for daily parenting tips & support!</p>
		<p>Happy Parenting! 🌿💚</p>
	`, link)

	// Call the sendEmail function to send the email
	err = SendEmail(email, "Welcome to We Grow Family! 💚", body)
	if err != nil {
		return fmt.Errorf("error sending verification email: %v", err)
	}

	fmt.Println("Verification email sent successfully")
	return nil
}
