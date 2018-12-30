package executor

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/robertkrimen/otto"
	"github.com/stretchr/testify/assert"
)

var jscode = `
function Exec(context) {
	this.context = context
}

function ExecLocal(context) {
	this.context = context
}

function ExecDelLocal(context) {
	this.context = context
}

Exec.prototype.hello = function(args) {
	return {args: args, action:"exec", context: this.context}
}

ExecLocal.prototype.hello = function(args) {
	return {args: args,  action:"execlocal", context: this.context}
}

ExecDelLocal.prototype.hello = function(args) {
	return {args: args, action:"execdellocal", context: this.context}
}
`

func callJsFunc(context *blockContext, code string, f string, args string) (otto.Value, error) {
	data, err := json.Marshal(context)
	if err != nil {
		return otto.Value{}, err
	}
	vm := otto.New()
	vm.Set("context", string(data))
	vm.Set("code", code)
	vm.Set("f", f)
	vm.Set("args", args)
	callfunc := "callcode(context, f, args)"
	return vm.Run(callcode + code + "\n" + callfunc)
}
func TestCallcode(t *testing.T) {
	value, err := callJsFunc(&blockContext{Height: 1}, jscode, "exec_hello", `{"hello2":"world2"}`)
	assert.Nil(t, err)
	assert.Equal(t, true, value.IsObject())
	action, err := value.Object().Get("action")
	assert.Nil(t, err)
	assert.Equal(t, "exec", action.String())
	args, err := value.Object().Get("args")
	assert.Nil(t, err)
	arg, err := args.Object().Get("hello2")
	assert.Nil(t, err)
	assert.Equal(t, "world2", arg.String())

	context, err := value.Object().Get("context")
	assert.Nil(t, err)
	cvalue, err := context.Object().Get("height")
	assert.Nil(t, err)
	assert.Equal(t, "1", cvalue.String())
	//test call error(invalid json input)
	_, err = callJsFunc(&blockContext{Height: 1}, jscode, "exec_hello", `{hello2":"world2"}`)
	_, ok := err.(*otto.Error)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, strings.Contains(err.Error(), "SyntaxError"))

	_, err = callJsFunc(&blockContext{Height: 1}, jscode, "hello", `{"hello2":"world2"}`)
	assert.Equal(t, true, strings.Contains(err.Error(), errInvalidFuncFormat.Error()))
	_, err = callJsFunc(&blockContext{Height: 1}, jscode, "hello_hello", `{"hello2":"world2"}`)
	assert.Equal(t, true, strings.Contains(err.Error(), errInvalidFuncPrefix.Error()))
	_, err = callJsFunc(&blockContext{Height: 1}, jscode, "exec_hello2", `{"hello2":"world2"}`)
	assert.Equal(t, true, strings.Contains(err.Error(), errFuncNotFound.Error()))
}
