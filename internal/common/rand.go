package common

import (
	"fmt"
	"strconv"
	"time"

	"golang.org/x/exp/rand"
)

func RandId() int {
	// 初始化随机数种子
	rand.Seed(uint64(time.Now().UnixNano()))

	// 生成一个随机整数
	randomInt := rand.Int()
	fmt.Println("Random integer:", randomInt)

	// 定义范围
	min := 100
	max := 900

	// 生成 [min, max) 范围内的随机整数
	randomIntInRange := rand.Intn(max-min) + min
	s := time.Now().Unix()
	str := fmt.Sprintf("%d%d", s, randomIntInRange)
	res, _ := strconv.Atoi(str)
	return int(res)
}
