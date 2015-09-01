package utils

import (
  	"bufio"
  	"os"
	"net/http"
)


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

func GetFileSize(url string) int64{
	resp, err := http.Head(url)
	if err != nil {
		//log.Printf("URL %s is not reachable", url)
		return -1
	}
	if c := resp.StatusCode; c == 200 || (c > 300 && c <= 308){
		return resp.ContentLength
	}
	return -1
}