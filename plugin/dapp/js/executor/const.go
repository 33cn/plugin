package executor

import "errors"

//ErrInvalidFuncFormat 错误的函数调用格式(没有_)
var errInvalidFuncFormat = errors.New("chain33.js: invalid function name format")

//ErrInvalidFuncPrefix not exec_ execloal execdellocal
var errInvalidFuncPrefix = errors.New("chain33.js: invalid function prefix format")

//ErrFuncNotFound 函数没有找到
var errFuncNotFound = errors.New("chain33.js: invalid function name not found")

var callcode = `
function kvcreator(dbtype) {
    this.data = {}
    this.kvs = []
    this.logs = []
    if (dbtype == "exec" || dbtype == "init") {
        this.get = getstatedb
    } else if (dbtype == "local") {
        this.get = getlocaldb
        this.list = listdb
    }
}

kvcreator.prototype.add = function(k, v) {
    var data = JSON.stringify(v) 
    this.data[k] = data
    this.kvs.push({key:k, value: data})
}

kvcreator.prototype.get = function(k) {
    var v
    if (this.data[k]) {
        v = this.data[k]
    } else {
        v = this.get(k)
    }
    if (!v) {
        return null
    }
    return JSON.parse(v)
}

kvcreator.prototype.listvalue = function(prefix, key, count, direction) {
   var values = this.list(prefix, key, count, direction)
   if (!values || values.length == 0) {
       return []
   }
   var objlist = []
   for (var i = 0; i < values.length; i++) {
       objlist.push(JSON.parse(values[i]))
   }
   return objlist
}

kvcreator.prototype.addlog = function(log) {
    if (this.list) {
        throw new Error("local or dellocal can't set log")
    }
    this.logs.push(JSON.stringify(log))
}

kvcreator.prototype.receipt = function() {
    return {kvs: this.kvs, logs: this.logs}
}

function callcode(context, f, args, loglist) {
	if (f == "init") {
		return Init(context)
	}
    var farr = f.split("_", 2)
    if (farr.length !=  2) {
        throw new Error("chain33.js: invalid function name format")
    }
    var prefix = farr[0]
    var funcname = farr[1]
    var runobj = {}
    var logs = []
    if (!Array.isArray(loglist)) {
        throw new Error("chain33.js: loglist must be array")
    }
    for (var i = 0; i < loglist.length; i++) {
        logs.push(JSON.parse(loglist[i]))
    }
    if (prefix == "exec") {
        runobj = new Exec(JSON.parse(context))
    } else if (prefix == "execlocal") {
        runobj = new ExecLocal(JSON.parse(context), logs)
    } else {
        throw new Error("chain33.js: invalid function prefix format")
    }
    var arg = JSON.parse(args)
    if (typeof runobj[funcname] != "function") {
        throw new Error("chain33.js: invalid function name not found")
    }
    return runobj[funcname](arg)
}
`
