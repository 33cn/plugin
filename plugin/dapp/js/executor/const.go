package executor

import "errors"

//ErrInvalidFuncFormat 错误的函数调用格式(没有_)
var errInvalidFuncFormat = errors.New("chain33.js: invalid function name format")

//ErrInvalidFuncPrefix not exec_ execloal_ query_
var errInvalidFuncPrefix = errors.New("chain33.js: invalid function prefix format")

//ErrFuncNotFound 函数没有找到
var errFuncNotFound = errors.New("chain33.js: invalid function name not found")

var callcode = `
var tojson = JSON.stringify
function kvcreator(dbtype) {
    this.data = {}
    this.kvs = []
	this.logs = []
	this.type = dbtype
	this.getstate = getstatedb
	this.getloal = getlocaldb
	this.list = listdb
    if (dbtype == "exec" || dbtype == "init") {
		this.get = getstatedb
    } else if (dbtype == "local") {
        this.get = getlocaldb
    } else if (dbtype == "query") {
		this.get = getlocaldb
	} else {
		throw new Error("chain33.js: dbtype error")
	}
}

kvcreator.prototype.add = function(k, v) {
	if (typeof v != "string") {
		v = JSON.stringify(v) 
	}
    this.data[k] = v
    this.kvs.push({key:k, value: v})
}

kvcreator.prototype.get = function(k) {
    var v
    if (this.data[k]) {
        v = this.data[k]
    } else {
		var dbvalue = this.get(k)
		if (dbvalue.err != "") {
			return null
		}
		v = dbvalue.value
    }
    if (!v) {
        return null
    }
    return JSON.parse(v)
}

kvcreator.prototype.listvalue = function(prefix, key, count, direction) {
   var dbvalues = this.list(prefix, key, count, direction)
   if (dbvalues.err != "") {
	   return []
   }
   var values = dbvalues.value
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
    if (this.type != "exec") {
        throw new Error("local or query can't set log")
	}
	if (typeof v != "string") {
		log = JSON.stringify(log) 
	}
    this.logs.push(log)
}

kvcreator.prototype.receipt = function() {
    return {kvs: this.kvs, logs: this.logs}
}

function callcode(context, f, args, loglist) {
	if (f == "init") {
		return Init(JSON.parse(context))
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
	} else if (prefix == "query") {
		runobj = new Query(JSON.parse(context))
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
