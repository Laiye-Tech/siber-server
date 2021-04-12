/**
* @Author: TongTongLiu
* @Date: 2020/3/10 8:13 下午
**/

package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func Test_graphQL(t *testing.T) {
	query := `"query getAccountSourceByHashId($hashId: String) {
		  getAccountSourceByHashId(hashId: $hashId) {
			id
			name
			registerContactTitle
			registerContactSubtitle
			registerContactImage
		  }
		}"`
	variable := `"{
		\"hashId\": "d51e79cd3ecafb8a9a3ed6cb8ae0861f"
		}"`

	operationName := "\"getAccountSourceByHashId\""

	requestStr := "\"query\":" + query + "," + "\"variables\":" + variable + "," + "\"operationName\":" + operationName
	requestStr = `{
    \"query\": \"query getAccountSourceByHashId($hashId: String) {\n getAccountSourceByHashId(hashId: $hashId) {\n                 id\n                    name\n                  registerContactTitle\n                  registerContactSubtitle\n                       registerContactImage\n            }\n           }\",
    \"variables\": {
        \"hashId\": \"d51e79cd3ecafb8a9a3ed6cb8ae0861f\"
    },
   \"operationName\": \"getAccountSourceByHashId\"
}`
	fmt.Println(requestStr)
	req, err := http.NewRequest("POST", "https://testplatform.wul.ai/api/paas-knowledge/graphql", strings.NewReader(requestStr))
	fmt.Println(req)
	fmt.Println(err)
}

func Test_queryDeal(t *testing.T){
	queryStr := "\nquery getAccountSourceByHashId($hashId: String) {\n  getAccountSourceByHashId(hashId: $hashId) {\n    id\n    name\n    registerContactTitle\n    registerContactSubtitle\n    registerContactImage\n  }\n}"
	fmt.Println("\"",queryStr, "\"")
	fmt.Println("-----")
	fmt.Println("\"",strings.Trim(queryStr, " "), "\"")
}