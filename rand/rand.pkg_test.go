package randpkg

import "testing"

func TestAll(t *testing.T) {
	TestIntn(t)
	TestIntRange(t)
	TestString(t)
	TestAlphanumeric(t)
	TestAlphabetLower(t)
	TestAlphabet(t)
	TestNumeralStr(t)
	TestHexStr(t)
	TestChoiceString(t)
	TestChoiceInt(t)
}

func TestIntn(t *testing.T) {
	t.Logf("Intn(10) = %v\n", Intn(10))
	t.Logf("Intn(9999) = %v\n", Intn(9999))
}

func TestIntRange(t *testing.T) {
	t.Logf("IntRange(1, 10) = %v\n", IntRange(1, 10))
	t.Logf("IntRange(100, 999) = %v\n", IntRange(100, 999))
}

func TestString(t *testing.T) {
	num := 16
	charset := CharsetAlphabet
	t.Logf("String(%d, %v) = %v\n", num, charset, String(num, charset))
}

func TestAlphanumeric(t *testing.T) {
	num := 16
	t.Logf("Alphanumeric(%d) = %v\n", num, Alphanumeric(num))
}

func TestAlphabetLower(t *testing.T) {
	num := 32
	t.Logf("AlphabetLower(%d) = %v\n", num, AlphabetLower(num))
}

func TestAlphabet(t *testing.T) {
	num := 32
	t.Logf("Alphabet(%d) = %v\n", num, Alphabet(num))
}

func TestNumeralStr(t *testing.T) {
	num := 4
	t.Logf("NumeralStr(%d) = %v\n", num, NumeralStr(num))
}

func TestHexStr(t *testing.T) {
	num := 16
	t.Logf("HexStr(%d) = %v\n", num, HexStr(num))
}

func TestChoiceString(t *testing.T) {
	stringList := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	t.Logf("ChoiceString(%+v) = %v\n", stringList, ChoiceString(stringList))
}

func TestChoiceInt(t *testing.T) {
	numberList := []int{1, 2, 33, 44, 5, 66, 7, 8, 9}
	t.Logf("ChoiceInt(%+v) = %v\n", numberList, ChoiceInt(numberList))
}
