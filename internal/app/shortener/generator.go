package shortener

const (
	// alphabet58 is the base58 alphabet without ambiguous characters
	alphabet58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	base58     = uint64(len(alphabet58))
)

func padLeft(s string, length int, padChar byte) string {
	if len(s) >= length {
		return s
	}
	padCount := length - len(s)
	padding := make([]byte, padCount)
	for i := range padding {
		padding[i] = padChar
	}
	return string(padding) + s
}

func Generate(num uint64, length int) string {
	if num == 0 {
		return padLeft(string(alphabet58[0]), length, alphabet58[0])
	}

	// Encode to base58: append digits in reverse order
	var chars []byte
	for num > 0 {
		idx := num % base58
		chars = append(chars, alphabet58[idx])
		num /= base58
	}

	// Reverse to correct order
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return padLeft(string(chars), length, alphabet58[0])
}
