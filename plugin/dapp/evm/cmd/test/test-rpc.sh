#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh

set -x
set +e

# shellcheck disable=SC2089
erc20_abi='[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"supply\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]'
erc20_code="0x608060405234801561001057600080fd5b50604051610bd1380380610bd18339818101604052608081101561003357600080fd5b810190808051604051939291908464010000000082111561005357600080fd5b90830190602082018581111561006857600080fd5b825164010000000081118282018810171561008257600080fd5b82525081516020918201929091019080838360005b838110156100af578181015183820152602001610097565b50505050905090810190601f1680156100dc5780820380516001836020036101000a031916815260200191505b50604052602001805160405193929190846401000000008211156100ff57600080fd5b90830190602082018581111561011457600080fd5b825164010000000081118282018810171561012e57600080fd5b82525081516020918201929091019080838360005b8381101561015b578181015183820152602001610143565b50505050905090810190601f1680156101885780820380516001836020036101000a031916815260200191505b5060409081526020828101519290910151865192945092506101af916003918701906101e9565b5082516101c39060049060208601906101e9565b5060028290556001600160a01b03166000908152602081905260409020555061027c9050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061022a57805160ff1916838001178555610257565b82800160010185558215610257579182015b8281111561025757825182559160200191906001019061023c565b50610263929150610267565b5090565b5b808211156102635760008155600101610268565b6109468061028b6000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c8063395093511161007157806339509351146101d957806370a082311461020557806395d89b411461022b578063a457c2d714610233578063a9059cbb1461025f578063dd62ed3e1461028b576100a9565b806306fdde03146100ae578063095ea7b31461012b57806318160ddd1461016b57806323b872dd14610185578063313ce567146101bb575b600080fd5b6100b66102b9565b6040805160208082528351818301528351919283929083019185019080838360005b838110156100f05781810151838201526020016100d8565b50505050905090810190601f16801561011d5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6101576004803603604081101561014157600080fd5b506001600160a01b03813516906020013561034f565b604080519115158252519081900360200190f35b61017361036c565b60408051918252519081900360200190f35b6101576004803603606081101561019b57600080fd5b506001600160a01b03813581169160208101359091169060400135610372565b6101c3610421565b6040805160ff9092168252519081900360200190f35b610157600480360360408110156101ef57600080fd5b506001600160a01b038135169060200135610426565b6101736004803603602081101561021b57600080fd5b50356001600160a01b0316610471565b6100b661048c565b6101576004803603604081101561024957600080fd5b506001600160a01b0381351690602001356104ed565b6101576004803603604081101561027557600080fd5b506001600160a01b038135169060200135610585565b610173600480360360408110156102a157600080fd5b506001600160a01b0381358116916020013516610599565b60038054604080516020601f60026000196101006001881615020190951694909404938401819004810282018101909252828152606093909290918301828280156103455780601f1061031a57610100808354040283529160200191610345565b820191906000526020600020905b81548152906001019060200180831161032857829003601f168201915b5050505050905090565b600061036361035c6105c4565b84846105c8565b50600192915050565b60025490565b600061037f8484846106b4565b6001600160a01b0384166000908152600160205260408120816103a06105c4565b6001600160a01b03166001600160a01b03168152602001908152602001600020549050828110156104025760405162461bcd60e51b815260040180806020018281038252602881526020018061087b6028913960400191505060405180910390fd5b6104168561040e6105c4565b8584036105c8565b506001949350505050565b600890565b60006103636104336105c4565b8484600160006104416105c4565b6001600160a01b03908116825260208083019390935260409182016000908120918b1681529252902054016105c8565b6001600160a01b031660009081526020819052604090205490565b60048054604080516020601f60026000196101006001881615020190951694909404938401819004810282018101909252828152606093909290918301828280156103455780601f1061031a57610100808354040283529160200191610345565b600080600160006104fc6105c4565b6001600160a01b03908116825260208083019390935260409182016000908120918816815292529020549050828110156105675760405162461bcd60e51b81526004018080602001828103825260258152602001806108ec6025913960400191505060405180910390fd5b61057b6105726105c4565b858584036105c8565b5060019392505050565b60006103636105926105c4565b84846106b4565b6001600160a01b03918216600090815260016020908152604080832093909416825291909152205490565b3390565b6001600160a01b03831661060d5760405162461bcd60e51b81526004018080602001828103825260248152602001806108c86024913960400191505060405180910390fd5b6001600160a01b0382166106525760405162461bcd60e51b81526004018080602001828103825260228152602001806108336022913960400191505060405180910390fd5b6001600160a01b03808416600081815260016020908152604080832094871680845294825291829020859055815185815291517f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259281900390910190a3505050565b6001600160a01b0383166106f95760405162461bcd60e51b81526004018080602001828103825260258152602001806108a36025913960400191505060405180910390fd5b6001600160a01b03821661073e5760405162461bcd60e51b81526004018080602001828103825260238152602001806108106023913960400191505060405180910390fd5b61074983838361080a565b6001600160a01b038316600090815260208190526040902054818110156107a15760405162461bcd60e51b81526004018080602001828103825260268152602001806108556026913960400191505060405180910390fd5b6001600160a01b038085166000818152602081815260408083208787039055938716808352918490208054870190558351868152935191937fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929081900390910190a350505050565b50505056fe45524332303a207472616e7366657220746f20746865207a65726f206164647265737345524332303a20617070726f766520746f20746865207a65726f206164647265737345524332303a207472616e7366657220616d6f756e7420657863656564732062616c616e636545524332303a207472616e7366657220616d6f756e74206578636565647320616c6c6f77616e636545524332303a207472616e736665722066726f6d20746865207a65726f206164647265737345524332303a20617070726f76652066726f6d20746865207a65726f206164647265737345524332303a2064656372656173656420616c6c6f77616e63652062656c6f77207a65726fa2646970667358221220bb703c9c726f60b54cd16fcdcf459351563c86ad2912372653cf90c5760d2a3c64736f6c63430007030033"
evm_creatorAddr="1JcF5wH8PuQHHcDQECaGoqxegPc2kZcKcn"
evm_creatorAddr_key="0x238905271d330592886f8b30bef8f95fa66c87023c96b9a59638a080f7503c67"
evm_contractAddr=""
evm_addr=""
evm_transferAddr="169Ld7r2QCt9nXE9EUJpL12DhtBJwCYoyD"
#evm_transferKey="0x847d767fa4973c5386adbe9e3e9a3c01e096c54b7b12ab01fd8c469579bc2632"
txHash=""
MAIN_HTTP=""
paraName=""
gas=0

evm_SignTxAndEstimate() {
    gas=0
    local txHex="$1"
    local from=$2
    local MAIN_HTTP=$3

    req='{"method":"Chain33.Query","params":[{"execer":"evm","funcName":"EstimateGas","payload":{"tx":"'${txHex}'", "from":"'${from}'"}}]}'
    chain33_Http "$req" "${MAIN_HTTP}" '(.result != null)' "EstimateGas" ".result.gas"
    gas=$((RETURN_RESP + 10000))
    echo "the estimate gas is = ${gas}"
}

#上述未签名交易使用以下指令进行创建
#./chain33-cli evm create -b '[{"inputs":[{"internalType":"string","name":"name_","type":"string"},{"internalType":"string","name":"symbol_","type":"string"},{"internalType":"uint256","name":"supply","type":"uint256"},{"internalType":"address","name":"owner","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"spender","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"spender","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"subtractedValue","type":"uint256"}],"name":"decreaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"addedValue","type":"uint256"}],"name":"increaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]' -c "0x608060405234801561001057600080fd5b50604051610bd1380380610bd18339818101604052608081101561003357600080fd5b810190808051604051939291908464010000000082111561005357600080fd5b90830190602082018581111561006857600080fd5b825164010000000081118282018810171561008257600080fd5b82525081516020918201929091019080838360005b838110156100af578181015183820152602001610097565b50505050905090810190601f1680156100dc5780820380516001836020036101000a031916815260200191505b50604052602001805160405193929190846401000000008211156100ff57600080fd5b90830190602082018581111561011457600080fd5b825164010000000081118282018810171561012e57600080fd5b82525081516020918201929091019080838360005b8381101561015b578181015183820152602001610143565b50505050905090810190601f1680156101885780820380516001836020036101000a031916815260200191505b5060409081526020828101519290910151865192945092506101af916003918701906101e9565b5082516101c39060049060208601906101e9565b5060028290556001600160a01b03166000908152602081905260409020555061027c9050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061022a57805160ff1916838001178555610257565b82800160010185558215610257579182015b8281111561025757825182559160200191906001019061023c565b50610263929150610267565b5090565b5b808211156102635760008155600101610268565b6109468061028b6000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c8063395093511161007157806339509351146101d957806370a082311461020557806395d89b411461022b578063a457c2d714610233578063a9059cbb1461025f578063dd62ed3e1461028b576100a9565b806306fdde03146100ae578063095ea7b31461012b57806318160ddd1461016b57806323b872dd14610185578063313ce567146101bb575b600080fd5b6100b66102b9565b6040805160208082528351818301528351919283929083019185019080838360005b838110156100f05781810151838201526020016100d8565b50505050905090810190601f16801561011d5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6101576004803603604081101561014157600080fd5b506001600160a01b03813516906020013561034f565b604080519115158252519081900360200190f35b61017361036c565b60408051918252519081900360200190f35b6101576004803603606081101561019b57600080fd5b506001600160a01b03813581169160208101359091169060400135610372565b6101c3610421565b6040805160ff9092168252519081900360200190f35b610157600480360360408110156101ef57600080fd5b506001600160a01b038135169060200135610426565b6101736004803603602081101561021b57600080fd5b50356001600160a01b0316610471565b6100b661048c565b6101576004803603604081101561024957600080fd5b506001600160a01b0381351690602001356104ed565b6101576004803603604081101561027557600080fd5b506001600160a01b038135169060200135610585565b610173600480360360408110156102a157600080fd5b506001600160a01b0381358116916020013516610599565b60038054604080516020601f60026000196101006001881615020190951694909404938401819004810282018101909252828152606093909290918301828280156103455780601f1061031a57610100808354040283529160200191610345565b820191906000526020600020905b81548152906001019060200180831161032857829003601f168201915b5050505050905090565b600061036361035c6105c4565b84846105c8565b50600192915050565b60025490565b600061037f8484846106b4565b6001600160a01b0384166000908152600160205260408120816103a06105c4565b6001600160a01b03166001600160a01b03168152602001908152602001600020549050828110156104025760405162461bcd60e51b815260040180806020018281038252602881526020018061087b6028913960400191505060405180910390fd5b6104168561040e6105c4565b8584036105c8565b506001949350505050565b600890565b60006103636104336105c4565b8484600160006104416105c4565b6001600160a01b03908116825260208083019390935260409182016000908120918b1681529252902054016105c8565b6001600160a01b031660009081526020819052604090205490565b60048054604080516020601f60026000196101006001881615020190951694909404938401819004810282018101909252828152606093909290918301828280156103455780601f1061031a57610100808354040283529160200191610345565b600080600160006104fc6105c4565b6001600160a01b03908116825260208083019390935260409182016000908120918816815292529020549050828110156105675760405162461bcd60e51b81526004018080602001828103825260258152602001806108ec6025913960400191505060405180910390fd5b61057b6105726105c4565b858584036105c8565b5060019392505050565b60006103636105926105c4565b84846106b4565b6001600160a01b03918216600090815260016020908152604080832093909416825291909152205490565b3390565b6001600160a01b03831661060d5760405162461bcd60e51b81526004018080602001828103825260248152602001806108c86024913960400191505060405180910390fd5b6001600160a01b0382166106525760405162461bcd60e51b81526004018080602001828103825260228152602001806108336022913960400191505060405180910390fd5b6001600160a01b03808416600081815260016020908152604080832094871680845294825291829020859055815185815291517f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259281900390910190a3505050565b6001600160a01b0383166106f95760405162461bcd60e51b81526004018080602001828103825260258152602001806108a36025913960400191505060405180910390fd5b6001600160a01b03821661073e5760405162461bcd60e51b81526004018080602001828103825260238152602001806108106023913960400191505060405180910390fd5b61074983838361080a565b6001600160a01b038316600090815260208190526040902054818110156107a15760405162461bcd60e51b81526004018080602001828103825260268152602001806108556026913960400191505060405180910390fd5b6001600160a01b038085166000818152602081815260408083208787039055938716808352918490208054870190558351868152935191937fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929081900390910190a350505050565b50505056fe45524332303a207472616e7366657220746f20746865207a65726f206164647265737345524332303a20617070726f766520746f20746865207a65726f206164647265737345524332303a207472616e7366657220616d6f756e7420657863656564732062616c616e636545524332303a207472616e7366657220616d6f756e74206578636565647320616c6c6f77616e636545524332303a207472616e736665722066726f6d20746865207a65726f206164647265737345524332303a20617070726f76652066726f6d20746865207a65726f206164647265737345524332303a2064656372656173656420616c6c6f77616e63652062656c6f77207a65726fa2646970667358221220bb703c9c726f60b54cd16fcdcf459351563c86ad2912372653cf90c5760d2a3c64736f6c63430007030033" -n "deploy erc20" -p "constructor(zbc, zbc,3300, 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt)" -s erc20 -f 1 --chainID 33 --paraName user.p.para.
function evm_createContract() {
    expire="120s"
    tx=$(curl -ksd '{"method":"evm.CreateDeployTx","params":[{"code":"'${erc20_code}'", "abi":"'"${erc20_abi}"'", "note": "deploy erc20", "alias": "zbc", "parameter": "constructor(zbc, zbc, 3300, '${evm_creatorAddr}')", "expire":"'${expire}'", "paraName":"'"${paraName}"'", "amount":0}]}' "${MAIN_HTTP}" | jq -r ".result")
    evm_SignTxAndEstimate "${tx}" "${evm_creatorAddr}" "$MAIN_HTTP"

    echo "evm_createContract :: the estimate gas is = ${gas}"
    tx=$(curl -ksd '{"method":"evm.CreateDeployTx","params":[{"code":"'${erc20_code}'", "abi":"'"${erc20_abi}"'", "fee":'${gas}', "note": "deploy erc20", "alias": "zbc", "parameter": "constructor(zbc, zbc, 3300, '${evm_creatorAddr}')", "expire":"'${expire}'", "paraName":"'"${paraName}"'", "amount":0}]}' "${MAIN_HTTP}" | jq -r ".result")
    chain33_SignAndSendTx "${tx}" "${evm_creatorAddr_key}" "$MAIN_HTTP" "${expire}" "${gas}"
    txHash=$RAW_TX_HASH

    queryTransaction "jq -r .result.receipt.tyName" "ExecOk"
    echo "CreateContract queryExecRes end"
    chain33_BlockWait 1 "$MAIN_HTTP"
}

# $1 parameter; $2 caller; $3 resok unpackData 匹配的数据
function evm_callQuery() {
    local parameter=$1
    local callerAddr=$2
    local resok=$3
    local name=$4

    req='{"method":"Chain33.Query","params":[{"execer":"evm","funcName":"GetPackData","payload":{"abi":"'${erc20_abi}'","parameter":"'${parameter}'"}}]}'
    chain33_Http "$req" "${MAIN_HTTP}" '(.result != null)' "GetPackData" ".result.packData"
    echo "$RETURN_RESP"

    req='{"method":"Chain33.Query","params":[{"execer":"evm","funcName":"Query","payload":{"address":"'${evm_contractAddr}'","input":"'${RETURN_RESP}'","caller":"'${callerAddr}'"}}]}'
    chain33_Http "$req" "${MAIN_HTTP}" '(.result != null)' "Query" ".result.rawData"
    echo "$RETURN_RESP"

    req='{"method":"Chain33.Query","params":[{"execer":"evm","funcName":"GetUnpackData","payload":{"abi":"'${erc20_abi}'","name":"'${name}'","data":"'${RETURN_RESP}'"}}]}'
    chain33_Http "$req" "${MAIN_HTTP}" '(.result != null)' "GetUnpackData" ".result.unpackData[0]"
    echo "$RETURN_RESP"

    [ "${RETURN_RESP}" == "${resok}" ]
    rst=$?
    echo_rst "$parameter query" "$rst"
}

function evm_addressCheck() {
    req='{"method":"Chain33.Query","params":[{"execer":"evm","funcName":"CheckAddrExists","payload":{"addr":"'${evm_contractAddr}'"}}]}'
    resok='(.result.contract == true) and (.result.contractAddr == "'"$evm_contractAddr"'")'
    chain33_Http "$req" "${MAIN_HTTP}" "${resok}" "CheckAddrExists"

    evm_callQuery "symbol()" "${evm_creatorAddr}" "zbc" "symbol"
}

function evm_transfer() {
    expire="120s"

    tx=$(curl -ksd '{"method":"evm.CreateCallTx","params":[{"abi":"'"${erc20_abi}"'", "note": "evm transfer rpc test", "parameter": "transfer('${evm_transferAddr}', 20)", "expire":"'${expire}'", "contractAddr":"'"${evm_contractAddr}"'", "paraName":"'"${paraName}"'"}]}' "${MAIN_HTTP}" | jq -r ".result")
    evm_SignTxAndEstimate "${tx}" "${evm_creatorAddr}" "$MAIN_HTTP"

    # shellcheck disable=SC2090
    tx=$(curl -ksd '{"method":"evm.CreateCallTx","params":[{"abi":"'"${erc20_abi}"'", "fee":'${gas}', "note": "evm transfer rpc test", "parameter": "transfer('${evm_transferAddr}', 20)", "expire":"'${expire}'", "contractAddr":"'"${evm_contractAddr}"'", "paraName":"'"${paraName}"'"}]}' "${MAIN_HTTP}" | jq -r ".result")
    chain33_SignAndSendTx "${tx}" "${evm_creatorAddr_key}" "$MAIN_HTTP" "${expire}" "${gas}"

    evm_callQuery "balanceOf(${evm_creatorAddr})" "${evm_creatorAddr}" "3280" "balanceOf"
    evm_callQuery "balanceOf(${evm_transferAddr})" "${evm_transferAddr}" "20" "balanceOf"
}

# 查询交易的执行结果
# 根据传入的规则，校验查询的结果 （参数1: 校验规则 参数2: 预期匹配结果）
function queryTransaction() {
    validators=$1
    expectRes=$2
    echo "txHash=${txHash}"

    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${txHash}"'"}]}' -H 'content-type:text/plain;' "${MAIN_HTTP}")

    times=$(echo "${validators}" | awk -F '|' '{print NF}')
    for ((i = 1; i <= times; i++)); do
        validator=$(echo "${validators}" | awk -F '|' '{print $'"$i"'}')
        res=$(echo "${res}" | ${validator})
    done

    if [ "${res}" != "${expectRes}" ]; then
        return 1
    else
        res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${txHash}"'"}]}' -H 'content-type:text/plain;' "${MAIN_HTTP}")
        if [ "${evm_addr}" == "" ]; then
            if [ "$ispara" == false ]; then
                evm_addr="user.evm.${txHash}"
            else
                evm_addr="user.p.para.user.evm.${txHash}"
            fi
        fi

        if [ "${evm_contractAddr}" == "" ]; then
            # 去掉 hash 前面的 0x
            txhash=${txHash:2}
            evm_contractAddr=$(curl -ksd '{"method":"evm.CalcNewContractAddr","params":[{"caller":"'"${evm_creatorAddr}"'","txhash":"'"${txhash}"'"}]}' "${MAIN_HTTP}" | jq -r ".result")
        fi
        return 0
    fi
}

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    local main_ip=${MAIN_HTTP//8901/8801}
    chain33_ImportPrivkey "${evm_creatorAddr_key}" "${evm_creatorAddr}" "evm_test" "${main_ip}"

    local from="${evm_creatorAddr}"
    local evm_addr=""

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$from" 12000000000 "${main_ip}"
        chain33_QueryBalance "${from}" "$main_ip"

        evm_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"evm"}]}' "${MAIN_HTTP}" | jq -r ".result")
    else
        paraName="user.p.para."

        chain33_applyCoins "$from" 1000000000 "${main_ip}"
        chain33_QueryBalance "${from}" "$main_ip"

        local para_ip="${MAIN_HTTP}"
        chain33_ImportPrivkey "${evm_creatorAddr_key}" "${evm_creatorAddr}" "evm_para_test" "$para_ip"

        chain33_applyCoins "$from" 12000000000 "${para_ip}"
        chain33_QueryBalance "${from}" "$para_ip"

        evm_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.evm"}]}' "${MAIN_HTTP}" | jq -r ".result")
        chain33_SendToAddress "$from" "$evm_addr" 10000000000 "$MAIN_HTTP"
    fi

    echo "evm=$evm_addr"
    chain33_SendToAddress "$from" "$evm_addr" 10000000000 "$MAIN_HTTP"
}

function run_test() {
    # 部署合约
    evm_createContract

    # 检查合约地址是否存在
    evm_addressCheck

    # 转帐
    evm_transfer
}

function main() {
    chain33_RpcTestBegin evm

    MAIN_HTTP=$1
    echo "main_ip=$MAIN_HTTP"

    init
    run_test

    chain33_RpcTestRst evm "$CASE_ERR"
}

chain33_debug_function main "$1"
