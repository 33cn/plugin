package executor

func calcLocalPrefix(execer []byte) []byte {
	s := append([]byte("LODB-"), execer...)
	s = append(s, byte('-'))
	return s
}

func calcStatePrefix(execer []byte) []byte {
	s := append([]byte("mavl-"), execer...)
	s = append(s, byte('-'))
	return s
}

func calcCodeKey(name string) []byte {
	return append([]byte("mavl-js-code-"), []byte(name)...)
}

func calcRollbackKey(hash []byte) []byte {
	return append([]byte("LODB-js-rollback-"), hash...)
}
