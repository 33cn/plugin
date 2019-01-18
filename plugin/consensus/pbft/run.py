#!/usr/bin/env python3.5

from subprocess import run
from subprocess import PIPE
from time import sleep
from json import loads
from json import dumps
from requests import post

def RUN(data, log=True):
    if log:
        print('>>>>> >>>>>')
        print(data)
        print('===== =====')

    ret = post('http://127.0.0.1:5005', json=data)
    if ret.ok == False:
        raise Exception('REQ not OK')
    ret = ret.json()

    if log:
        print(ret)
        print('<<<<< <<<<<')
    return ret

run('docker image inspect 33-pbft:latest >/dev/null 2>&1 || \
    docker build -t go-pbft .', shell=True)

run('docker network inspect PBFT >/dev/null 2>&1 || \
    docker network create \
    --driver=bridge --subnet=172.28.0.0/16 --ip-range=172.28.5.0/24 --gateway=172.28.5.254 PBFT', \
    shell=True)

run('docker run -dit \
    --net PBFT --ip 172.28.0.6 --name replica-1 -p 127.0.0.1:5001:8801 --rm 33-pbft pbft1.toml', shell=True)
run('docker run -dit \
    --net PBFT --ip 172.28.0.7 --name replica-2 -p 127.0.0.1:5002:8801 --rm 33-pbft pbft2.toml', shell=True)
run('docker run -dit \
    --net PBFT --ip 172.28.0.8 --name replica-3 -p 127.0.0.1:5003:8801 --rm 33-pbft pbft3.toml', shell=True)
run('docker run -dit \
    --net PBFT --ip 172.28.0.9 --name replica-4 -p 127.0.0.1:5004:8801 --rm 33-pbft pbft4.toml', shell=True)
run('docker run -dit \
    --net PBFT --ip 172.28.0.10 --name client -p 127.0.0.1:5005:8801 --rm 33-pbft pbftc.toml', shell=True)

sleep(10)


# wait for Sync
for i in range(16):
    ret = RUN({
        "jsonrpc": "2.0",
        "id": 0,
        "method": "Chain33.IsSync",
        "params": [],
    })
    ret = ret['result']
    if ret:
        break
    sleep(2**i)


ret = RUN({
    "jsonrpc": "2.0",
    "id": 0,
    "method": "Chain33.GetPeerInfo",
    "params": [],
})

assert ret['error'] == None

ret = RUN({
    "jsonrpc": "2.0",
    "id": 0,
    "method": "Chain33.IsSync",
    "params": [],
})

assert ret['error'] == None

ret = RUN({
    "jsonrpc": "2.0",
    "id": 0,
    "method": "Chain33.GenSeed",
    "params": [{
        "lang": 0,
    }],
})

assert ret['error'] == None

ret = ret['result']
ret = ret['seed']

ret = RUN({
    "jsonrpc": "2.0",
    "id": 0,
    "method": "Chain33.SaveSeed",
    "params": [{
        "seed": ret,
        "passwd": "pwd",
    }],
})

assert ret['error'] == None
assert ret['result']['isOK'] == True

RUN({
    "jsonrpc": "2.0",
    "id": 0,
    "method": "Chain33.UnLock",
    "params": [{
        "passwd": "pwd",
        "walletorticket": False,
    }],
})

assert ret['error'] == None
assert ret['result']['isOK'] == True

ret = RUN({
    "jsonrpc": "2.0",
    "id": 0,
    "method": "Chain33.ImportPrivkey",
    "params": [{
        "privkey": "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944",
        "label": "origin",
    }],
})

assert ret['error'] == None

ret = ret['result']
ret = ret['acc']
ret = ret['addr']

origin = ret

ret = RUN({
    "jsonrpc": "2.0",
    "id": 0,
    "method": "Chain33.NewAccount",
    "params": [{
        "label": "alex",
    }],
})

assert ret['error'] == None

ret = ret['result']
ret = ret['acc']
ret = ret['addr']

alex = ret

ret = RUN({
    "jsonrpc": "2.0",
    "id": 0,
    "method": "Chain33.NewAccount",
    "params": [{
        "label": "bob",
    }],
})

assert ret['error'] == None

ret = ret['result']
ret = ret['acc']
print(ret)
ret = ret['addr']

bob = ret

ret = RUN({
    "jsonrpc": "2.0",
    "id": 1,
    "method": "Chain33.CreateRawTransaction",
    "params": [{
        "to": alex,
        "amount": 10000000000,
        "isToken": False,
        "isWithdraw": False,
    }],
})

assert ret['error'] == None
ret = ret['result']

ret = RUN({
    "jsonrpc": "2.0",
    "id": 2,
    "method": "Chain33.SignRawTx",
    "params": [{
        "addr": origin,
        "txHex": ret,
        "expire": "1h",
        "model": 0,
    }],
})

assert ret['error'] == None
ret = ret['result']

ret = RUN({
    "jsonrpc": "2.0",
    "id": 3,
    "method": "Chain33.SendTransaction",
    "params": [{
        "data": ret,
    }],
})

assert ret['error'] == None
ret = ret['result']

for i in range(16):
    _ret = RUN({
        "jsonrpc": "2.0",
        "id": 3,
        "method": "Chain33.QueryTransaction",
        "params": [{
            "hash": ret,
        }],
    })
    if _ret['error'] == None and _ret['result']:
        break
    sleep(2**i)

ret = RUN({
    "jsonrpc": "2.0",
    "id": 1,
    "method": "Chain33.CreateRawTransaction",
    "params": [{
        "to": bob,
        "amount": 10000000000,
        "isToken": False,
        "isWithdraw": False,
    }],
})

assert ret['error'] == None
ret = ret['result']

ret = RUN({
    "jsonrpc": "2.0",
    "id": 2,
    "method": "Chain33.SignRawTx",
    "params": [{
        "addr": origin,
        "txHex": ret,
        "expire": "1h",
        "model": 0,
    }],
})

assert ret['error'] == None
ret = ret['result']

ret = RUN({
    "jsonrpc": "2.0",
    "id": 3,
    "method": "Chain33.SendTransaction",
    "params": [{
        "data": ret,
    }],
})

assert ret['error'] == None
ret = ret['result']

for i in range(16):
    _ret = RUN({
        "jsonrpc": "2.0",
        "id": 3,
        "method": "Chain33.QueryTransaction",
        "params": [{
            "hash": ret,
        }],
    })
    if _ret['error'] == None and _ret['result']:
        break
    sleep(2**i)

ret = RUN({
    "jsonrpc": "2.0",
    "id": 3,
    "method": "Chain33.GetAccounts",
    "params": [],
})

assert ret['error'] == None


for i in range(10):
    ret = RUN({
        "jsonrpc": "2.0",
        "id": 1,
        "method": "Chain33.CreateRawTransaction",
        "params": [{
            "to": bob,
            "amount": 1000000000,
            "isToken": False,
            "isWithdraw": False,
        }],
    })

    assert ret['error'] == None
    ret = ret['result']

    ret = RUN({
        "jsonrpc": "2.0",
        "id": 2,
        "method": "Chain33.SignRawTx",
        "params": [{
            "addr": origin,
            "txHex": ret,
            "expire": "1h",
            "model": 0,
        }],
    })

    assert ret['error'] == None
    ret = ret['result']

    ret = RUN({
        "jsonrpc": "2.0",
        "id": 3,
        "method": "Chain33.SendTransaction",
        "params": [{
            "data": ret,
        }],
    })

    assert ret['error'] == None
    ret = ret['result']

    for i in range(16):
        _ret = RUN({
            "jsonrpc": "2.0",
            "id": 3,
            "method": "Chain33.QueryTransaction",
            "params": [{
                "hash": ret,
            }],
        })
        if _ret['error'] == None and _ret['result']:
            break
        sleep(2**i)

ret = RUN({
    "jsonrpc": "2.0",
    "id": 3,
    "method": "Chain33.GetAccounts",
    "params": [],
})

print(origin)
print(alex)
print(bob)

run('docker stop $(docker ps -a -q)', shell=True)