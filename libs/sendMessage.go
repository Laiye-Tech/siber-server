package libs

import (
	"api-test/configs"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func GetRandomString(length int64) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytess := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := int64(0); i < length; i++ {
		result = append(result, bytess[r.Intn(int(len(bytess)))])
	}
	return string(result)
}

func GetHeaders(choose int) map[string]string {
	var secret, pubkey string
	if choose == 0 {
		secret = configs.GetGlobalConfig().Wechat.WechatSecret
		pubkey = configs.GetGlobalConfig().Wechat.WechatPubkey
	} else {
		secret = configs.GetGlobalConfig().Wechat.GroupSecret
		pubkey = configs.GetGlobalConfig().Wechat.GroupKey
	}
	nonce := GetRandomString(32)
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	s := fmt.Sprintf("%s%s%s", nonce, timestamp, secret)
	t := sha1.New()
	_, _ = io.WriteString(t, s)
	sign := fmt.Sprintf("%x", t.Sum(nil))
	return map[string]string{
		"Api-Auth-pubkey":    pubkey,
		"Api-Auth-nonce":     nonce,
		"Api-Auth-timestamp": timestamp,
		"Api-Auth-sign":      sign,
	}
}

func sendPicture(user string, content string) {
	if user == "" {
		return
	}
	res := GetHeaders(1)
	v := map[string]interface{}{}
	vv := map[string]interface{}{}
	v["image"] = vv
	vv["resource_url"] = content
	message := map[string]interface{}{
		"msg_body": v,
		"send_ts":  res["Api-Auth-timestamp"],
		"user_id":  user,
	}
	bytesRepresentation, err := json.Marshal(&message)
	if err != nil {
		Log().Error(nil, "Json unmarshal error: ", err)
		return
	}
	sending(bytesRepresentation, res)
}

func sending(bytesRepresentation []byte, res map[string]string) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", configs.GetGlobalConfig().Wechat.WechatUrl, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		Log().Error(nil, "Make wechat request error: ", err)
	}

	req.Header.Set("Api-Auth-pubkey", res["Api-Auth-pubkey"])
	req.Header.Set("Api-Auth-nonce", res["Api-Auth-nonce"])
	req.Header.Set("Api-Auth-timestamp", res["Api-Auth-timestamp"])
	req.Header.Set("Api-Auth-sign", res["Api-Auth-sign"])
	resp, err := client.Do(req)
	if err != nil {
		Log().Error(nil, "Send wechat error: ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	Log().Info(nil, "Send message: %+v", string(body))
}

func sendWechat(user string, content string) {
	if user == "" {
		return
	}
	res := GetHeaders(0)
	v := map[string]interface{}{}
	vv := map[string]interface{}{}
	v["text"] = vv
	vv["content"] = content
	message := map[string]interface{}{
		"msg_body": v,
		"send_ts":  res["Api-Auth-timestamp"],
		"user_id":  user,
	}
	bytesRepresentation, err := json.Marshal(&message)
	if err != nil {
		Log().Error(nil, "Json unmarshal error: ", err)
		return
	}
	sending(bytesRepresentation, res)
}

func sendWechatGroup(user string, content string) {
	if user == "" {
		return
	}
	Log().Info(nil, "Wechat user is: %+v", user)
	Log().Info(nil, "Wechat content is: %+v", content)
	res := GetHeaders(1)
	v := map[string]interface{}{}
	vv := map[string]interface{}{}
	v["text"] = vv
	vv["content"] = content
	message := map[string]interface{}{
		"msg_body": v,
		"send_ts":  res["Api-Auth-timestamp"],
		"user_id":  user,
	}
	bytesRepresentation, err := json.Marshal(&message)
	if err != nil {
		Log().Error(nil, "Json unmarshal failed: %+v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", configs.GetGlobalConfig().Wechat.WechatUrl, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		Log().Error(nil, "Send message failed: %+v", err)
	}

	req.Header.Set("Api-Auth-pubkey", res["Api-Auth-pubkey"])
	req.Header.Set("Api-Auth-nonce", res["Api-Auth-nonce"])
	req.Header.Set("Api-Auth-timestamp", res["Api-Auth-timestamp"])
	req.Header.Set("Api-Auth-sign", res["Api-Auth-sign"])
	resp, err := client.Do(req)
	if err != nil {
		Log().Error(nil, "Send message failed: %+v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	Log().Info(nil, "Message body: %+v", body)
}
