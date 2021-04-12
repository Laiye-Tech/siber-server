package libs

import (
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/bxcodec/faker/v3/support/slice"
	"math/rand"
	"time"
)

// Random 随机数类
type Random struct {
}

func NewRandom() *Random {
	return &Random{}
}

// GetRandomString 生成随机字符串
func (random *Random) GetRandomString(length int64) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := int64(0); i < length; i++ {
		result = append(result, bytes[r.Intn(int(len(bytes)))])
	}
	return string(result)
}

func (random *Random) GetRandomInt(intRange int) int {
	randSource := rand.NewSource(time.Now().Unix())
	randInt := rand.New(randSource).Intn(intRange)
	return randInt
}

func (random *Random) GetRandomPhone() string {
	out := ""
	boxDigitsStart := []string{"3", "4", "5", "6", "7", "8", "9"}
	ints, _ := faker.RandomInt(0, 9)
	for i := 0; i < 9; i++ {
		out += slice.IntToString(ints)[i]
	}
	return fmt.Sprintf("%s%s%s", "1", boxDigitsStart[rand.Intn(len(boxDigitsStart))], out)
}
