//简单的猜数字游戏
//游戏规则: 庄家出一个 0 - 10 的数字 hash(随机数 + 9) (一共的赔偿金额) NewGame()
//用户可以猜这个数字，多个用户都可以猜测。 Guess()
//开奖 CloseGame()
function Init(context) {
    this.kvc = new kvcreator("init")
    this.context = context
    return this.kvc.receipt()
}

function Exec(context) {
    this.kvc = new kvcreator("exec")
	this.context = context
}

function ExecLocal(context, logs) {
    this.kvc = new kvcreator("local")
	this.context = context
    this.logs = logs
}

function Query(context) {
	this.kvc = new kvcreator("query")
	this.context = context
}

Exec.prototype.NewGame = function(args) {
    var game = {__type__ : "game"}
    game.id = this.context.txhash
    game.index = this.context.height * 100000 + this.index
    game.height = this.context.height
    game.randhash = args.hash
    game.bet = args.bet
    game.status = 1 //open
    if (game.bet < 10) {
        throw new Error("bet too low")
    }
    this.kvc.add(game.id, game)
    this.kvc.addlog(game)
	return this.kvc.receipt()
}

Exec.prototype.Guess = function(args) {
    var match = {__type__ : "match"}
    match.gameid = args.gameid
    match.bet = args.bet
    match.id = this.context.txhash
    match.index = this.context.height * 100000 + this.index
    match.addr = this.context.from
    var game = this.kvc.get(match.gameid)
    if (!game) {
        throw new Error("game id not found")
    }
    if (game.status != 1) {
        throw new Error("game status not open")
    }
    this.kvc.add(match.id, match)
    this.kvc.addlog(match)
	return this.kvc.receipt()
}

Exec.prototype.CloseGame = function(args) {
    var local = new MatchLocalTable(this.kvc)
    var game = this.kvc.get(args.id)
    if (!game) {
        throw new Error("game id not found")
    }
    var matches = local.getmath(args.id)
    if (!matches) {
        matches = []
    }
    var n = -1
    for (var i = 0; i < 10; i ++) {
        if (sha256(args.randstr + i) == game.randhash) {
            n = i
        }
    }
    if (n == -1) {
        throw new Error("err rand str")
    }
    if (this.context.height - game.height < 10) {
        throw new Error("close game must wait 10 block")
    }
    for (var i = 0; i < matches.length; i++) {
        var match = matches[i]
        if (match.num == n) {
            win(this.kvc, game, match)
        } else {
            fail(this.kvc, game, match)
        }
    }
    game.status = 2
    this.kvc.add(game.id, game)
    this.kvc.addlog(game)
	return this.kvc.receipt()
}

Exec.prototype.ForceCloseGame = function(args) {
    var local = new MatchLocalTable(this.kvc)
    var game = this.kvc.get(args.id)
    if (!game) {
        throw new Error("game id not found")
    }
    var matches = local.getmath(args.id)
    if (!matches) {
        matches = []
    }
    if (this.context.height - game.height < 100) {
        throw new Error("force close game must wait 100 block")
    }
    for (var i = 0; i < matches.length; i++) {
        var match = matches[i]
        win(this.kvc, game, match)
    }
    game.status = 2
    this.kvc.add(game.id, game)
    this.kvc.addlog(game)
	return this.kvc.receipt()
}

ExecLocal.prototype.NewGame = function(args) {
    var local = new MatchGameTable(this.kvc)
    local.add(this.logs)
    local.table.save()
	return this.kvc.receipt()
}

ExecLocal.prototype.Guess = function(args) {
    var local = new MatchGameTable(this.kvc)
    local.add(this.logs)
    local.table.save()
	return this.kvc.receipt()
}

ExecLocal.prototype.CloseGame = function(args) {
    var local = new MatchGameTable(this.kvc)
    local.add(this.logs)
    local.table.save()
	return this.kvc.receipt()
}

ExecLocal.prototype.ForceCloseGame = function(args) {
    var local = new GameLocalTable(this.kvc)
    local.add(this.logs)
    local.table.save()
    return this.kvc.receipt()
}

Query.prototype.ListGameByAddr = function(args) {
    var local = new GameLocalTable(this.kvc)
    return local.query(args)
}

/*
game ->(1 : n) match
game.id -> primary
game.index -> index

match.gameid -> fk
match.id -> primary
match.index -> index
*/

function GameLocalTable(kvc) {
    this.config = {
        "#tablename" : "game",
        "#primary" : "id",
        "#db" : "localdb",
        "id"    : "%s",
        "index" : "%18d",
        "status" : "%d",
    }
    this.defaultvalue = {
        "id" : "0",
        "index" : 0,
        "status" : 0,
    }
    this.kvc = kvc
    this.table = new Table(this.kvc, this.config, this.defaultvalue) 
}

function MatchLocalTable(kvc) {
    this.config = {
        "#tablename" : "match",
        "#primary" : "id",
        "#db" : "localdb",
        "id"    : "%s",
        "gameid" : "%s",
        "index" : "%18d",
        "addr" : "%s",
    }
    this.defaultvalue = {
        "id" : "0",
        "index" : 0,
        "gameid" : "0",
        "addr" : "",
    }
    this.kvc = kvc
    this.table = new Table(this.kvc, this.config, this.defaultvalue)  
}

function MatchGameTable(kvc) {
    this.left = MatchLocalTable(kvc)
    this.right = GameLocalTable(kvc)
    this.table = new JoinTable(left, right, "addr#status")
}

MatchGameTable.prototype.add = function(data) {
    if (data.__type__ == "match") {
        this.left.table.replace(data)
    }
    if (data.__type__ == "game") {
        this.right.table.replace(data)
    }
}