package init

import (
	_ "github.com/33cn/plugin/plugin/store/kvdb"   // register kvdb package
	_ "github.com/33cn/plugin/plugin/store/kvmvcc" // register kvmvcc package
	_ "github.com/33cn/plugin/plugin/store/mpt"    // register mpt package
)
