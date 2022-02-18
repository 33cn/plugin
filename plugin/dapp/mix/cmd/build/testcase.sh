#!/usr/bin/env bash

#1ks returner chain31
MIX_CLI31="docker exec ${NODE1} /root/chain33-cli "
#1jr  authorize chain32
MIX_CLI32="docker exec ${NODE2} /root/chain33-cli "
#1nl receiver  chain30
MIX_CLI30="docker exec ${NODE4} /root/chain33-cli "

# shellcheck source=/dev/null
#source test-rpc.sh

function mix_set_wallet() {
    echo "=========== # mix set wallet ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    mix_import_wallet "${MIX_CLI31}" "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" "returner"
    #1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
    mix_import_wallet "${MIX_CLI32}" "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" "authorizer"
    #1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
    mix_import_wallet "${MIX_CLI30}" "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" "receiver1"
    #1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
    mix_import_key "${MIX_CLI30}" "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" "receiver2"

    mix_enable_privacy

}

function mix_import_wallet() {
    local lable=$3
    echo "=========== # save seed to wallet ============="
    result=$(${1} seed save -p 1314fuzamei -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib" | jq ".isok")
    if [ "${result}" = "false" ]; then
        echo "save seed to wallet error seed, result: ${result}"
        exit 1
    fi

    echo "=========== # unlock wallet ============="
    result=$(${1} wallet unlock -p 1314fuzamei -t 0 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi

    echo "=========== # import private key ============="
    echo "key: ${2}"
    result=$(${1} account import_key -k "${2}" -l "$lable" | jq ".label")
    if [ -z "${result}" ]; then
        exit 1
    fi

    echo "=========== # wallet status ============="
    ${1} wallet status
}

function mix_import_key() {
    local lable=$3
    echo "=========== # import private key ============="
    echo "key: ${2}"
    result=$(${1} account import_key -k "${2}" -l "$lable" | jq ".label")
    if [ -z "${result}" ]; then
        exit 1
    fi
}

function mix_enable_privacy() {
    ${MIX_CLI31} mix wallet enable -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    ${MIX_CLI32} mix wallet enable -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
    ${MIX_CLI30} mix wallet enable -a "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k,1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"

}
function mix_transfer() {
    echo "=========== # mix chain transfer ============="

    ${CLI} send coins send_exec -a 10 -e mix -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

    ##config mix key
    ${CLI} send coins transfer -a 10 -n transfer -t 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #authorize
    ${CLI} send coins transfer -a 10 -n transfer -t 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #receiver
    ${CLI} send coins transfer -a 10 -n transfer -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    ${CLI} send coins transfer -a 10 -n transfer -t 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

    #receiver key config
    #12q
    ${CLI} send mix config register -r 9609397526062191255833207775509487457674460306914263472031059870638285140780 -e fd1383f79872c41d9af716e64f4a72653faff01858a58122d6a8480ae1eafb04 -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #14k
    ${CLI} send mix config register -r 15016592569780695699649849235075665744274166257234430418287460744404832890230 -e 001a01b0e39a4e06a6a0470a8436be3d6107ce7312d7c56d41fccb91ffa2031c -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944
    ##1ks
    ${CLI} send mix config register -r 16678747381284372741157128409332526143974006721672403765375251027071805395166 -e 78e2dd2c33f9cd7a94b69962a164da935e91c3c7fef8cfbf810491a128ef396b -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    #1jr
    ${CLI} send mix config register -r 19382709574058928399389231120651265926433088739829655203747560667222689803490 -e e5362c31a903cd5522c4b84c324f90e96851292194fdfb33a8a1244bf1bb9f13 -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    #1nl
    ${CLI} send mix config register -r 7734493727243297635064531306209257758671783528296286360590183768080317922454 -e 0abaa15456580365b90f84f22186f99250f4198f8df7319bcced1606085a1e01 -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115
    #1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
    ${CLI} send mix config register -r 18437326986701045682163784849869247633492934399146571227371858493337922483431 -e a97592e700eb0f87c5738b35c8d460ce33a4a59bde6128081ddd042c3c262f76 -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -k 0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71

    ##config deposit circuit vk
    ${CLI} send mix config vk -c 0 -z 96c05da56b3c1b7f4a4583a69dee138671ca451f70613c2942cfcfbfccfa2c93d0bd4156dbfc7897c7a92cb23c17385acbc8ab936307b837d473b7f11bc81f8d8896b18f7bdfdce4ecea5ca3cadfb60667f3c10b23f86a386d5139a95960a1172c4b7030d5fde47a678b13649db23c94046169f6d130614d99f250b7b98fa04aaaaf899ac00250216537690bc15515573156178b7ee1069c1d0c061d9ea12a7929be1929f520ecfc430fbe57f05de78ba05d251a8e94595763401d9b41c110eedbc1e830f31b2bc6dcb28f6c7b3efa126c4f1dd7b64457c08b8b3f1f5f481fe78763991d196124b216432df35883a9c6a342488a88cf4943959841fdd251f7d127bc9015d53971820e4a19693ea0d32313cf8f22ff190d5b354b8172eef5a30200000003a9a75fa99ff41836250ffa38f8cc1d306a24265871ae8c741985836d06a15a578758067cb6e975171a1b1251d7beeb14d0d0433ba440e8accebeea294ffec74990cf64198dac4373e477767991b854c0af29574444f4e2a6061089f58494ebba -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

    ##config withdraw vk
    ${CLI} send mix config vk -c 1 -z a6213e326b273793c91399708596994e12026598895567c8165568798f24e12490162ebde464cee8e793d392aa2602bb1d98566761978db11ff0412d954520ebe494bff0aa1156816829b71a323608645df0bc1a389d47a277b446c88c47c3422fd630d5b3f29b0d1cff8f0dc8bb553143d995b07a56092622a7fc20ce9f72eca1ad6410c7017a79306cf6be0e696f9d993fcc068201e1246f3cdfe38d9260141c50877c31bb4b597ea15c689350a51e488df1e5e4bb1783d6c72c7af77ef5b1e51aed7c2208029a3e78002952f8e5685cd9c10d21a0acf8c12b815ba5acf52d9dbcd50811c0fe0e8a3ce939a3187b7f78b4766e1076b7249c154d622bb7ffde20394f62d3d47a0dd356adec9fdd9857b06bcccea951e59ff5e1a0092b11138100000005da40b9b3327bfc2963ffe4033192377750e2ddb1bdf0ed2ee2e94e324fd90875c0bfc8c5e9687fc255d60dd93c6e71d585463f2312b2e1f2e1555a5cdedaf5a9dc346971ebf0c5ecd26b508909ed8f89d637bcad85144abfd765b28406e02b40ed7c1e4fa64c9ba0bba557008f27051063c1b0e556b732cb7ce68a21b4c85a6cc8abf84d7d8055c3d58ed4120fe685bc16fc5ab0db257fabfc19425e21de1d3e -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #transferInput
    ${CLI} send mix config vk -c 2 -z e62fc1b372f90851e6b5ef8bf2beea90f68c808d454f0ba9279d8770fbcb7aaaaf64ff166d17757c67387741bdc0cd8c0dad41627e112ce4c70378b7967b4ca5daf38993dc121bf0a4d99f63ff3a15530a1001cfa1f95314a56c1f5a20ea4baa06c3fa101d995befcea48349b2b48a9a5ea90c4d3fca65e8b580219bdb8ad84fcec067e59acbae6556aa582b97526af31d6a14a1a32fb55fe5b405ff938a811900fb5b132133d2bb664a94eb158b34eeb5564d1c95edcda368040698340b2eda98b0852e04bf4b6444f16988845bd9e9b177588f0196b1318589c75ee12a2aa4ddd3c56886dbb2e3413ba8c610f90d0862219101441b819754e1545e5fc68dc6135d28c850f28e93ddd72249e4bf4d8017f7f535f292271c5f2d0221006b484800000008ec08e311aff48ee0ec202be84894e450c010a66fb961d634768ff1b64e275f28ea4b4933cf1ed3551d265fead95cf2a7d82fe8cd54cf1385e82f2ab07aa077b3eb39b9533d340f1cce729aa8a549c0bbb43e4cd85320b51214ea662a9f587a7dc7c510baa2b8376cf8e6ba342eacd457597e0ad97064a52e0d666d7226f340d795a014476dea597f6489e8c78d83e9a80431e76930eff1a93114f263eed366e58feeeee1ba63dd4b7a2c591368fd6f0b2f108df757db819fecb213040e8700fee5fd47a0cb15eee6f0a29a90c924fd02111ad560ab180f8a5152ac8afa24f0e58d997310eb4b2d3ef5ec940f567fe6873a727d57dfc141ebf5c95e2a5bda7e43 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #transferOutput
    ${CLI} send mix config vk -c 3 -z 904696fe115469a4026e900f74ebf5ad29e82a53a985a90fb78d5e96f207d0ea9aa339f4ab794ee63105345c8652deaccb1fef36af69e6558869eea58397a16cadd326a5c409626a41bb5ec7bdb31be60e00c21f8ff00e32e00d26a010b7e5590fc5afe303fe45bfe00c9b6fef6b11ded99f1464107e400abcff1ad6cddecec3cd08cc2901330f9d348f91f161d96ec3b8aadcb9e4341d42edac2180255635ff2ed6903eebc1847a1aeb684cc5efc88ce270a672d7b4aba25444a5a6f2270e9e8db7418df39e7ebbfe9ad4edc295f4b83bc68d8a95ff23217fb31de2e9fe975594eb7d9ced4b19ce3fea6ba0a29995731ec40bcdc0a24a3047edc1dcd3df71cc077d15d8ea35f0fc28b58e3442cc596dc567ecc31914a5697e947de6725e33a7000000068b80a3f1b925b3c0675b99e15061b395fb23e17d5e8fbc1d0f9fdc36e430048cd55b0dd4287afa27a5fa4a8ad9b0ea26932c61de19d89561e6d721b22a4a920c9ab786ea56d1ba59714cdab6d57b171c6d7f08859b61264b84b490799b078a8693b42e03e0d34118fccd71b3d2949b3d33c12f1cacc7ab9b8b7c167a6d15f7b89af4637cfa3922afb46fc0288dbc254cb6a626996001392bce507f9b4279c05189811b9e3efee247a4a5d86e38dd2fbe8aa39f45df2b82b736437f00e26d7ff6 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #auth
    ${CLI} send mix config vk -c 4 -z c8a76b883915c30b1d7a0ef7b09d8dcd47f29e4b73f9f4e51def17634650f303c7592549b023673bae38b0ed0879a98b80e837c41e333b1f71e6e50c78704b86a335139865b62bad3edffa90ab341b9e69cdf8a93a24d9f97656e40c0672165d0b5f575333dac1ba990137a5f363660487d1b0d60be2439a2bfc9d0a8bb22a54eb7f0d01ad674a68b50df035050f9982646f1d716b95328abe289d19b36fed651c78b8ae59d146be8ed989f83e8cbf2be7e8207877282158820f0b1bef98754e976b7c57c9c18713fcb4288751ec12ec0157b1bdeda3c74e5839d0e0cb090bbea2c7c0e335d175a665975f6a0d694ba132c4a762c0094090cb361858a1f18de40ffd7b5244bd387126edc6a905d5026eb460eaa9273bcf2df714278d7f038bd100000004e46d01b0741050a8754dd8ed2cd37a491b1eeba636e6fdd40ce1f64c4e7d7c349833a8cdf9b25eead8d3118b2cf55b7615ae4cbe0564770647bfee0224ee8c5cdcca724cb511f57b7b5be5e36cd382b975a6e5471165595f55b83fb91cf921d09ff7bb8e51aee8bd6492d58b8fb2a9ce674eb2b87903f2ca7dfe78308ff1acfb -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

}

function mix_deposit() {
    hash=$(${CLI} send mix deposit -m 1000000000 -p ./gnark/ -v true -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -r 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -e coins -s bty -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI31}" 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 3
    query_note "${MIX_CLI32}" 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR 3
    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 3

    echo "auth"
    authHash=$(${MIX_CLI32} mix wallet notes -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -s 3 | jq -r ".notes[0].noteHash")
    authKey=$(${MIX_CLI32} mix wallet notes -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -s 3 | jq -r ".notes[0].secret.returnKey")
    echo "authHash=$authHash,authKey=$authKey"
    rawData=$(${MIX_CLI32} mix auth -n "$authHash" -a "$authKey" -p ./gnark/ -v true -e coins -s bty)
    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI31}" 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 1

    echo "transfer to 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
    transHash=$(${MIX_CLI31} mix wallet notes -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -s 1 | jq -r ".notes[0].noteHash")
    rawData=$(${MIX_CLI31} mix transfer -m 600000000 -n "$transHash" -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -p ./gnark/ -v true -e coins -s bty)
    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 1

    echo "withdraw"
    withdrawHash=$(${MIX_CLI30} mix wallet notes -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -s 1 | jq -r ".notes[0].noteHash")
    rawData=$(${MIX_CLI30} mix withdraw -m 600000000 -n "$withdrawHash" -p ./gnark/ -v true -e coins -s bty)
    signData=$(${CLI} wallet sign -d "$rawData" -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    hash=$(${CLI} wallet send -d "$signData")

    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 2

    ${CLI} account balance -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -e mix
    balance=$(${CLI} account balance -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -e mix | jq -r ".balance")
    if [ "${balance}" != "6.0000" ]; then
        echo "account 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k should be 6.0000, real is $balance"
        exit 1
    fi

}

function mix_token_test() {
    echo "config token fee"
    tokenAddr=$(${CLI} mix query txfee -e token -s GD | jq -r ".data")
    echo "tokenAddr=$tokenAddr"
    hash=$(${CLI} send coins transfer -a 10 -n transfer -t "$tokenAddr" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    echo "token-blacklist"
    hash=$(${CLI} send config config_tx -o add -c "token-blacklist" -v "BTY" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    echo "precreate"
    hash=$(${CLI} send token precreate -f 0.001 -i test -n guodunjifen -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -p 0 -s GD -t 10000 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    echo "finishcreate"
    hash=$(${CLI} send token finish -f 0.001 -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -s GD -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    ${CLI} token created

    echo "send_exec"
    hash=$(${CLI} send token send_exec -a 100 -e mix -s GD -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    echo "mix deposit"
    hash=$(${CLI} send mix deposit -m 1000000000 -p ./gnark/ -v true -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -e token -s GD -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 1
    echo "transfer to 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
    transHash=$(${MIX_CLI30} mix wallet notes -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -s 1 | jq -r ".notes[0].noteHash")
    rawData=$(${MIX_CLI30} mix transfer -m 600000000 -n "$transHash" -t 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -p ./gnark/ -v true -e token -s GD)
    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs 1

    echo "withdraw token GD"
    withdrawHash=$(${MIX_CLI30} mix wallet notes -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -s 1 | jq -r ".notes[0].noteHash")
    rawData=$(${MIX_CLI30} mix withdraw -m 600000000 -n "$withdrawHash" -p ./gnark/ -v true -e token -s GD)
    signData=$(${CLI} wallet sign -d "$rawData" -k 0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs 2

    ${CLI} account balance -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -e mix
    balance=$(${CLI} asset balance -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -e mix --asset_exec token --asset_symbol GD | jq -r ".balance")
    if [ "${balance}" != "6.0000" ]; then
        echo "account 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs should be 6.0000, real is $balance"
        exit 1
    fi
}
function query_note() {
    block_wait "${1}" 1

    local times=200
    while true; do
        ret=$(${1} mix wallet notes -a "${2}" -s "${3}" | jq -r ".notes[0].status")
        echo "query wallet notes addr=${2},status=$3 return ${ret} "
        if [ "${ret}" != "${3}" ]; then
            block_wait "${1}" 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "query notes addr=${2} failed"
                exit 1
            fi
        else
            echo "query notes addr=${2} ,status=$3 success"
            ${1} mix wallet notes -a "${2}" -s "${3}"
            break
        fi
    done
}

function query_tx() {
    block_wait "${1}" 1

    local times=200
    while true; do
        ret=$(${1} tx query -s "${2}" | jq -r ".tx.hash")
        echo "query hash is ${2}, return ${ret} "
        if [ "${ret}" != "${2}" ]; then
            block_wait "${1}" 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "query tx=$2 failed"
                exit 1
            fi
        else
            echo "query tx=$2  success"
            break
        fi
    done
}

function mix_test() {
    echo "=========== # mix chain test ============="
    mix_deposit
    mix_token_test
}

function mix() {
    if [ "${2}" == "init" ]; then
        echo "mix init"
    elif [ "${2}" == "config" ]; then
        mix_set_wallet
        mix_transfer

    elif [ "${2}" == "test" ]; then
        mix_test "${1}"
    fi

}
