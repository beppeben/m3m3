package utils

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"regexp"
	"crypto/md5"
	"math/rand"
	"encoding/hex"
	"net/smtp"
	"io"
	"strconv"
	"time"
	//"image"
	"image/jpeg"
	"errors"
	"github.com/nfnt/resize"
)

var (
	emailVal		*regexp.Regexp
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type User struct {
	Name 	string
	Pass 	string
	Email 	string
}

type Comment struct {
	Id			int64		`json:"id,omitempty"`
	Item_id		int64		`json:"-"`
	Text			string		`json:"text,omitempty"`
	Author		string		`json:"author,omitempty"`
	Likes		int 			`json:"likes,omitempty"`
	Time			time.Time	`json:"-"`
}


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

func PanicOnErr(err error) {
    if err != nil {
        panic(err.Error())
    }
}


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
	//log.Printf("Measuring URL %s", url)
	resp, err := client.Head(url)	
	if err != nil {
		//log.Printf("URL %s is not reachable", url)
		return -1
	}
	defer resp.Body.Close()
	if c := resp.StatusCode; c == 200 || (c > 300 && c <= 308) {
		return resp.ContentLength
	}
	return -1
}

func SaveTempImage (url string, min_px_size int, exit_size uint, client *http.Client) error {
	//log.Printf("Getting URL %s", url)
	resp, err := client.Get(url)	
	if err != nil {
		//log.Printf("URL %s is not reachable", url)
		return err
	}
	defer resp.Body.Close()
	image, err := jpeg.Decode(resp.Body)
	if err != nil {
		return err
	}
	width := image.Bounds().Max.X - image.Bounds().Min.X
	height:= image.Bounds().Max.Y - image.Bounds().Min.Y

	if width < min_px_size && height < min_px_size {
		return errors.New("Image is too small")
	}
	image = resize.Thumbnail(exit_size, exit_size, image, resize.NearestNeighbor)
	filename := ComputeMd5(url) + ".jpg"
	file, err := os.Create(GetTempImgDir() + filename)
	if err != nil {
		log.Print("error creating file")
		return err
	}
	defer file.Close()
	opts := &jpeg.Options{Quality: 60}
	err = jpeg.Encode(file, image, opts)	
	return err
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

//replace with copy from tempid directory
/*
func SaveImageIfNeeded(item *Item){	
	if (item.Img_url == "" || item.Id == 0) {
		return
	}
	name := strconv.FormatInt(item.Id, 10) + ".jpg"
	if _, err := os.Stat(GetImgDir() + name); err == nil {
    		return
	}
	log.Printf("Saving image %s", item.Img_url)
	resp, err := http.Get(item.Img_url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	file, err := os.Create(GetImgDir() + name)	
 	if err != nil {
		log.Printf("[SERV] Could not create the file: %v", err)
     	return
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)	
 	if err != nil {
		log.Printf("[SERV] Could not save the image: %v", err)
     	return
	}	
	item.Img_url = ""
}
*/

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
