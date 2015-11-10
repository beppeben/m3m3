package utils

import (
	log "github.com/Sirupsen/logrus"
	"net/smtp"
	"time"
	"math/rand"
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"os"
	"bufio"
	"net/http"
	"io"
	"image/jpeg"
	"github.com/nfnt/resize"
	"errors"
	"strconv"
	"archive/zip"
	"mime/multipart"
	"path/filepath"
)


const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"


func SendEmailOnce(toEmail string, subject string, body string) error {
	auth := smtp.PlainAuth("", GetServiceEmail(), GetEmailPass(), GetSMTP())
	to := []string{toEmail}
	msg := []byte(
		"To: " + toEmail + "\r\n" +
		"From: " + GetServiceEmail() + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")
	err := smtp.SendMail(GetSMTP() + ":" + GetSMTPPort(), auth, GetServiceEmail(), to, msg)
	if err != nil {
		log.Infof("Can't send email to %s: %v", toEmail, err)
	}
	
	return err
}

func SendEmail(toEmail string, subject string, body string) error {
	retries := 4
	var err error
	for retries > 0 {
		err = SendEmailOnce(toEmail, subject, body)
		if err == nil {
			break
		} else {
			retries--
			time.Sleep(time.Second*10)
		}
	}
	if err != nil {
		log.Warnf("Can't send email to %s: %v", toEmail, err)
	}
	return err
}

func RandString(n int) string {
 	b := make([]byte, n)
  	for i := range b {
 		b[i] = letterBytes[rand.Intn(len(letterBytes))]
  	}
 	return string(b)
}

func Hash(text string) string {
	array := md5.Sum([]byte(text))
	return hex.EncodeToString(array[:])
}

func ValidateEmail(email string) bool {
 	emailVal := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
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
	resp, err := client.Head(url)	
	if err != nil {
		log.Debug(err.Error())
		return -1
	}
	defer resp.Body.Close()
	if c := resp.StatusCode; c == 200 || (c > 300 && c <= 308) {
		return resp.ContentLength
	}
	return -1
}
/*
func Log(msg interface{}) {
	log.Printf("%s", msg)
}

func LogAndWrite(w io.Writer, err error, msg string) {
	if err != nil {
		log.Printf("%s", err)
		fmt.Fprintf(w, msg)
		return
	}
}
*/

func SaveTempImage (url string, client *http.Client) (string, error) {
	resp, err := client.Get(url)	
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	image, err := jpeg.Decode(resp.Body)
	if err != nil {
		return "", err
	}
	width := image.Bounds().Max.X - image.Bounds().Min.X
	height:= image.Bounds().Max.Y - image.Bounds().Min.Y

	if (width < 500 && height < 500) || height > width {
		return "", errors.New("Bad image format")
	}
	if width > 1000 {
		image = resize.Thumbnail(600, 600, image, resize.NearestNeighbor)
	}	
	hash := Hash(url)
	file, err := os.Create(GetTempImgDir() + hash + ".jpg")
	if err != nil {
		return "", err
	}
	defer file.Close()
	opts := &jpeg.Options{Quality: 60}
	err = jpeg.Encode(file, image, opts)	
	return hash, err
}

func PersistTempImage (tid string, id int64) error {
	if (tid == "" || id == 0) {
		return errors.New("Error persisting image, empty names")
	}
	orig_name := tid + ".jpg"
	orig, err := os.Open(GetTempImgDir() + orig_name)
	if err != nil {
		return err
	}
	defer orig.Close()
	dest_name := strconv.FormatInt(id, 10) + ".jpg"
	dest, err := os.Create(GetImgDir() + dest_name)
	if err != nil {
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, orig)
	return err
}

func DeleteAllImages() error {
	err := DeleteFilesInDir(GetImgDir())
	if err != nil {
		return err
	}
	err = DeleteFilesInDir(GetTempImgDir())
	return err
}

func DeleteFilesInDir(dir string) error {
	d, err := os.Open(dir)
		if err != nil {
     	return err
 	}
	defer d.Close()
	files, err := d.Readdir(-1)
	if err != nil {
    		return err
	}
	for _, file := range files {
    		if file.Mode().IsRegular() {
        		err = os.Remove(dir + file.Name()) 
   		}
 	}
	return err
}

func DeleteTempImage(tid string) error {
	err := os.Remove(GetTempImgDir() + tid + ".jpg")
	return err
}

func ExtractZipToHttpDir(file multipart.File, length int64) error {
	r, err := zip.NewReader(file, length)
	if err != nil {
        return err
    }
	dest := GetHTTPDir()
    os.MkdirAll(dest, 0755)

    extractAndWriteFile := func(f *zip.File) error {
        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer func() {
            if err := rc.Close(); err != nil {
                panic(err)
            }
        }()

        path := filepath.Join(dest, f.Name)

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, f.Mode())
        } else {
            f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer func() {
                if err := f.Close(); err != nil {
                    panic(err)
                }
            }()

            _, err = io.Copy(f, rc)
            if err != nil {
                return err
            }
        }
        return nil
    }

    for _, f := range r.File {
        err := extractAndWriteFile(f)
        if err != nil {
            return err
        }
    }

    return nil
}

func PositivePart (n int) int {
	if n >= 0 {
		return n
	} else {
		return 0
	}
}

