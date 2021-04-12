package core

import (
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"testing"
)

//func TestPlan_Run(t *testing.T) {
//	trigger := &Trigger{
//		Protocol: "grpc",
//	}
//
//	path := configs.GetGlobalConfig().ProtoFile.RootPath
//	importPaths := []string{
//		path,
//		fmt.Sprintf("%s/protos", path),
//		fmt.Sprintf("%s/protos/saas.openapi.v2", path),
//		fmt.Sprintf("%s/protos/im_user_attribute", path),
//		fmt.Sprintf("%s/protos/im_user_attribute", path),
//		fmt.Sprintf("%s/protos/manage_user", path),
//	}
//	protoFiles := []string{fmt.Sprintf("%s/protos/manage_user/manage_user.proto", path)}
//	fmt.Println(protoFiles)
//	g := &api.GRPCInterface{
//		Interface:   nil,
//		ImportPaths: importPaths,
//		ProtoFiles:  protoFiles,
//		MethodName:  "manage_user.ManageUserService/CreateUser",
//	}
//	g2 := &api.GRPCInterface{
//		Interface:   nil,
//		ImportPaths: importPaths,
//		ProtoFiles:  protoFiles,
//		MethodName:  "manage_user.ManageUserService/CreateUser",
//	}
//
//	methodCase1 := &api.Method{Interfaces: map[api.Protocol]api.Interface{
//		"grpc": g,
//	}}
//	methodCase2 := &api.Method{Interfaces: map[api.Protocol]api.Interface{
//		"grpc": g2,
//	}}
//
//	plan := &Plan{
//		Id:   1,
//		Name: "测试plan1",
//		Items: []*PlanItem{
//			{
//				Id:    1,
//				Order: 1,
//				Flow: &Flow{
//					Id:   1,
//					Name: "测试flow1",
//					Items: []*FlowItem{
//						{
//							Id:    1,
//							Order: 1,
//							Case: &Case{
//								Id:     1,
//								Name:   "测试case1",
//								Method: methodCase1,
//								Request: &payload.Request{
//									Payload: payload.Payload{
//										Body: []byte(`{
//		"channelId": 9577968,
//		"username":  "t0905-03",
//		"nickname":  "Nicktest",
//		"userType":  1,
//		"status":    1
//		}`),
//									},
//								},
//								Actions: []CaseAction{
//									CheckPoint{
//										Selector: &payload.HeaderSelector{
//											Key: "content-type",
//										},
//										Checker: &payload.EqualChecker{
//											ExpectValue: "application/grpc",
//										},
//									},
//									CheckPoint{
//										Selector: &payload.StatusCodeSelector{},
//										Checker: payload.EqualChecker{
//											ExpectValue: 0,
//										},
//									},
//									CheckPoint{
//										Selector: &payload.BodySelector{
//											Key: "username",
//										},
//										Checker: payload.EqualChecker{
//											ExpectValue: "t0905-03",
//										},
//									},
//									SleepPoint{
//										SleepDuration: time.Second * 3,
//									},
//									InjectPoint{
//										Selector: &payload.BodySelector{
//											Key: "username",
//										},
//										VariableName: "username",
//									},
//								},
//							},
//						},
//						{
//							Id:    2,
//							Order: 2,
//							Case: &Case{
//								Id:     2,
//								HashId: "case_hash2",
//								Name:   "测试case2",
//								Method: methodCase2,
//								Request: &payload.Request{
//									Payload: payload.Payload{
//										Body: []byte(`{
//		"channelId": 9577968,
//		"username":  "{{VARIABLE.case_hash1.username}}",
//		"nickname":  "newNicktest",
//		"userType":  1,
//		"status":    1
//		}`),
//									},
//								},
//								Actions: []CaseAction{
//									CheckPoint{
//										Selector: &payload.BodySelector{
//											Key: "username",
//										},
//										Checker: payload.EqualChecker{
//											ExpectValue: "t0905-03",
//										},
//									},
//								},
//							},
//						},
//					},
//					Variable:      &payload.Variable{},
//					BeforeActions: []*FlowActionItem{},
//				},
//			},
//		},
//	}
//	for _, flow := range plan.Items {
//		flow.Flow.Plan = plan
//		for _, c := range flow.Flow.Items {
//			c.Case.Flow = flow.Flow
//		}
//	}
//	assert.Nil(t, plan.Run(context.Background(), trigger))
//}

/*
func TestPlan_RunRestful(t *testing.T) {
	trigger := &Trigger{
		Protocol: "restful",
	}
	g := &api.HTTPInterface{
		Interface: nil,
		ReqMode:   "POST",
		ReqHeader: nil,
		ReqPath:   "http://172.17.202.22:51480/im-saas/user/create",
	}
	g2 := &api.HTTPInterface{
		Interface: nil,
		ReqMode:   "POST",
		ReqHeader: nil,
		ReqPath:   "http://172.17.202.22:51480/im-saas/user/create",
	}

	methodCase1 := &api.Method{Interfaces: map[api.Protocol]api.Interface{
		"restful": g,
	}}
	methodCase2 := &api.Method{Interfaces: map[api.Protocol]api.Interface{
		"restful": g2,
	}}

	plan := &Plan{
		Id:   1,
		Name: "测试plan1",
		Flows: []*Flow{
			{
				Id:    1,

				Flow: &Flow{
					Id:   1,
					Name: "测试flow1",
					Items: []*FlowItem{
						{
							Id:    1,
							Order: 1,
							Case: &Case{
								Id:     1,
								Name:   "测试case1",
								Method: methodCase1,
								Request: &payload.Request{
									Payload: payload.Payload{
										Body: []byte(`{
		"channelId": 9577968,
		"username":  "t0905-03",
		"nickname":  "Nicktest",
		"userType":  1,
		"status":    1
		}`),
									},
								},
								Actions: []CaseAction{
									CheckPoint{
										Selector: &payload.HeaderSelector{
											Key: "Content-Type",
										},
										Checker: &payload.EqualChecker{
											ExpectValue: "application/json",
										},
									},
									CheckPoint{
										Selector: &payload.StatusCodeSelector{},
										Checker: payload.EqualChecker{
											ExpectValue: 200,
										},
									},
									CheckPoint{
										Selector: &payload.BodySelector{
											Key: "username",
										},
										Checker: payload.EqualChecker{
											ExpectValue: "t0905-03",
										},
									},
									SleepPoint{
										SleepDuration: time.Second * 3,
									},
									InjectPoint{
										Selector: &payload.BodySelector{
											Key: "username",
										},
										VariableName: "username",
									},
								},
							},
						},
						{
							Id:    2,
							Order: 2,
							Case: &Case{
								Id:     2,
								Name:   "测试case2",
								Method: methodCase2,
								Request: &payload.Request{
									Payload: payload.Payload{
										Body: []byte(`{
		"channelId": 9577968,
		"username":  "{{VARIABLE.case_hash1.username}}",
		"nickname":  "newNicktest",
		"userType":  1,
		"status":    1
		}`),
									},
								},
								Actions: []CaseAction{
									CheckPoint{
										Selector: &payload.BodySelector{
											Key: "username",
										},
										Checker: payload.EqualChecker{
											ExpectValue: "t0905-03",
										},
									},
								},
							},
						},
					},
					Variable:      &payload.Variable{},
					BeforeActions: []*FlowActionItem{},
				},
			},
		},
	}
	//for _, flow := range plan.Flows {
	//	flow..Plan = plan
	//	for _, c := range flow.Flow.Items {
	//		c.Case.Flow = flow.Flow
	//	}
	//}
	assert.Nil(t, plan.Run(context.Background(), trigger))
}
*/

//func TestPlan_RunMongo(t *testing.T) {
//	initial.Initial()
//	planInput := &siber.PlanInfo{
//		//manage-user
//		PlanId: "5ddf3e59e17563001f29139c",
//	}
//	p, err := DescribePlan(context.Background(), planInput, "")
//	if err != nil {
//		return
//	}
//	err = p.Run(context.Background())
//	return
//
//}

//func TestPlan_RunOpenAPI(t *testing.T) {
//	planInput := &siber.PlanInfo{
//		PlanId: "5ddf3e59e17563001f29139c",
//	}
//	p, err := DescribePlan(context.Background(), planInput, "")
//	if err != nil {
//		return
//	}
//	err = p.Run(context.Background())
//	return
//}

func Test_planFormatVerify(t *testing.T) {
	info := &siber.PlanInfo{

		TriggerCondition: []*siber.TriggerCondition{
			&siber.TriggerCondition{
				TriggerCron: "*/3 * */13 * * ?",
			},
		},
	}
	err := planFormatVerify(context.Background(), info)
	fmt.Println(err)
}
