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

function ExecInit(execthis) {
    execthis.acc = new account(this.kvc, "coins", "bty")
}

Exec.prototype.NewGame = function(args) {
    var game = {__type__ : "game"}
    game.id = this.context.txhash
    game.index = this.txID()
    game.height = this.context.height
    game.randhash = args.hash
    game.bet = args.bet
    game.obet = game.bet
    game.addr = this.context.from
    game.status = 1 //open
    //最大值是 9000万,否则js到 int 会溢出
    if (game.bet < 10 * COINS || game.bet > 10000000 * COINS) {
        throwerr("bet low than 10 or hight than 10000000")
    }
    var err = this.acc.execFrozen(this.name, this.context.from, game.bet)
    throwerr(err)
    this.kvc.add(game.id, game)
    this.kvc.addlog(game)
    return this.kvc.receipt()
}

Exec.prototype.Guess = function(args) {
    var match = {__type__ : "match"}
    match.gameid = args.gameid
    match.bet = args.bet
    match.id = this.txID()
    match.addr = this.context.from
    var game = this.kvc.get(match.gameid)
    if (!game) {
        throwerr("game id not found")
    }
    if (game.status != 1) {
        throwerr("game status not open")
    }
    if (match.bet < 1 * COINS || match.bet > game.bet / RAND_MAX) {
        throwerr("match bet litte than 1 or big than game.bet/10")
    }
    var err = this.acc.execFrozen(this.name, this.context.from, game.bet)
    throwerr(err)
    this.kvc.add(match.id, match)
    this.kvc.addlog(match)
    return this.kvc.receipt()
}

Exec.prototype.CloseGame = function(args) {
    var local = new MatchLocalTable(this.kvc)
    var game = this.kvc.get(args.id)
    if (!game) {
        throwerr("game id not found")
    }
    var matches = local.getmath(args.id)
    if (!matches) {
        matches = []
    }
    var n = -1
    for (var i = 0; i < RAND_MAX; i ++) {
        if (sha256(args.randstr + i) == game.randhash) {
            n = i
        }
    }
    if (n == -1) {
        throwerr("err rand str")
    }
    //必须可以让用户可以有一个区块的竞猜时间
    if (this.context.height - game.height < MIN_WAIT_BLOCK) {
        throwerr("close game must wait 2 block")
    }
    for (var i = 0; i < matches.length; i++) {
        var match = matches[i]
        if (match.num == n) {
            //不能随便添加辅助函数，因为可以被外界调用到，所以辅助函数都是传递 this
            win(this, game, match)
        } else {
            fail(this, game, match)
        }
    }
    if (game.bet > 0) {
        var err = this.acc.execActive(this.name, game.addr, game.bet)
        throwerr(err)
        game.bet = 0
    }
    game.status = 2
    this.kvc.add(game.id, game)
    this.kvc.addlog(game)
    return this.kvc.receipt()
}

function win(this, game, match) {
    var amount = (RAND_MAX - 1) * match.bet
    if (game.bet - amount < 0) {
        amount = game.bet
    }
    var err 
    if (amount > 0) {
        err = this.acc.execTransFrozenToActive(this.name, game.addr, match.addr, amount)
        throwerr(err)
        game.bet -= amount
    }
    err = this.acc.execActive(match.addr, match.bet)
    throwerr(err)
}

function fail(this, game, match) {
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
        win(this.kvc, game, match)
    }
    if (game.bet > 0) {
        var err = this.acc.execActive(this.name, game.addr, game.bet)
        throwerr(err)
        game.bet = 0
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

match.gameid -> fk
match.id -> primary
*/
function GameLocalTable(kvc) {
    this.config = {
        "#tablename" : "game",
        "#primary" : "id",
        "#db" : "localdb",
        "id"    : "%018d",
        "status" : "%d",
        "addr" : "%s",
    }
    this.defaultvalue = {
        "id" : "0",
        "status" : 0,
        "addr" : "",
    }
    this.kvc = kvc
    this.table = new Table(this.kvc, this.config, this.defaultvalue) 
}

function MatchLocalTable(kvc) {
    this.config = {
        "#tablename" : "match",
        "#primary" : "id",
        "#db" : "localdb",
        "id"    : "%018d",
        "gameid" : "%s",
        "addr" : "%s",
    }
    this.defaultvalue = {
        "id" : 0,
        "gameid" : 0,
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