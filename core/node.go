/**
* @Author: TongTongLiu
* @Date: 2020/4/20 7:07 下午
**/

package core

import (
	"api-test/api"
	"api-test/dao"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetNodeInfo(ctx context.Context, request *siber.NodeInfoRequest) (response *siber.NodeInfoResponse, err error) {
	if request == nil || request.Id == "" {
		xzap.Logger(ctx).Warn("request or node id is nil")
		err = status.Errorf(codes.InvalidArgument, "request or node id is nil")
		return
	}
	// 这里filter要求精确，不同于列表的模糊搜索
	fileter := &siber.FilterInfo{
		FilterContent: map[string]string{
			request.NodeType: request.Id,
		},
	}
	var children []*siber.NodeInfoResponse
	var childrenNum int
	switch request.NodeType {
	case dao.PackageNode:
		resp, err := api.GetServiceList(ctx, fileter)
		if err != nil {
			return nil, err
		}
		for i, r := range resp.ServiceName {
			child := &siber.NodeInfoResponse{
				Id:   r,
				NodeName: r,
				NodeType: dao.ServiceNode,
			}
			children = append(children, child)
			childrenNum = i + 1
		}
	case dao.ServiceNode:
		resp, err := api.GetMethodList(ctx, fileter)
		if err != nil {
			return nil, err
		}
		for i, r := range resp.MethodList {
			child := &siber.NodeInfoResponse{
				Id:   r.MethodName,
				NodeName: r.MethodName,
				NodeType: dao.MethodNode,
			}
			children = append(children, child)
			childrenNum = i + 1
		}
	case dao.MethodNode:
		resp, err := ManageCaseList(ctx, fileter)
		if err != nil {
			return nil, err
		}
		for _, r := range resp.CaseInfoList {
			child := &siber.NodeInfoResponse{
				Id:   r.CaseId,
				NodeName: r.CaseName,
				NodeType: dao.CaseNode,
			}
			children = append(children, child)
		}
		childrenNum = int(resp.TotalNum)
	case dao.CaseNode:
		resp, err := ManageFlowList(ctx, fileter)
		if err != nil {
			return nil, err
		}
		for _, r := range resp.FlowInfoList {
			child := &siber.NodeInfoResponse{
				Id:   r.FlowId,
				NodeName: r.FlowName,
				NodeType: dao.FlowNode,
			}
			children = append(children, child)
		}
		childrenNum = int(resp.TotalNum)
	case dao.FlowNode:
		resp, err := ManagePlanList(ctx, fileter)
		if err != nil {
			return nil, err
		}
		for _, r := range resp.PlanInfoList {
			child := &siber.NodeInfoResponse{
				Id:   r.PlanId,
				NodeName: r.PlanName,
				NodeType: dao.PlanNode,
			}
			children = append(children, child)
		}
		childrenNum = int(resp.TotalNum)
	default:
		xzap.Logger(ctx).Warn("unsupported node type: %s", zap.String("type", request.NodeType))
		err = status.Errorf(codes.InvalidArgument, "unsupported node type: %s", request.NodeType)
		return
	}

	response = &siber.NodeInfoResponse{
		Id:       request.Id,
		NodeName:     request.Id,
		NodeType:     request.NodeType,
		ChildrenNum:  int32(childrenNum),
		Children: children,
	}
	return
}
