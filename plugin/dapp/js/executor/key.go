package executor

func calcCodeKey(name string) []byte {
	return append([]byte("mavl-js-code-"), []byte(name)...)
}
