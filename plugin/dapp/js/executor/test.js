//数据结构设计
function Init(context) {
    this.kvc = new kvcreator("init")
    this.context = context
    this.kvc.add("action", "init")
    this.kvc.add("context", this.context)
    this.kvc.add("randnum", GetRandnum())
    return this.kvc.receipt()
}

Exec.prototype.hello = function(args) {
    this.kvc.add("args", args)
    this.kvc.add("action", "exec")
    this.kvc.add("context", this.context)
    this.kvc.addlog({"key1": "value1"})
    this.kvc.addlog({"key2": "value2"})
	return this.kvc.receipt()
}

ExecLocal.prototype.hello = function(args) {
    this.kvc.add("args", args)
    this.kvc.add("action", "execlocal")
    this.kvc.add("log", this.logs)
    this.kvc.add("context", this.context)
	return this.kvc.receipt()
}

//return a json string
Query.prototype.hello = function(args) {
	var obj = getlocaldb("context")
	return tojson(obj)
}
