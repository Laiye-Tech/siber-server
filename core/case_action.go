package core

import (
	"api-test/dao"
	"api-test/payload"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"strings"
	"time"
)

type CaseAction interface {
	Render(ctx context.Context, v *payload.Variable) (err error)
	Create()
	LinkedByCase(p *Case)
	Execute(ctx context.Context, p *Case) error
}

const (
	SUCCESS = "SUCCESS"
	FAILED  = "FAILED"
)

const (
	ResponseBody   = "$ResponseBody"
	ResponseHeader = "$ResponseHeader"
	ResponseTime   = "$ResponseTime"
	ResponseStatus = "$ResponseStatus"
	RequestBody    = "$RequestBody"
	RequestHeader  = "$RequestHeader"
)

const (
	EqualCheckerType      = "*payload.EqualChecker"
	NotEqualChecker       = "*payload.NotEqualChecker"
	IncludeCheckerType    = "*payload.IncludeChecker"
	NotIncludeCheckerType = "*payload.NotIncludeChecker"
)

type ActionConsequence struct {
	Status    string
	ErrorInfo error
}
type CheckPoint struct {
	CaseAction  `json:"CaseAction,omitempty"`
	Selector    payload.Selector
	Checker     payload.Checker
	Consequence *ActionConsequence
}

type InjectPoint struct {
	CaseAction   `json:"InjectPoint,omitempty"`
	VariableName string
	Selector     payload.Selector
	Consequence  *ActionConsequence
}

type SleepPoint struct {
	CaseAction    `json:"SleepPoint,omitempty"`
	SleepDuration time.Duration
}

func (cp *CheckPoint) Render(ctx context.Context, v *payload.Variable) (err error) {
	if cp == nil || cp.Checker == nil {
		err = status.Errorf(codes.FailedPrecondition, "checkpoint or checkpoint checker is nil")
		return
	}
	checkerType := reflect.TypeOf(cp.Checker).String()
	err = cp.Selector.Render(ctx, v)
	if err != nil {
		return
	}
	if checkerType == EqualCheckerType || checkerType == NotEqualChecker || checkerType == IncludeCheckerType || checkerType == NotIncludeCheckerType {
		err = cp.Checker.Render(ctx, v)
		return
	}
	return
}

func (sp *SleepPoint) Render(ctx context.Context, v *payload.Variable) (err error) {
	return
}

func (ip *InjectPoint) Render(ctx context.Context, v *payload.Variable) (err error) {
	err = ip.Selector.Render(ctx, v)
	if err != nil {
		return
	}
	return
}

func createSelector(key string) (selector payload.Selector) {
	keyInfo := strings.Split(key, ".")
	var selectorName string
	if len(keyInfo) == 0 {
		return
	} else {
		selectorName = keyInfo[0]
	}
	switch selectorName {
	case ResponseStatus:
		return &payload.StatusCodeSelector{
			Name: selectorName,
		}
	case ResponseTime:
		return &payload.CostTimeSelector{
			Name: selectorName,
		}

	}
	if len(keyInfo) <= 1 {
		return
	}
	keyName := strings.Join(keyInfo[1:], ".")
	switch selectorName {
	case ResponseBody:
		return &payload.ResponseBodySelector{Key: keyName, Name: selectorName}
	case RequestBody:
		return &payload.RequestBodySelector{Key: keyName, Name: selectorName}
	case ResponseHeader:
		return &payload.ResponseHeaderSelector{Key: keyName, Name: selectorName}
	case RequestHeader:
		return &payload.RequestHeaderSelector{Key: keyName, Name: selectorName}

	}
	return
}

func createChecker(ctx context.Context, actionSub *dao.CheckerStandard) (checker payload.Checker, err error) {
	switch actionSub.Relation {
	case "=":
		checker = &payload.EqualChecker{
			ExpectValue: actionSub.Content,
			Name:        actionSub.Relation,
		}
	case "!=":
		checker = &payload.NotEqualChecker{
			ExpectValue: actionSub.Content,
			Name:        actionSub.Relation,
		}
	case "length":
		length, ok := actionSub.Content.(float64)
		if ok {
			checker = &payload.LengthChecker{
				Name:   actionSub.Relation,
				Length: length,
			}
		} else {
			err = status.Errorf(codes.InvalidArgument, "length is not int,%v", actionSub.Content)
			xzap.Logger(ctx).Warn("length is not int", zap.Any("content", actionSub.Content))
			return
		}
	case "in":
		checker = &payload.InChecker{
			ExpectValue: actionSub.Content,
			Name:        actionSub.Relation,
		}
	case "include":
		checker = &payload.IncludeChecker{
			ExpectValue: actionSub.Content,
			Name:        actionSub.Relation,
		}
	case "not include":
		checker = &payload.NotIncludeChecker{
			ExpectValue: actionSub.Content,
			Name:        actionSub.Relation,
		}
	case "exist":
		content, ok := actionSub.Content.(bool)
		if ok {
			checker = &payload.ExistChecker{
				ExpectValue: content,
				Name:        actionSub.Relation,
			}
		} else {
			err = status.Errorf(codes.InvalidArgument, "exist is not bool,%v", actionSub.Content)
			xzap.Logger(ctx).Warn("exist is not bool", zap.Any("content", actionSub.Content))
			return
		}
	case payload.GreaterThan, payload.LessThan, payload.GreaterThanOrEqualTo, payload.LessThanOrEqualTo:
		checker = &payload.ComparisonChecker{
			ExpectValue: actionSub.Content,
			Name:        actionSub.Relation,
		}
	}

	if checker == nil {
		err = status.Errorf(codes.InvalidArgument, "unsupported checker type, %s", actionSub.Relation)
		xzap.Logger(ctx).Warn("unsupported checker type", zap.Any("content", actionSub.Content))
		return
	}
	return
}

func (ip *InjectPoint) Execute(ctx context.Context, p *Case) error {
	_, value, err := ip.Selector.Select(ctx, p.Request, p.Response)
	if err != nil {
		ip.Consequence = &ActionConsequence{
			Status:    FAILED,
			ErrorInfo: err,
		}
		return err
	}
	p.Flow.Variable.Set(ctx, ip.VariableName, value)
	ip.Consequence = &ActionConsequence{
		Status: SUCCESS,
	}
	return nil
}

func (cp *CheckPoint) Execute(ctx context.Context, p *Case) (err error) {
	err = p.actionRender(ctx, p.Flow.Variable)
	if err != nil {
		return
	}
	if cp.Checker == nil || cp.Selector == nil {
		xzap.Logger(ctx).Warn("selector or checker is nil")
		return status.Errorf(codes.NotFound, "selector or checker is nil")
	}
	keyExists, value, err := cp.Selector.Select(ctx, p.Request, p.Response)
	if err != nil {
		cp.Consequence = &ActionConsequence{
			Status:    FAILED,
			ErrorInfo: err,
		}
		return err
	}
	if cp.Checker.GetCheckerName() != "ExistChecker" && !keyExists {
		err = status.Errorf(codes.InvalidArgument, "key does not exist in the json")
		cp.Consequence = &ActionConsequence{
			Status:    FAILED,
			ErrorInfo: err,
		}
		return err
	}
	err = cp.Checker.Check(keyExists, value)
	if err != nil {
		xzap.Logger(ctx).Warn("checker check failed", zap.Any("err", err))
		cp.Consequence = &ActionConsequence{
			Status:    FAILED,
			ErrorInfo: err,
		}
		return err
	}
	cp.Consequence = &ActionConsequence{
		Status: SUCCESS,
	}
	return err
}

func (sp *SleepPoint) Execute(ctx context.Context, p *Case) error {
	time.Sleep(sp.SleepDuration * time.Second)
	return nil
}
