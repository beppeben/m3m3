package utils

import (
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"os"
	"bufio"
	"net/http"
	"io"
	"image/jpeg"
	"image"
	"github.com/nfnt/resize"
	"errors"
	"strconv"
	"archive/zip"
	"mime/multipart"
	"path/filepath"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func PositivePart (n int) int {
	if n >= 0 {
		return n
	} else {
		return 0
	}
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

func FilterResizeImage (image image.Image) (image.Image, error) {
	width := image.Bounds().Max.X - image.Bounds().Min.X
	height:= image.Bounds().Max.Y - image.Bounds().Min.Y
	if (width < 300 && height < 300) || height > width*13/10 {
		return image, errors.New("Bad image format")
	}
	if width > 1000 {
		image = resize.Thumbnail(600, 600, image, resize.NearestNeighbor)
	}	
	return image, nil
}

func GetJpegFromUrl (url string, client *http.Client) (image.Image, error) {
	var result image.Image
	resp, err := client.Get(url)	
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	return jpeg.Decode(resp.Body)	
}


func SaveTempImage (url string, client *http.Client, dir string) (string, error) {
	image, err := GetJpegFromUrl(url, client)
	if err != nil {
		return "", err
	}
	image, err = FilterResizeImage(image)
	if err != nil {
		return "", err
	}
	hash := Hash(url)
	file, err := os.Create(dir + hash + ".jpg")
	if err != nil {
		return "", err
	}
	defer file.Close()
	opts := &jpeg.Options{Quality: 60}
	err = jpeg.Encode(file, image, opts)	
	return hash, err
}

func PersistTempImage (tid string, id int64, tempdir string, dir string) error {
	if (tid == "" || id == 0) {
		return errors.New("Error persisting image, empty names")
	}
	orig_name := tid + ".jpg"
	orig, err := os.Open(tempdir + orig_name)
	if err != nil {
		return err
	}
	defer orig.Close()
	dest_name := tid + "-" + strconv.FormatInt(id, 10) + ".jpg"
	dest, err := os.Create(dir + dest_name)
	if err != nil {
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, orig)
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

func ExtractZipToDir(file multipart.File, length int64, dest string) error {
	r, err := zip.NewReader(file, length)
	if err != nil {
        return err
    }
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
