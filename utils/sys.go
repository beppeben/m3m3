package utils

import (
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

type SysConfig interface {
	GetTempImgDir() string
	GetImgDir() string
	GetHTTPDir() string
	GetDropFile() string
	GetCreateFile() string
	GetRSSFile() string
}

type SysUtils struct {
	config SysConfig
}

func NewSysUtils(config SysConfig) *SysUtils {
	return &SysUtils{config}
}

func (u *SysUtils) GetDropStatements() ([]string, error) {
	return ReadLines(u.config.GetDropFile())
}

func (u *SysUtils) GetCreateStatements() ([]string, error) {
	return ReadLines(u.config.GetCreateFile())
}

func (u *SysUtils) GetRSSDefs() ([]string, error) {
	return ReadLines(u.config.GetRSSFile())
}

func (u *SysUtils) SaveTempImage(url string, client *http.Client) (string, error) {
	return SaveTempImage(url, client, u.config.GetTempImgDir())
}

func (u *SysUtils) PersistTempImage(tid string, id int64) error {
	return PersistTempImage(tid, id, u.config.GetTempImgDir(), u.config.GetImgDir())
}

func (u *SysUtils) DeleteImages() error {
	return DeleteFilesInDir(u.config.GetImgDir())
}

func (u *SysUtils) DeleteTempImages() error {
	return DeleteFilesInDir(u.config.GetTempImgDir())
}

func (u *SysUtils) DeleteTempImage(tid string) error {
	err := os.Remove(u.config.GetTempImgDir() + tid + ".jpg")
	return err
}

func (u *SysUtils) DeleteImage(id int64, tid string) error {
	err := os.Remove(u.config.GetImgDir() + tid + "-" + strconv.FormatInt(id, 10) + ".jpg")
	return err
}

func (u *SysUtils) ExtractZipToHttpDir(file multipart.File, length int64) error {
	return ExtractZipToDir(file, length, u.config.GetHTTPDir())
}
