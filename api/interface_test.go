package api

//
//func TestGRPCInterface(t *testing.T) {
//	cc, err := grpc.Dial("172.17.202.23:9000", grpc.WithInsecure())
//	assert.Nil(t, err)
//
//	path := "/Users/dongmengnan/Works/programs/SaaS/im-saas-msgs-protos/"
//	importPaths := []string{
//		path,
//		fmt.Sprintf("%s/protos", path),
//		fmt.Sprintf("%s/protos/saas.openapi.v2", path),
//		fmt.Sprintf("%s/protos/im_user_attribute", path),
//	}
//	protoFiles := []string{fmt.Sprintf("%s/protos/saas.openapi.v2/openapi.v2.proto", path)}
//	descSource, err := grpcurl.DescriptorSourceFromProtoFiles(importPaths, protoFiles...)
//	assert.Nil(t, err)
//	request := strings.NewReader(`{
//	"user_id": "1",
//	"msg_body": {
//		"text": {"content": "123"}
//	}
//}`)
//	rf, formatter, err := grpcurl.RequestParserAndFormatterFor(grpcurl.FormatJSON, descSource, true, true, request)
//	assert.Nil(t, err)
//	h := grpcurl.NewDefaultEventHandler(os.Stdout, descSource, formatter, true)
//	err = grpcurl.InvokeRPC(context.Background(), descSource, cc, "saas.openapi.v2.MessageService/GetBotResponse", []string{}, h, rf.Next)
//	assert.Nil(t, err)
//	grpcurl.PrintStatus(os.Stderr, h.Status, formatter)
//	t.Logf("%+v", formatter)
//}
//
//func TestGRPC_Invoke(t *testing.T) {
//	g := new(GRPCInterface)
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
//	g.ImportPaths = importPaths
//	g.ProtoFiles = protoFiles
//	g.MethodName = "manage_user.ManageUserService/CreateUser"
//	request := payload.Request{
//		Payload: payload.Payload{
//			Body: []byte(`{
//		"channelId": 9577968,
//		"username":  "t0809",
//		"nickname":  "Nicktest",
//		"userType":  1,
//		"status":    1
//		}`),
//		},
//	}
//
//	pResp := g.Invoke(&request)
//
//	fmt.Println("pResp is :", libs.JsonWrapper{pResp})
//}
//
//func TestRESTFUL_Invoke(t *testing.T) {
//	r := &HTTPInterface{
//		Interface: nil,
//		ReqMode:   "POST",
//		ReqPath:   "http://172.17.202.22:51480/im-saas/user/create",
//		ReqHeader: nil,
//	}
//	r.URI()
//	request := payload.Request{
//		Payload: payload.Payload{
//			Body: []byte(`{
//		"channelId": 9577968,
//		"username":  "h0905",
//		"nickname":  "Nicktest",
//		"userType":  1,
//		"status":    1
//		}`),
//		},
//	}
//
//	resp := r.Invoke(&request)
//
//	fmt.Println("resp is :", resp)
//}
