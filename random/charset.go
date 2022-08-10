package random

var (
	charsetAlpha string
	charsetASCII string
	charset      string
)

const charsetDigit = "0123456789"

func Charset() string      { return charset }
func CharsetAlpha() string { return charsetAlpha }
func CharsetDigit() string { return charsetDigit }
func CharsetASCII() string { return charsetASCII }

func init() {
	initCharset()
	initCharsetAlpha()
	initCharsetASCII()
}

func initCharsetAlpha() {
	for ch := 'a'; ch <= 'z'; ch++ {
		charsetAlpha += string(ch)
	}
	for ch := 'A'; ch <= 'Z'; ch++ {
		charsetAlpha += string(ch)
	}
}

func initCharsetASCII() {
	for i := 0; i <= 255; i++ {
		charsetASCII += string(rune(i))
	}
}

func initCharset() {
	type CharsetRange struct {
		Start int
		End   int
	}
	for _, r := range []CharsetRange{
		{
			Start: 33,
			End:   126,
		},
		{
			Start: 161,
			End:   767,
		},
		{
			Start: 880,
			End:   1159,
		},
		{
			Start: 1162,
			End:   1364,
		},
		{
			Start: 1567,
			End:   1610,
		},
		{
			Start: 1634,
			End:   1747,
		},
	} {
		for i := r.Start; i <= r.End; i++ {
			charset += string(rune(i))
		}
	}
}
