package api

import (
	"api-test/configs"
	"context"
	"encoding/json"
	"fmt"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const PrivateProject = "Private"
var IterationCache = cache.New(8*time.Hour, 60*time.Minute)

type TapdRes struct {
	Status int32
	Data   []DataItem
	Info   string
}

type DataItem struct {
	Iteration Iteration
}

type Iteration struct {
	Id          string
	Name        string
	WorkspaceId string
	Startdate   string
	Enddate     string
	Status      string
	Creator     string
	Created     string
	Modified    string
	ReleaseId   string
	Description string
}

func IfContainsNumbers(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Number, v) {
			count++
		}
		if count >= 2 {
			return true
		}
	}
	return false
}

func RegexVersion(str string) (version int32, err error) {
	reg1 := regexp.MustCompile(`(\d+\.)?(\d+\.)?(\d+)`)
	versionInfo := reg1.FindString(str)
	s := strings.Split(versionInfo, ".")
	if len(s) == 3 {
		versionInfo = fmt.Sprintf("%s%02s%02s", s[0], s[1], s[2])
	}
	if len(s) == 2 {
		versionInfo = fmt.Sprintf("%s%02s00", s[0], s[1])
	}
	versionNumber, errInfo := strconv.ParseInt(versionInfo, 10, 32)
	return int32(versionNumber), errInfo
}

func CIIterations(ctx context.Context, request *siber.GetIterationsRequest) (*siber.GetIterationsResponse, error) {
	privateDeploy := configs.GetGlobalConfig().Flag.PrivateDeploy
	if privateDeploy {
		version := configs.GetGlobalConfig().Flag.Version
		resp := &siber.GetIterationsResponse{
			IterationList: []*siber.IterationsInfo{{
				ProjectName: PrivateProject,
				CurrentIteration:  version,
				HistoryIterations: []string{version},
			},
			},
		}
		return resp, nil
	}
	IterationsResponseCache, found := IterationCache.Get("tempCache")
	if found {
		return IterationsResponseCache.(*siber.GetIterationsResponse), nil
	}
	var IterationListAppend []*siber.IterationsInfo
	for _, r := range configs.GetGlobalConfig().Tapd {
		client := &http.Client{}
		tapdUrl := fmt.Sprintf("https://api.tapd.cn/iterations?workspace_id=%s&limit=25", r.WorkId)
		httpReq, err := http.NewRequest("GET", tapdUrl, nil)
		if err != nil {
			xzap.Sugar(ctx).Errorf("Make http request error: %+v", err)
			return nil, err
		}
		httpReq.SetBasicAuth(r.ApiUser, r.ApiPassword)
		resp, err := client.Do(httpReq)
		if err != nil {
			xzap.Sugar(ctx).Errorf("http error: %+v, body: %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var res TapdRes
		err = json.Unmarshal(body, &res)
		if err != nil {
			xzap.Sugar(ctx).Errorf("Json unmarshal error: %+v, body: %v", err, string(body))
			return nil, err
		}
		var currentIteration string
		var iterationsDetail []*siber.IterationsDetail
		// 计算出当前版本
		for _, v := range res.Data {
			if IfContainsNumbers(v.Iteration.Name) == false {
				continue
			}
			AfterIterationVersion, err := RegexVersion(v.Iteration.Name)
			if err != nil {
				return nil, err
			}
			IterationDetail := &siber.IterationsDetail{AfterVersion: AfterIterationVersion, CurrentVersion: v.Iteration.Name}
			iterationsDetail = append(iterationsDetail, IterationDetail)
			timeLayout := "2006-01-02 15:04:05"
			loc, _ := time.LoadLocation("Local")
			startTime, _ := time.ParseInLocation(timeLayout, v.Iteration.Startdate+" 10:00:00", loc)
			endTime, _ := time.ParseInLocation(timeLayout, v.Iteration.Enddate+" 10:00:00", loc)
			if startTime.Sub(time.Now()) <= 0 && endTime.Sub(time.Now()) >= 0 {
				currentIteration = v.Iteration.Name
			}
		}

		sort.Slice(iterationsDetail, func(i, j int) bool {
			return iterationsDetail[i].AfterVersion > iterationsDetail[j].AfterVersion
		})
		if currentIteration == "" && iterationsDetail != nil {
			currentIteration = iterationsDetail[0].CurrentVersion
		}
		var historyIterations []string
		for _, v := range iterationsDetail {
			historyIterations = append(historyIterations, v.CurrentVersion)
		}
		GetIterationsResponse := &siber.IterationsInfo{ProjectName: r.Name, CurrentIteration: currentIteration, HistoryIterations: historyIterations, IterationsDetail: iterationsDetail}
		IterationListAppend = append(IterationListAppend, GetIterationsResponse)
	}
	GetIterationsResponse := &siber.GetIterationsResponse{IterationList: IterationListAppend}
	IterationCache.Set("tempCache", GetIterationsResponse, cache.DefaultExpiration)
	return GetIterationsResponse, nil
}
