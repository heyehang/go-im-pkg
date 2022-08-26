package trand

import (
	"crypto/rand"
	"math/big"
	"strings"
	"sync"

	"github.com/google/uuid"
)

const (
	RandSourceNumber             = "0123456789"
	RandSourceLetterAndNumber    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	RandSourceUppercase          = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	RandSourceLetter             = "abcdefghijklmnopqrstuvwxyz"
	RandSourceLetterAndUppercase = RandSourceUppercase + RandSourceLetter
	RandSourceSymbols            = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~" // 32
	RandSourceSymbolAndLetter    = RandSourceSymbols + RandSourceLetterAndNumber
	saltBit                      = uint(8)             // 随机因子二进制位数
	saltShift                    = uint(8)             // 随机因子移位数
	increasShift                 = saltBit + saltShift // 自增数移位数
)

var _mi *Mist

func init() {
	_mi = newMist()
}

/*
获取安全随机数 , 获取n以内的随机数
*/
func RandInt(n int64) (randNum int64) {
	num, _ := rand.Int(rand.Reader, big.NewInt(n))
	randNum = num.Int64()
	return
}

/*
min 获取区间的最小值
max 获取区间的最大值
@return 获取 min - max 区间的随机数
*/
func RandNInt(min, max int64) int64 {
	randValue := int64(0)
	randValue = RandInt(max)
	if randValue < min {
		randValue = randValue + min
	}
	return randValue
}

/*
source 随机串种子
len 长度
*/
func RandNString(source string, n int) string {
	src := []rune(source)
	builder := new(strings.Builder)
	for i := 0; i < n; i++ {
		builder.WriteRune(src[RandInt(int64(len(src)))])
	}
	return builder.String()
}

type Mist struct {
	sync.Mutex       // 互斥锁
	increas    int64 // 自增数
	saltA      int64 // 随机因子一
	saltB      int64 // 随机因子二
}

/* 初始化 Mist 结构体*/
func newMist() *Mist {
	mist := Mist{increas: 0}
	return &mist
}

/* 生成int型唯一编号 */
func (c *Mist) generateIntID() int64 {
	c.Lock()
	c.increas++
	// 获取随机因子数值 ｜ 使用真随机函数提高性能
	randA, _ := rand.Int(rand.Reader, big.NewInt(255))
	c.saltA = randA.Int64()
	randB, _ := rand.Int(rand.Reader, big.NewInt(255))
	c.saltB = randB.Int64()
	// 通过位运算实现自动占位
	mist := int64((c.increas << increasShift) | (c.saltA << saltShift) | c.saltB)
	c.Unlock()
	return mist
}

/*
生成int型唯一编号,自增的
总耗时  28631695000 生成个数  100000000
每秒350万个左右
*/
func GetIntIncreId() int64 {
	return _mi.generateIntID()
}

/*
获取安全字符串,包含大写，小写，和 symbol
n 字符串长度,n 必须大于4
*/
func GetSecKey(n int) string {
	if n < 4 {
		return RandNString(RandSourceLetterAndNumber, 4)
	}
	secStr := ""
	// 是否包含特殊字符
	sym := RandNString(RandSourceSymbols, n/4)
	// 大写
	up := RandNString(RandSourceUppercase, n/4)
	// 是否包含小写
	letter := RandNString(RandSourceLetter, n/4)
	numLen := n - ((n / 4) * 3)
	// 数字
	num := RandNString(RandSourceNumber, numLen)
	secStr = sym + up + letter + num
	// 乱序
	srcRuneList := []rune(secStr)
	for i := n - 1; i > 0; i-- {
		num := RandInt(int64(i + 1))
		srcRuneList[i], srcRuneList[num] = srcRuneList[num], srcRuneList[i]
	}
	secStr = string(srcRuneList)
	return secStr
}

/*
获取唯一uuid
*/
func GetUUID() string {
	uuid.SetRand(rand.Reader)
	uid := uuid.New()
	return uid.String()
}
