/**
* @Author: TongTongLiu
* @Date: 2019/12/11 11:16 上午
**/

package dao

import (
	"context"
	"github.com/globalsign/mgo"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type CaseVersionStandard struct {
	CaseId         string
	VersionControl string
	RequestHeader  map[string]string
	UrlParameter   string
	RequestBody    string
	CheckPoint     []*CheckerStandard
	InjectPoint    []*siber.InjectSub
	SleepPoint     int64
	Remark         string
	InvalidDate    int64
	UserUpdate     string
	InsertTime     int64
	UpdateTime     int64
}
func VersionFormat(version string)(formatVersion string){
	formatVersion = strings.ToUpper(version)
	formatVersion = strings.TrimSpace(formatVersion)
	return
}
func (dao *Dao) SelectCaseVersion(ctx context.Context, caseVersionInput *CaseVersionStandard) (caseVersionOutput *CaseVersionStandard, err error) {
	err = isInputNull(ctx, caseVersionInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionCaseVersion)
	if err != nil {
		return
	}
	caseVersionOutput = new(CaseVersionStandard)
	// 根据caseID和version查询
	if caseVersionInput.CaseId == "" || caseVersionInput.VersionControl == "" {
		err = status.Errorf(codes.InvalidArgument, "SelectCaseVersion failed, caseid or version is nil")
		return
	}
	caseVersionInput.VersionControl = VersionFormat(caseVersionInput.VersionControl)
	condition := []bson.M{
		bson.M{"caseid": caseVersionInput.CaseId},
		bson.M{"versioncontrol": caseVersionInput.VersionControl},
		bson.M{"invaliddate": ValidTag},
	}
	err = c.Find(bson.M{"$and": condition}).One(&caseVersionOutput)
	return
}

func (dao *Dao) InsertCaseVersion(ctx context.Context, caseVersionInput *CaseVersionStandard) (caseVersionOutput *CaseVersionStandard, err error) {
	if caseVersionInput == nil {
		err = status.Errorf(codes.InvalidArgument, "InsertCaseVersion failed, caseVersionInput is nil")
		return
	}
	c, err := getMongoCollection(ctx, CollectionCaseVersion)
	if err != nil {
		return
	}
	caseVersionInput.InsertTime = time.Now().Unix()
	caseVersionInput.UpdateTime = time.Now().Unix()
	caseVersionInput.VersionControl = VersionFormat(caseVersionInput.VersionControl)

	err = c.Insert(caseVersionInput)
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "duplicate case version")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "failed to persistence case version")
		return
	}
	caseVersionOutput, err = NewDao().SelectCaseVersion(ctx, caseVersionInput)
	return
}

func (dao *Dao) UpdateCaseVersion(ctx context.Context, caseVersionInput *CaseVersionStandard) (caseVersionOutput *CaseVersionStandard, err error) {
	err = isInputNull(ctx, caseVersionInput)
	if err != nil {
		return
	}
	if caseVersionInput.CaseId == "" || caseVersionInput.VersionControl == "" {
		err = status.Errorf(codes.InvalidArgument, "UpdateCaseVersion failed ,CaseId or VersionControl is nil")
		return
	}
	c, err := getMongoCollection(ctx, CollectionCaseVersion)
	if err != nil {
		return
	}
	caseVersionInput.UpdateTime = time.Now().Unix()
	caseVersionInput.VersionControl = VersionFormat(caseVersionInput.VersionControl)
	condition := []bson.M{
		bson.M{"caseid": caseVersionInput.CaseId},
		bson.M{"versioncontrol": caseVersionInput.VersionControl},
		bson.M{"invaliddate": ValidTag},
	}
	_, err = c.Upsert(bson.M{"$and": condition}, bson.M{"$set": caseVersionInput})
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "case version duplicate")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "failed to update case version")
		return
	}
	caseVersionOutput, err = NewDao().SelectCaseVersion(ctx, caseVersionInput)
	if caseVersionOutput == nil {
		return
	}
	return
}

func (dao *Dao) DeleteCaseVersion(ctx context.Context, caseVersionInput *CaseVersionStandard) (caseVersionOutput *CaseVersionStandard, err error) {
	err = isInputNull(ctx, caseVersionInput)
	if err != nil {
		return
	}
	if caseVersionInput.CaseId == "" || caseVersionInput.VersionControl == "" {
		err = status.Errorf(codes.InvalidArgument, "DeleteCaseVersion failed ,CaseId or VersionControl is nil")
		return
	}
	c, err := getMongoCollection(ctx, CollectionCaseVersion)
	if err != nil {
		return
	}
	caseVersionInput.VersionControl = VersionFormat(caseVersionInput.VersionControl)
	condition := []bson.M{
		bson.M{"caseid": caseVersionInput.CaseId},
		bson.M{"versioncontrol": caseVersionInput.VersionControl},
		bson.M{"invaliddate": ValidTag},
	}
	selector := bson.M{"$and": condition}
	data := bson.M{"$set": bson.M{"invaliddate": time.Now().Unix()}}
	err = c.Update(selector, data)
	caseVersionOutput = new(CaseVersionStandard)
	return
}

// 严格查询
// 用于展示指定case的version信息
func (dao *Dao) ListCaseVersionByID(ctx context.Context, caseVersionInfo *CaseVersionStandard) (list *[]*CaseVersionStandard, err error) {
	if caseVersionInfo == nil || caseVersionInfo.CaseId == "" {
		err = status.Errorf(codes.InvalidArgument, "ListCaseVersionByID failed, caseVersionInfo is nil")
		return
	}
	c, err := getMongoCollection(ctx, CollectionCaseVersion)
	if err != nil {
		return
	}
	var caseVersions []*CaseVersionStandard
	err = c.Find(bson.M{"caseid": caseVersionInfo.CaseId, "invaliddate": ValidTag}).Select(bson.M{"caseid": -1, "versioncontrol": -1}).All(&caseVersions)
	if err != nil {
		return
	}
	list = &caseVersions
	return
}

// 用于混合/统计类查询
// 比如：查询某个迭代新增case数等
func (dao *Dao) ListCaseVersion(ctx context.Context, info siber.FilterInfo) (caseVersionList *[]*CaseVersionStandard, err error) {
	return
}
