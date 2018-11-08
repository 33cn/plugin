package init

import (
	_ "./store/kvdb"
	_ "./store/kvmvcc"
	_ "./store/mpt"
)
