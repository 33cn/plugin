package plugin

import (
	_ "github.com/33cn/plugin/plugin/consensus/init" //consensus init
	_ "github.com/33cn/plugin/plugin/crypto/init"    //crypto init
	_ "github.com/33cn/plugin/plugin/dapp/init"      //dapp init
	_ "github.com/33cn/plugin/plugin/store/init"     //store init
)