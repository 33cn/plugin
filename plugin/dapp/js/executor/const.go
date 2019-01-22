package executor

var callcode = `
var tojson = JSON.stringify
//table warp
function Table(kvc, config, defaultvalue) {
    var ret = table_new(tojson(config), tojson(defaultvalue))
    if (ret.err) {
        throw new Error(ret.err)
    }
    this.kvc = kvc
    this.id = ret.id
    this.config = config
    this.name = config["#tablename"]
    this.defaultvalue = defaultvalue
}

function isstring(obj) {
    return typeof obj === "string"
}

Table.prototype.add = function(obj) {
    if (!isstring(obj)) {
        obj = tojson(obj)
    }
    var ret = table_add(this.id, obj)
    return ret.err
}

Table.prototype.joinkey = function(left, right) {
    return table_joinkey(left, right)
}

Table.prototype.get = function(key, row) {
    if (!isstring(row)) {
        row = tojson(row)
    }
    var ret = table_get(this.id, key, row)
    if (ret.err) {
        throwerr(ret.err)
    }
    return ret.value
}

function query_list(indexName, prefix, primaryKey, count, direction) {
    if (count !== 0 && !count) {
        count = 20
    }
    if (!direction) {
        direction = 0
    }
    if (!primaryKey) {
        primaryKey = ""
    }
    if (!prefix) {
        prefix = ""
    }
    if (!indexName) {
        indexName = ""
    }
    var q = table_query(this.id, indexName, prefix, primaryKey, count, direction)
    if (q.err) {
        return null
    }
    for (var i = 0; i < q.length; i++) {
        if (q[i].left) {
            q[i].left = JSON.parse(q[i].left)
        }
        if (q[i].right) {
            q[i].right = JSON.parse(q[i].right)
        }
    }
    return q
}

Table.prototype.query = function(indexName, prefix, primaryKey, count, direction) {
    return query_list.call(this, indexName, prefix, primaryKey, count, direction)
}

Table.prototype.replace = function(obj) {
    if (!isstring(obj)) {
        obj = tojson(obj)
    }
    var ret = table_replace(this.id, obj)
    return ret.err
}

Table.prototype.del = function(obj) {
    if (!isstring(obj)) {
        obj = tojson(obj)
    }
    var ret = table_del(this.id, obj)
    return ret.err
}

Table.prototype.save = function() {
    var ret = table_save(this.id)
    if (!this.kvc) {
        this.kvc.save(ret)
    }
    return ret
}

Table.prototype.close = function() {
    var ret = table_close(this.id)
    return ret.err
}

function JoinTable(lefttable, righttable, index) {
    this.lefttable = lefttable
    this.righttable = righttable
    if (this.lefttable.kvc != this.righttable.kvc) {
        throw new Error("the kvc of left and right must same")
    }
    this.index = index
    var ret = new_join_table(this.lefttable.id, this.righttable.id, index)
    if (ret.err) {
        throw new Error(ret.err)
    }
    this.id = ret.id
    this.kvc = this.lefttable.kvc
}

function print(obj) {
    if (typeof obj === "string") {
        console.log(obj)
        return
    }
    console.log(tojson(obj))
}

JoinTable.prototype.save = function() {
    var ret = table_save(this.id)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret
}

JoinTable.prototype.get = function(key, row) {
    if (!isstring(row)) {
        row = tojson(row)
    }
    var ret = table_get(this.id, key, row)
    if (ret.err) {
        throwerr(ret.err)
    }
    return ret.value
}

JoinTable.prototype.query = function(indexName, prefix, primaryKey, count, direction) {
    return query_list.call(this, indexName, prefix, primaryKey, count, direction)
}

function querytojson(data) {
    if (!data) {
        return "[]"
    }
    return tojson(data)
}

JoinTable.prototype.close = function() {
    table_close(this.lefttable.id)
    table_close(this.righttable.id)
    var ret = table_close(this.id)
    return ret.err
}

JoinTable.prototype.addlogs = function(data) {
    var err
    for (var i = 0; i < data.length; i++) {
        if (data[i].format != "json") {
            continue
        }
        var log = JSON.parse(data[i].log)
        if (log.__type__ == this.lefttable.name) {
            err = this.lefttable.replace(data[i].log)
            throwerr(err)
        }
        if (log.__type__ == this.righttable.name) {
            err = this.righttable.replace(data[i].log)
            throwerr(err)
        }
    }
}

//account warp
function account(kvc, execer, symbol) {
    this.execer = execer
    this.symbol = symbol
    this.kvc = kvc
}

account.prototype.genesisInit = function(addr, amount) {
    var ret = genesis_init(this, addr, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

account.prototype.execGenesisInit = function(execer, addr, amount) {
    var ret = genesis_init_exec(this, execer, addr, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

account.prototype.getBalance = function(addr) {
    return load_account(this, addr)
}

account.prototype.execGetBalance = function(execer, addr) {
    return get_balance(this, execer, addr)
}

//本合约转移资产，或者转移到其他合约，或者从其他合约取回资产
account.prototype.transfer = function(from, to, amount) {
    var ret = transfer(this, from, to, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

account.prototype.transferToExec = function(execer, from, amount) {
    var ret = transfer_to_exec(this, execer, from, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

account.prototype.withdrawFromExec = function(execer, to, amount) {
    var ret = withdraw(this, execer, to, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

//管理其他合约的资产转移到这个合约中
account.prototype.execActive = function(execer, addr, amount) {
    var ret = exec_active(this, execer, addr, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

account.prototype.execFrozen = function(execer, addr, amount) {
    var ret = exec_frozen(this, execer, addr, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

account.prototype.execDeposit = function(execer, addr, amount) {
    var ret = exec_deposit(this, execer, addr, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

account.prototype.execWithdraw = function(execer, addr, amount) {
    var ret = exec_withdraw(this, execer, addr, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

account.prototype.execTransfer = function(execer, from, to, amount) {
    var ret = exec_transfer(this, execer, from, to, amount)
    if (this.kvc) {
        this.kvc.save(ret)
    }
    return ret.err
}

//from frozen -> to active
account.prototype.execTransFrozenToActive = function(execer, from, to, amount) {
    var err
    err = this.execActive(execer, from, amount)
    if (err) {
        return err
    }
    err = this.execTransfer(execer, from, to, amount)
}

//from frozen -> to frozen
account.prototype.execTransFrozenToFrozen = function(execer, from, to, amount) {
    var err
    err = this.execActive(execer, from, amount)
    if (err) {
        return err
    }
    err = this.execTransfer(execer, from, to, amount)
    if (err) {
        return err
    }
    return this.execFrozen(execer, to, amount)
}

account.prototype.execTransActiveToFrozen = function(execer, from, to, amount) {
    var err
    err = this.execTransfer(execer, from, to, amount)
    if (err) {
        return err
    }
    return this.execFrozen(execer, to, amount)
}

COINS = 100000000

function kvcreator(dbtype) {
    this.data = {}
    this.kvs = []
	this.logs = []
	this.type = dbtype
	this.getstate = getstatedb
	this.getloal = getlocaldb
	this.list = listdb
    if (dbtype == "exec" || dbtype == "init") {
        this.getdb = this.getstate
    } else if (dbtype == "local") {
        this.getdb = this.getlocal
    } else if (dbtype == "query") {
		this.getdb = this.getlocal
	} else {
		throw new Error("chain33.js: dbtype error")
	}
}

kvcreator.prototype.add = function(k, v, prefix) {
	if (typeof v != "string") {
		v = JSON.stringify(v) 
	}
    this.data[k] = v
    this.kvs.push({key:k, value: v, prefix: !!prefix})
}

kvcreator.prototype.get = function(k, prefix) {
    var v
    if (this.data[k]) {
        v = this.data[k]
    } else {
        var dbvalue = this.getdb(k, !!prefix)
		if (dbvalue.err) {
			return null
		}
		v = dbvalue.value
    }
    if (!v) {
        return null
    }
    return JSON.parse(v)
}

kvcreator.prototype.save = function(receipt) {
    if (Array.isArray(receipt.logs)) {
        for (var i = 0; i < receipt.logs.length; i++) {
            var item = receipt.logs[i]
            this.addlog(item.log, item.ty, item.format)
        }
    }
    if (Array.isArray(receipt.kvs)) {
        for (var i = 0; i < receipt.kvs.length; i++) {
            var item = receipt.kvs[i]
            this.add(item.key, item.value, item.prefix)
        }
    }
}

kvcreator.prototype.listvalue = function(prefix, key, count, direction) {
   var dbvalues = this.list(prefix, key, count, direction)
   if (dbvalues.err) {
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

kvcreator.prototype.addlog = function(log, ty, format) {
    if (this.type != "exec") {
        throw new Error("local or query can't set log")
	}
	if (!isstring(log)) {
		log = JSON.stringify(log) 
    }
    if (!ty) {
        ty = 0
    }
    if (!format) {
        format = "json"
    }
    this.logs.push({log:log, ty: ty, format: format})
}

kvcreator.prototype.receipt = function() {
    return {kvs: this.kvs, logs: this.logs}
}

function GetExecName() {
    var exec = execname()
    throwerr(exec.err)
    return exec.value
}

function GetRandnum() {
    var n = randnum()
    throwerr(n.err)
    return n.value
}

function ExecAddress(name) {
    var addr = execaddress(name)
    if (addr.err) {
        return ""
    }
    console.log(addr.value)
    return addr.value
}

function Sha256(data) {
    var hash = sha256(data)
    if (hash.err) {
        return ""
    }
    return hash.value
}

function Exec(context) {
    this.kvc = new kvcreator("exec")
    this.context = context
    this.name = GetExecName()
    if (typeof ExecInit === "function") {
        ExecInit.call(this)
    }
}

Exec.prototype.txID = function() {
    return this.context.height * 100000 + this.context.index
}

function ExecLocal(context, logs) {
    this.kvc = new kvcreator("local")
	this.context = context
    this.logs = logs
    this.name = GetExecName()
    if (typeof ExecLocalInit === "function") {
        ExecLocalInit.call(this)
    }
}

function Query(context) {
	this.kvc = new kvcreator("query")
    this.context = context
    this.name = GetExecName()
    if (typeof QueryInit === "function") {
        QueryInit.call(this)
    }
}

Query.prototype.JoinKey = function(args) {
    return table_joinkey(args.left, args.right).value
}

function throwerr(err, msg) {
    if (err) {
        throw new Error(err + ":" + msg)
    }
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
        throw new Error("chain33.js: invalid function name not found->" + funcname)
    }
    return runobj[funcname](arg)
}
//Long
!function(t,i){"object"==typeof exports&&"object"==typeof module?module.exports=i():"function"==typeof define&&define.amd?define([],i):"object"==typeof exports?exports.Long=i():t.Long=i()}("undefined"!=typeof self?self:this,function(){return function(t){function i(h){if(n[h])return n[h].exports;var e=n[h]={i:h,l:!1,exports:{}};return t[h].call(e.exports,e,e.exports,i),e.l=!0,e.exports}var n={};return i.m=t,i.c=n,i.d=function(t,n,h){i.o(t,n)||Object.defineProperty(t,n,{configurable:!1,enumerable:!0,get:h})},i.n=function(t){var n=t&&t.__esModule?function(){return t.default}:function(){return t};return i.d(n,"a",n),n},i.o=function(t,i){return Object.prototype.hasOwnProperty.call(t,i)},i.p="",i(i.s=0)}([function(t,i){function n(t,i,n){this.low=0|t,this.high=0|i,this.unsigned=!!n}function h(t){return!0===(t&&t.__isLong__)}function e(t,i){var n,h,e;return i?(t>>>=0,(e=0<=t&&t<256)&&(h=l[t])?h:(n=r(t,(0|t)<0?-1:0,!0),e&&(l[t]=n),n)):(t|=0,(e=-128<=t&&t<128)&&(h=f[t])?h:(n=r(t,t<0?-1:0,!1),e&&(f[t]=n),n))}function s(t,i){if(isNaN(t))return i?p:m;if(i){if(t<0)return p;if(t>=c)return q}else{if(t<=-w)return _;if(t+1>=w)return E}return t<0?s(-t,i).neg():r(t%d|0,t/d|0,i)}function r(t,i,h){return new n(t,i,h)}function o(t,i,n){if(0===t.length)throw Error("empty string");if("NaN"===t||"Infinity"===t||"+Infinity"===t||"-Infinity"===t)return m;if("number"==typeof i?(n=i,i=!1):i=!!i,(n=n||10)<2||36<n)throw RangeError("radix");var h;if((h=t.indexOf("-"))>0)throw Error("interior hyphen");if(0===h)return o(t.substring(1),i,n).neg();for(var e=s(a(n,8)),r=m,u=0;u<t.length;u+=8){var g=Math.min(8,t.length-u),f=parseInt(t.substring(u,u+g),n);if(g<8){var l=s(a(n,g));r=r.mul(l).add(s(f))}else r=r.mul(e),r=r.add(s(f))}return r.unsigned=i,r}function u(t,i){return"number"==typeof t?s(t,i):"string"==typeof t?o(t,i):r(t.low,t.high,"boolean"==typeof i?i:t.unsigned)}t.exports=n;var g=null;try{g=new WebAssembly.Instance(new WebAssembly.Module(new Uint8Array([0,97,115,109,1,0,0,0,1,13,2,96,0,1,127,96,4,127,127,127,127,1,127,3,7,6,0,1,1,1,1,1,6,6,1,127,1,65,0,11,7,50,6,3,109,117,108,0,1,5,100,105,118,95,115,0,2,5,100,105,118,95,117,0,3,5,114,101,109,95,115,0,4,5,114,101,109,95,117,0,5,8,103,101,116,95,104,105,103,104,0,0,10,191,1,6,4,0,35,0,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,126,34,4,66,32,135,167,36,0,32,4,167,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,127,34,4,66,32,135,167,36,0,32,4,167,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,128,34,4,66,32,135,167,36,0,32,4,167,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,129,34,4,66,32,135,167,36,0,32,4,167,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,130,34,4,66,32,135,167,36,0,32,4,167,11])),{}).exports}catch(t){}n.prototype.__isLong__,Object.defineProperty(n.prototype,"__isLong__",{value:!0}),n.isLong=h;var f={},l={};n.fromInt=e,n.fromNumber=s,n.fromBits=r;var a=Math.pow;n.fromString=o,n.fromValue=u;var d=4294967296,c=d*d,w=c/2,v=e(1<<24),m=e(0);n.ZERO=m;var p=e(0,!0);n.UZERO=p;var y=e(1);n.ONE=y;var b=e(1,!0);n.UONE=b;var N=e(-1);n.NEG_ONE=N;var E=r(-1,2147483647,!1);n.MAX_VALUE=E;var q=r(-1,-1,!0);n.MAX_UNSIGNED_VALUE=q;var _=r(0,-2147483648,!1);n.MIN_VALUE=_;var B=n.prototype;B.toInt=function(){return this.unsigned?this.low>>>0:this.low},B.toNumber=function(){return this.unsigned?(this.high>>>0)*d+(this.low>>>0):this.high*d+(this.low>>>0)},B.toString=function(t){if((t=t||10)<2||36<t)throw RangeError("radix");if(this.isZero())return"0";if(this.isNegative()){if(this.eq(_)){var i=s(t),n=this.div(i),h=n.mul(i).sub(this);return n.toString(t)+h.toInt().toString(t)}return"-"+this.neg().toString(t)}for(var e=s(a(t,6),this.unsigned),r=this,o="";;){var u=r.div(e),g=r.sub(u.mul(e)).toInt()>>>0,f=g.toString(t);if(r=u,r.isZero())return f+o;for(;f.length<6;)f="0"+f;o=""+f+o}},B.getHighBits=function(){return this.high},B.getHighBitsUnsigned=function(){return this.high>>>0},B.getLowBits=function(){return this.low},B.getLowBitsUnsigned=function(){return this.low>>>0},B.getNumBitsAbs=function(){if(this.isNegative())return this.eq(_)?64:this.neg().getNumBitsAbs();for(var t=0!=this.high?this.high:this.low,i=31;i>0&&0==(t&1<<i);i--);return 0!=this.high?i+33:i+1},B.isZero=function(){return 0===this.high&&0===this.low},B.eqz=B.isZero,B.isNegative=function(){return!this.unsigned&&this.high<0},B.isPositive=function(){return this.unsigned||this.high>=0},B.isOdd=function(){return 1==(1&this.low)},B.isEven=function(){return 0==(1&this.low)},B.equals=function(t){return h(t)||(t=u(t)),(this.unsigned===t.unsigned||this.high>>>31!=1||t.high>>>31!=1)&&(this.high===t.high&&this.low===t.low)},B.eq=B.equals,B.notEquals=function(t){return!this.eq(t)},B.neq=B.notEquals,B.ne=B.notEquals,B.lessThan=function(t){return this.comp(t)<0},B.lt=B.lessThan,B.lessThanOrEqual=function(t){return this.comp(t)<=0},B.lte=B.lessThanOrEqual,B.le=B.lessThanOrEqual,B.greaterThan=function(t){return this.comp(t)>0},B.gt=B.greaterThan,B.greaterThanOrEqual=function(t){return this.comp(t)>=0},B.gte=B.greaterThanOrEqual,B.ge=B.greaterThanOrEqual,B.compare=function(t){if(h(t)||(t=u(t)),this.eq(t))return 0;var i=this.isNegative(),n=t.isNegative();return i&&!n?-1:!i&&n?1:this.unsigned?t.high>>>0>this.high>>>0||t.high===this.high&&t.low>>>0>this.low>>>0?-1:1:this.sub(t).isNegative()?-1:1},B.comp=B.compare,B.negate=function(){return!this.unsigned&&this.eq(_)?_:this.not().add(y)},B.neg=B.negate,B.add=function(t){h(t)||(t=u(t));var i=this.high>>>16,n=65535&this.high,e=this.low>>>16,s=65535&this.low,o=t.high>>>16,g=65535&t.high,f=t.low>>>16,l=65535&t.low,a=0,d=0,c=0,w=0;return w+=s+l,c+=w>>>16,w&=65535,c+=e+f,d+=c>>>16,c&=65535,d+=n+g,a+=d>>>16,d&=65535,a+=i+o,a&=65535,r(c<<16|w,a<<16|d,this.unsigned)},B.subtract=function(t){return h(t)||(t=u(t)),this.add(t.neg())},B.sub=B.subtract,B.multiply=function(t){if(this.isZero())return m;if(h(t)||(t=u(t)),g){return r(g.mul(this.low,this.high,t.low,t.high),g.get_high(),this.unsigned)}if(t.isZero())return m;if(this.eq(_))return t.isOdd()?_:m;if(t.eq(_))return this.isOdd()?_:m;if(this.isNegative())return t.isNegative()?this.neg().mul(t.neg()):this.neg().mul(t).neg();if(t.isNegative())return this.mul(t.neg()).neg();if(this.lt(v)&&t.lt(v))return s(this.toNumber()*t.toNumber(),this.unsigned);var i=this.high>>>16,n=65535&this.high,e=this.low>>>16,o=65535&this.low,f=t.high>>>16,l=65535&t.high,a=t.low>>>16,d=65535&t.low,c=0,w=0,p=0,y=0;return y+=o*d,p+=y>>>16,y&=65535,p+=e*d,w+=p>>>16,p&=65535,p+=o*a,w+=p>>>16,p&=65535,w+=n*d,c+=w>>>16,w&=65535,w+=e*a,c+=w>>>16,w&=65535,w+=o*l,c+=w>>>16,w&=65535,c+=i*d+n*a+e*l+o*f,c&=65535,r(p<<16|y,c<<16|w,this.unsigned)},B.mul=B.multiply,B.divide=function(t){if(h(t)||(t=u(t)),t.isZero())throw Error("division by zero");if(g){if(!this.unsigned&&-2147483648===this.high&&-1===t.low&&-1===t.high)return this;return r((this.unsigned?g.div_u:g.div_s)(this.low,this.high,t.low,t.high),g.get_high(),this.unsigned)}if(this.isZero())return this.unsigned?p:m;var i,n,e;if(this.unsigned){if(t.unsigned||(t=t.toUnsigned()),t.gt(this))return p;if(t.gt(this.shru(1)))return b;e=p}else{if(this.eq(_)){if(t.eq(y)||t.eq(N))return _;if(t.eq(_))return y;return i=this.shr(1).div(t).shl(1),i.eq(m)?t.isNegative()?y:N:(n=this.sub(t.mul(i)),e=i.add(n.div(t)))}if(t.eq(_))return this.unsigned?p:m;if(this.isNegative())return t.isNegative()?this.neg().div(t.neg()):this.neg().div(t).neg();if(t.isNegative())return this.div(t.neg()).neg();e=m}for(n=this;n.gte(t);){i=Math.max(1,Math.floor(n.toNumber()/t.toNumber()));for(var o=Math.ceil(Math.log(i)/Math.LN2),f=o<=48?1:a(2,o-48),l=s(i),d=l.mul(t);d.isNegative()||d.gt(n);)i-=f,l=s(i,this.unsigned),d=l.mul(t);l.isZero()&&(l=y),e=e.add(l),n=n.sub(d)}return e},B.div=B.divide,B.modulo=function(t){if(h(t)||(t=u(t)),g){return r((this.unsigned?g.rem_u:g.rem_s)(this.low,this.high,t.low,t.high),g.get_high(),this.unsigned)}return this.sub(this.div(t).mul(t))},B.mod=B.modulo,B.rem=B.modulo,B.not=function(){return r(~this.low,~this.high,this.unsigned)},B.and=function(t){return h(t)||(t=u(t)),r(this.low&t.low,this.high&t.high,this.unsigned)},B.or=function(t){return h(t)||(t=u(t)),r(this.low|t.low,this.high|t.high,this.unsigned)},B.xor=function(t){return h(t)||(t=u(t)),r(this.low^t.low,this.high^t.high,this.unsigned)},B.shiftLeft=function(t){return h(t)&&(t=t.toInt()),0==(t&=63)?this:t<32?r(this.low<<t,this.high<<t|this.low>>>32-t,this.unsigned):r(0,this.low<<t-32,this.unsigned)},B.shl=B.shiftLeft,B.shiftRight=function(t){return h(t)&&(t=t.toInt()),0==(t&=63)?this:t<32?r(this.low>>>t|this.high<<32-t,this.high>>t,this.unsigned):r(this.high>>t-32,this.high>=0?0:-1,this.unsigned)},B.shr=B.shiftRight,B.shiftRightUnsigned=function(t){return h(t)&&(t=t.toInt()),0==(t&=63)?this:t<32?r(this.low>>>t|this.high<<32-t,this.high>>>t,this.unsigned):32===t?r(this.high,0,this.unsigned):r(this.high>>>t-32,0,this.unsigned)},B.shru=B.shiftRightUnsigned,B.shr_u=B.shiftRightUnsigned,B.rotateLeft=function(t){var i;return h(t)&&(t=t.toInt()),0==(t&=63)?this:32===t?r(this.high,this.low,this.unsigned):t<32?(i=32-t,r(this.low<<t|this.high>>>i,this.high<<t|this.low>>>i,this.unsigned)):(t-=32,i=32-t,r(this.high<<t|this.low>>>i,this.low<<t|this.high>>>i,this.unsigned))},B.rotl=B.rotateLeft,B.rotateRight=function(t){var i;return h(t)&&(t=t.toInt()),0==(t&=63)?this:32===t?r(this.high,this.low,this.unsigned):t<32?(i=32-t,r(this.high<<i|this.low>>>t,this.low<<i|this.high>>>t,this.unsigned)):(t-=32,i=32-t,r(this.low<<i|this.high>>>t,this.high<<i|this.low>>>t,this.unsigned))},B.rotr=B.rotateRight,B.toSigned=function(){return this.unsigned?r(this.low,this.high,!1):this},B.toUnsigned=function(){return this.unsigned?this:r(this.low,this.high,!0)},B.toBytes=function(t){return t?this.toBytesLE():this.toBytesBE()},B.toBytesLE=function(){var t=this.high,i=this.low;return[255&i,i>>>8&255,i>>>16&255,i>>>24,255&t,t>>>8&255,t>>>16&255,t>>>24]},B.toBytesBE=function(){var t=this.high,i=this.low;return[t>>>24,t>>>16&255,t>>>8&255,255&t,i>>>24,i>>>16&255,i>>>8&255,255&i]},n.fromBytes=function(t,i,h){return h?n.fromBytesLE(t,i):n.fromBytesBE(t,i)},n.fromBytesLE=function(t,i){return new n(t[0]|t[1]<<8|t[2]<<16|t[3]<<24,t[4]|t[5]<<8|t[6]<<16|t[7]<<24,i)},n.fromBytesBE=function(t,i){return new n(t[4]<<24|t[5]<<16|t[6]<<8|t[7],t[0]<<24|t[1]<<16|t[2]<<8|t[3],i)}}])});
`
var jscode = `
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
`
var _ = jscode
var gamecode = `
//简单的猜数字游戏
//游戏规则: 庄家出一个 0 - 10 的数字 hash(随机数 + 9) (一共的赔偿金额) NewGame()
//用户可以猜这个数字，多个用户都可以猜测。 Guess()
//开奖 CloseGame()
function Init(context) {
    this.kvc = new kvcreator("init")
    this.context = context
    return this.kvc.receipt()
}

var MIN_WAIT_BLOCK = 2
var RAND_MAX = 10

function ExecInit() {
    this.acc = new account(this.kvc, "coins", "bty")
}

Exec.prototype.NewGame = function(args) {
    var game = {__type__ : "game"}
    game.gameid = this.txID()
    game.height = this.context.height
    game.randhash = args.randhash
    game.bet = args.bet
    game.hash = this.context.txhash
    game.obet = game.bet
    game.addr = this.context.from
    game.status = 1 //open
    //最大值是 9000万,否则js到 int 会溢出
    if (game.bet < 10 * COINS || game.bet > 10000000 * COINS) {
        throwerr("bet low than 10 or hight than 10000000")
    }
    if (this.kvc.get(game.randhash)) { //如果randhash 已经被使用了
        throwerr("dup rand hash")
    }
    var err = this.acc.execFrozen(this.name, this.context.from, game.bet)
    throwerr(err)
    this.kvc.add(game.gameid, game)
    this.kvc.add(game.randhash, "ok")
    this.kvc.addlog(game)
    return this.kvc.receipt()
}

Exec.prototype.Guess = function(args) {
    var match = {__type__ : "match"}
    match.gameid = args.gameid
    match.bet = args.bet
    match.id = this.txID()
    match.addr = this.context.from
    match.hash = this.context.txhash
    match.num = args.num
    var game = this.kvc.get(match.gameid)
    if (!game) {
        throwerr("guess: game id not found")
    }
    if (game.status != 1) {
        throwerr("guess: game status not open")
    }
    if (this.context.from == game.addr) {
        throwerr("guess: game addr and match addr is same")
    }
    if (match.bet < 1 * COINS || match.bet > game.bet / RAND_MAX) {
        throwerr("match bet litte than 1 or big than game.bet/10")
    }
    var err = this.acc.execFrozen(this.name, this.context.from, match.bet)
    console.log(this.name, this.context.from, err)
    throwerr(err)
    this.kvc.add(match.id, match)
    this.kvc.addlog(match)
    return this.kvc.receipt()
}

Exec.prototype.CloseGame = function(args) {
    var local = MatchLocalTable(this.kvc)
    var game = this.kvc.get(args.gameid)
    if (!game) {
        throwerr("game id not found")
    }
    var querykey = local.get("gameid", args)
    var matches = local.query("gameid", querykey, "", 0, 1)
    if (!matches) {
        matches = []
    }
    var n = -1
    for (var i = 0; i < RAND_MAX; i++) {
        if (Sha256(args.randstr + i) == game.randhash) {
            n = i
        }
    }
    if (n == -1) {
        throwerr("err rand str")
    }
    //必须可以让用户可以有一个区块的竞猜时间
    if (this.context.height - game.height < MIN_WAIT_BLOCK) {
        throwerr("close game must wait "+MIN_WAIT_BLOCK+" block")
    }
    for (var i = 0; i < matches.length; i++) {
        var match = matches[i].left
        if (match.num == n) {
            //不能随便添加辅助函数，因为可以被外界调用到，所以辅助函数都是传递 this
            win.call(this, game, match)
        } else {
            fail.call(this, game, match)
        }
    }
    if (game.bet > 0) {
        var err = this.acc.execActive(this.name, game.addr, game.bet)
        throwerr(err)
        game.bet = 0
    }
    game.status = 2
    this.kvc.add(game.gameid, game)
    this.kvc.addlog(game)
    return this.kvc.receipt()
}

function win(game, match) {
    var amount = (RAND_MAX - 1) * match.bet
    if (game.bet - amount < 0) {
        amount = game.bet
    }
    var err 
    if (amount > 0) {
        err = this.acc.execTransFrozenToActive(this.name, game.addr, match.addr, amount)
        throwerr(err, "execTransFrozenToActive")
        game.bet -= amount
    }
    err = this.acc.execActive(this.name, match.addr, match.bet)
    throwerr(err, "execActive")
}

function fail(game, match) {
    var amount = match.bet
    err = this.acc.execTransFrozenToFrozen(this.name, match.addr, game.addr, amount)
    throwerr(err)
    game.bet += amount
}

Exec.prototype.ForceCloseGame = function(args) {
    var local = new MatchLocalTable(this.kvc)
    var game = this.kvc.get(args.id)
    if (!game) {
        throwerr("game id not found")
    }
    var matches = local.getmath(args.id)
    if (!matches) {
        matches = []
    }
    if (this.context.height - game.height < 100) {
        throwerr("force close game must wait 100 block")
    }
    for (var i = 0; i < matches.length; i++) {
        var match = matches[i]
        win.call(this.kvc, game, match)
    }
    if (game.bet > 0) {
        var err = this.acc.execActive(this.name, game.addr, game.bet)
        throwerr(err)
        game.bet = 0
    }
    game.status = 2
    this.kvc.add(game.gameid, game)
    this.kvc.addlog(game)
    return this.kvc.receipt()
}

ExecLocal.prototype.NewGame = function(args) {
    return localprocess.call(this, args)
}

ExecLocal.prototype.Guess = function(args) {
    return localprocess.call(this, args)
}

ExecLocal.prototype.CloseGame = function(args) {
    return localprocess.call(this, args)
}

ExecLocal.prototype.ForceCloseGame = function(args) {
    return localprocess.call(this, args)
}

function localprocess(args) {
    var local = MatchGameTable(this.kvc)
    local.addlogs(this.logs)
    local.save()
    return this.kvc.receipt()
}

Query.prototype.ListGameByAddr = function(args) {
    var local = GameLocalTable(this.kvc)
    var q = local.query("addr", args.addr, args.primaryKey, args.count, args.direction)
    return querytojson(q)
}

Query.prototype.ListMatchByAddr = function(args) {
    var local = MatchGameTable(this.kvc)
    var q= local.query("addr#status", args["addr#status"], args.primaryKey, args.count, args.direction)
    return querytojson(q)
}

function GameLocalTable(kvc) {
    this.config = {
        "#tablename" : "game",
        "#primary" : "gameid",
        "#db" : "localdb",
        "gameid"    : "%018d",
        "status" : "%d",
        "hash" : "%s",
        "addr" : "%s",
    }
    this.defaultvalue = {
        "gameid" : 0,
        "status" : 0,
        "hash" : "",
        "addr" : "",
    }
    return new Table(kvc, this.config, this.defaultvalue) 
}

function MatchLocalTable(kvc) {
    this.config = {
        "#tablename" : "match",
        "#primary" : "id",
        "#db" : "localdb",
        "id"    : "%018d",
        "gameid" : "%018d",
        "hash" : "%s",
        "addr" : "%s",
    }
    this.defaultvalue = {
        "id" : 0,
        "gameid" : 0,
        "hash" : "",
        "addr" : "",
    }
    return new Table(kvc, this.config, this.defaultvalue)  
}

function MatchGameTable(kvc) {
    return new JoinTable(MatchLocalTable(kvc), GameLocalTable(kvc), "addr#status")
}`
var _ = gamecode
