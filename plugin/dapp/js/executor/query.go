package executor

import (
	"fmt"

	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

func (c *js) Query_Query(payload *jsproto.Call) (types.Message, error) {
	execer := types.ExecName("user." + ptypes.JsX + "." + payload.Name)
	c.prefix = types.CalcLocalPrefix([]byte(execer))
	jsvalue, err := c.callVM("query", payload, nil, 0, nil)
	if err != nil {
		fmt.Println("query", err)
		return nil, err
	}
	str, err := getString(jsvalue, "result")
	if err != nil {
		fmt.Println("result", err)
		return nil, err
	}
	return &jsproto.QueryResult{Data: str}, nil
}
