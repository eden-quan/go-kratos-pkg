package randpkg

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"time"
)

const (
	// Set of characters to use for generating random strings
	CharsetLowercase    = "abcdefghijklmnopqrstuvwxyz"
	CharsetAlphabet     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	CharsetNumeral      = "1234567890"
	CharsetHex          = "1234567890abcdef"
	CharsetAlphanumeric = CharsetAlphabet + CharsetNumeral
)

var random *rand.Rand

func init() {
	seed, err := NewCryptoSeed()
	if err != nil {
		seed = NewTimeSeededSource()
	}
	// 根据种子生成一个新的随机器
	random = rand.New(rand.NewSource(seed))
}

func NewCryptoSeed() (int64, error) {
	var seed int64
	err := binary.Read(crand.Reader, binary.BigEndian, &seed)
	if err != nil {
		return 0, err
	}
	return seed, nil
}

func NewTimeSeededSource() int64 {
	return time.Now().UnixNano()
}

// Intn returns a random integer in the range [0, n).
func Intn(n int) int {
	return random.Intn(n)
}

// IntRange returns a random integer in the range from min to max.
func IntRange(min, max int) int {
	if min > max {
		return 0
	}
	if min == max {
		return min
	}
	r := random.Intn(max - min)
	return min + r
}

// String returns a random string n characters long, composed of entities
// from charset.
func String(n int, charset string) string {
	randStr := make([]byte, n) // Random string to return
	charLen := len(charset)
	for i := 0; i < n; i++ {
		j := random.Intn(charLen)
		randStr[i] = charset[j]
	}
	return string(randStr)
}

// Alphanumeric 从数字字符集+大小写字符集生成指定长度的随机字符串
func Alphanumeric(n int) string {
	return String(n, CharsetAlphanumeric)
}

// AlphabetLower 从小写字符集生成指定长度的随机字符串
func AlphabetLower(n int) string {
	return String(n, CharsetLowercase)
}

// Alphabet 从大小写字符集生成指定长度的随机字符串
func Alphabet(n int) string {
	return String(n, CharsetAlphabet)
}

// NumeralStr 从数字字符集生成指定长度的随机字符串
func NumeralStr(n int) string {
	return String(n, CharsetNumeral)
}

// HexStr 从16进制字符集生成指定长度的随机字符串
func HexStr(n int) string {
	return String(n, CharsetHex)
}

// ChoiceString 从字符串数组中随机选择一个字符串
func ChoiceString(choices []string) string {
	var res string
	length := len(choices)
	i := random.Intn(length)
	res = choices[i]
	return res
}

// ChoiceInt 从整数数组中随机选择一个整数
func ChoiceInt(choices []int) int {
	var res int
	length := len(choices)
	i := random.Intn(length)
	res = choices[i]
	return res
}
