package executor_test

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
