package libs

import (
	"encoding/base64"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"strconv"
)

const MaxSize = 10 * 1024 * 1024

func RemoteBase64(path string) (value string, err error) {
	resp, err := http.Get(path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		err = status.Errorf(codes.InvalidArgument, "Resource url error")
		return
	}
	cl := resp.Header.Get("Content-Length")
	imgSize, err := strconv.ParseInt(cl, 10, 64)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "Resource url error")
		return
	}
	if imgSize > MaxSize {
		err = status.Errorf(codes.InvalidArgument, "Image must not exceed a file size of 10MB")
		return
	}

	ct := resp.Header.Get("Content-Type")
	switch ct {
	case "image/gif", "image/jpeg", "image/jpg", "image/png", "image/tiff", "image/bmp", "application/pdf":
		imgBase64str := base64.StdEncoding.EncodeToString(body)
		return imgBase64str, err
	}
	err = status.Errorf(codes.InvalidArgument, "Resource type error")
	return "", err
}
