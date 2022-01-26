package types

import "fmt"

// some errors definition
var (
	ErrAssetAmount  = fmt.Errorf("%s", "The asset amount is not valid!")
	ErrAssetPrice   = fmt.Errorf("%s", "The asset price is not valid!")
	ErrAssetOp      = fmt.Errorf("%s", "The asset op is not define!")
	ErrAssetBalance = fmt.Errorf("%s", "Insufficient balance!")
	ErrOrderSatus   = fmt.Errorf("%s", "The order status is reovked or completed!")
	ErrAddr         = fmt.Errorf("%s", "Wrong Addr!")
	ErrAsset        = fmt.Errorf("%s", "The asset's execer or symbol can't be nil,The same assets cannot be exchanged!")
	ErrCount        = fmt.Errorf("%s", "The param count can't large  20")
	ErrDirection    = fmt.Errorf("%s", "The direction only 0 or 1!")
	ErrStatus       = fmt.Errorf("%s", "The status only in  0 , 1, 2!")
	ErrOrderID      = fmt.Errorf("%s", "Wrong OrderID!")

	ErrCfgFmt   = fmt.Errorf("%s", "ErrCfgFmt")
	ErrBindAddr = fmt.Errorf("%s", "The address is not bound")
)
