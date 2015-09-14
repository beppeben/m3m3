package utils

import (
	"bufio"
	//"log"
	"net/http"
	"os"
	"regexp"
	"crypto/md5"
	"math/rand"
	"encoding/hex"
	"net/smtp"
)

var (
	emailVal		*regexp.Regexp
)


func SendEmail(toEmail string, subject string, body string) error {
	auth := smtp.PlainAuth("", GetServiceEmail(), GetEmailPass(), GetSMTP())	
	to := []string{toEmail}
	msg := []byte(
		"To: " + toEmail + "\r\n" +
		"From: " + GetServiceEmail() + "\r\n" +
		"Subject: " + subject + "\r\n" +
		/*
		"Content-type: text/HTML; charset=UTF-8" +
		"format: flowed" +
		"Content-Transfer-Encoding: 8bit" +
		*/
		"\r\n" + body + "\r\n")
	err := smtp.SendMail(GetSMTP() + ":" + GetSMTPPort(), auth, GetServiceEmail(), to, msg)
	return err
}

type User struct {
	Name string
	Pass string
	Email string
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

func ComputeMd5(text string) string {
	array := md5.Sum([]byte(text))
	return hex.EncodeToString(array[:])
}

func ValidateEmail(email string) bool {
 	emailVal = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
 	return emailVal.MatchString(email)
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func GetFileSize(url string, client *http.Client) int64 {
	//resp, err := http.Head(url)
	resp, err := client.Head(url)	
	if err != nil {
		//log.Printf("URL %s is not reachable", url)
		return -1
	}
	defer resp.Body.Close()
	if c := resp.StatusCode; c == 200 || (c > 300 && c <= 308) {
		return resp.ContentLength
	}
	//log.Print("failed to get file size")
	return -1
}
