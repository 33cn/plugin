#!/usr/bin/env bash

#1ks returner chain31
MIX_CLI31="docker exec ${NODE1} /root/chain33-cli "
#1jr  authorize chain32
MIX_CLI32="docker exec ${NODE2} /root/chain33-cli "
#1nl receiver  chain30
MIX_CLI30="docker exec ${NODE4} /root/chain33-cli "

xsedfix=""
if [ "$(uname)" == "Darwin" ]; then
    xsedfix=".bak"
fi

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
    ${CLI} send mix config register -r 14248533008859289736671040497097112376729713980306862599828676705210740679404 -e fd1383f79872c41d9af716e64f4a72653faff01858a58122d6a8480ae1eafb04 -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #14k
    ${CLI} send mix config register -r 15542837613302913931804198235633080071021890362349316969360913659116440971972 -e 001a01b0e39a4e06a6a0470a8436be3d6107ce7312d7c56d41fccb91ffa2031c -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944
    ##1ks
    ${CLI} send mix config register -r 6803726063008093410366112941811416911449216219005125852762197815950128434240 -e 78e2dd2c33f9cd7a94b69962a164da935e91c3c7fef8cfbf810491a128ef396b -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    #1jr
    ${CLI} send mix config register -r 11499201522891581270391851171638982479327681756562559547094861564688053581237 -e e5362c31a903cd5522c4b84c324f90e96851292194fdfb33a8a1244bf1bb9f13 -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    #1nl
    ${CLI} send mix config register -r 20573898766049640519279377248992057483378639946007599793634981932019201439513 -e 0abaa15456580365b90f84f22186f99250f4198f8df7319bcced1606085a1e01 -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115
    #1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
    ${CLI} send mix config register -r 16664447758440542061402448540177654829488285379043162144052840929545481936061 -e a97592e700eb0f87c5738b35c8d460ce33a4a59bde6128081ddd042c3c262f76 -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -k 0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71

    ##config deposit circuit vk
    ${CLI} send mix config vk -c 0 -z 11d72c948d3d6b88a49d99b55ac035875fbecc3654a0e275bb731b2da3acb94610bdc3e5aea2738b230cac2988e1290cd711f9170808e1c8eee0b4d1682c663825c8f7d853ca655f502524091ce0a48e7229f05df8e40bd31839ba19ef9ea3b81d5d15428b59e5c1b523917a989c189a2b9ed799cb18318160b47829868138cb2b49730f693a497e022a14344374410004ed29ab3333a2c4ff5a63d486d841210969ea406d44fc0af8f5a3ded0fe76f9a90c6c42c518e3725d07f91c0bcda1330982f9732d7a542b11ef8c90c2bb39d9f85f8bc240c2bea852d4dda3961be1e22a33bf4a7120e37df97c58aa10ac88fc365f156bffad5f0709615a0389b6542401689987d98818e7c0c55ca092ae21dd9895a5baba579a68bac5807e4f37677e128265980bf81da9622334495acd490d986068d12a148ec26f6b16d74e22f19d2ac20e521154c443fe93dba26c2b1cb6c8fa2533b8029ba28c168b80398b2a980268c89af9b92ed851e5cb18f1a688ce8d654b7db547c6753516776e6a16662c14ef921c2f016514a09a2265112b8f24da3267828793c9e37a5fa156dd3f5e8b279ac92cb2f35e64e2620def2c45b51503ee4552a1da0cfb055711efc08d89fe0e3a407989f7b14a8b0353a5ca8907be1d2551b2c085b2104e3afba15e30ad0d0527b06bd267ab6dffdc80b5ba6ff86adfb8c3f87e48a4ab3ba46de238e400ca1c7f4a396c30d31b20c756a1789213657c0fff66d7e7f8b1a3ec9728a1aada9622b0c732c615c9c78322da5da67ab6861ee30eaba4b0103c956265310e743f65000000031b3e8a8ac3b0ff4f009474b7dfb2e99e8c1368440219b8418351d2de4ec193c61b8fad37c1d5028da1e6a6b67e2c6858868710598175f4c6c07850d5e8ab53c120380af328d225d79a6ad6a716b7421de1b6f26793683fb83fae14ab6858486d062af23453da53472221872353ae3ac2900f711ab285de1cd9562eed856333c425767d1fb4797b1a64836d343541193cbfcc0e748fbb8a8b2f315505927c20fe1a87bc97b5f5d4e61216b061f2daa0703510a9813d253f710d73ad862c916603 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    ##config withdraw vk
    ${CLI} send mix config vk -c 1 -z 17497f382733020c2441ee2179d468e4a871353ecf52e7bc7f57600703e6cd0417c8a91b86c1e6e55316bd1ba88db8c817b646add9c5289497faf705e2e0a50510a31a9fbbcbec3399e15f1caea1e372a51827e08335606f7f1a9428539e4a3025ea43142772f0300fcb6e2e03057be5deebbc7e00cad1019da0452c2ab90d582caa5e4c22103eadf15f5ded717187934e4e9a44997766d6a770c42ff520e00d1d5a65abd94ac973cfb60d07499505d4db9584ccb428041336610d422a6d203007171dc5d586c536b3f49c2daf4d3b0a095a67ea759b49f182c87b41ac03848722b838c495beb5537c43fa7175bc5346743a2eb620856d0daa92614f283027ae2aff76e2a53c74648f52d9bd65ccf3064f2b9d9230e45bdaaa2214aa593d90180e85e2e045eac57a2b2b86d3eb23215bdc6b8b0ecb80e37374d41e1f7101acb502a75d203e2e058258f53f8c58192c8f0849a67a5ab5c7d26d782e4e1a10669c03285a2eae35d76a09c20841c49e72e7d7896da7a788d146836d5ba60bd2981b0eeabf8ee62fd2406056b556db292d18cbda9bcbe008791cb96fcb17ca6a33b0090652012ee82cce78f6ba28fe807cd3f668ba1fbdf795002d5ca1335a5de65307936b649d057637f8ff815d86aef66fc49960d02a8244cbd64658bc5758ee2a271e7a651fd9ef47bd7683c5457c724365c4a49e5e4ab88c980e33292ca1467c1e4b36114f7ed29b48bc2cf422be5c8ca374647971eb7daa1016348546190a762cc190a0d413e00c81f098a84e880bb53024be29917ebba4a55d59c40a606c0d0000000519dd32ea5a02431cbd4c38143012542385d9085b4c5d0b4f176285feae96d232061ed041ec53ee1897dc1eb82f1456bcb939b7512fe60548c97873bc602606c91f1d2dea1819016ef4bf1f1f3a6e074f04f04117d0628375f15779bc612078a302aa853d4d2e348cf05422efffd585fea5826ae63770b35be4ba8087092cf5150560d8d5e21d5b60413658c69f52d0ab2b399490f7e572ad8612a08472721ceb107ffe99a3a4de489fa4849a63eb5a574ddac2ccf7afe89c06c77c43a94e4d4b1c1250247c5352bd7b34dcd075f52cb31c5f4c96ba3120355a3dc010d330c4e02707c707d9cc3772ce61f7ff8bbd50d2633453a21d10edf821ceea93ba205bee2955c296632546fffbbdfdeea2f1aa67a9b1a90903ea12a3818d5e44f9964a6109afe1cbce7b5bccdf71718f22a3d57f6afd1a7976a75c093410520cf79bebf8 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #transferInput
    ${CLI} send mix config vk -c 2 -z 153cd9b6c5bd77f41c63012def12cfa719cedbbb7c376353ada8865bd9d386a22ebdbbd1da14114ff44d1ffd49d3a1f108a6affd4614c9c519fc84425bec6293145cb2f2e7c8945ed7e9a7a7ef02c1555bb0886e7c15a8be254cbab1adbbe33919b40f19c82819808f1f87ac6ae96657a60ff37557ec55e73043f51a3ae17ade00973ce23495e3dc7941b468ff0e9b9b0ae73d722fc5401983a8def9f48ab76020b317cf6ccbc216c4a20210c34e061b2d3f74e8a1645b8be4b7ff073abdb22923d6b82003695d70bf025442860d9979f4b547569d1c79bcb36bd85498d66969304a7edfa12d4ef559b07b93834024f424cc73d115e231d31c0489d22a19403321583c42261bb161b3566c574c42bfc80ef00bb5890e40b90ef3361eadb38f181144271f45c753afda2f81bdd2868e1adb7b2519d2f75ddb05e7fe7250a2bfc00beb6540deba4e021b1edc5d0c5e08015674d5db33c591714297acf1f02c35af1069621ead725fc0527099915ea5de69282cb9850de96495e86b61670f9146b411a9ea9a4df35af15a333495475ddd34df53692471d09e7fbc337471829f64d203a199091300223531a0298b17ad923c29f8982b32862b1c091d970f9b5cc9ef23a7761a669a5dd9d631424ea8879b89b22afcee997fc6ab95cdf9dc9baac6461c1bc5f5736b339568767e9ccab43e63dbfbd3f3a94f1100f566fa5f72029b701fda13678ff8db802156bce08bcc2ededa2914ac45d4be6d459abd612fbdd77718874dcb119724bcd71c8de5e10f7c79f4b66434f8b9a1e2a174d92cb10b1ca70000000628e6eaab99088247e8106f52f47b5940f3c5294d2283006bf3c29532c59a4c5a0e931758722334e0299af3b6cb36e9457fcac0bcd5440d80355babd3c05e88211e1e7f5bcdc4863afa2d49c743652305f971c92920c0bb61e76a1bb51f08156303ac2e2f58c0a6f2627b820ff08f0cedbbbfd99fa840b79de57e0f6ebf5969ce2efddfb68b6c57535a6be061335009ea879087b15def360546c5f1a4b0f737bf0d785f07ada66b23c6a037ce372994fd072a97c16d74c513f63e987217ec8dc210cbcb5c6dd8ffe745a63e88470fc09cade0ef877c4f107c08070fa19674618813e5be4d5db40af84bd83db80fccadee83af1549c92f4c935256bd5dd68b97f32e8dff5346fb5fa554c9ce617774c2a43cd7b6c5d5edbc9cd0dce0ba74b5ec87043b59f5b19c20057780c99db499d2565bc9f90c252f27cc37b2f5c9c69f504c13357153d572be7823f67ca794ca52088e20396eb7bef837109f542041fdd56310c5bd398cbfa8e588fd8342cbd103f58249481f5af3eac33f15add8fe4a64e4 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #transferOutput
    ${CLI} send mix config vk -c 3 -z 13934cfcc6ff87d92bd9bf14fc22abdda42ecfd544a0dd8b5b38d8936cb91f7d2ab93a5f2e54a2b09ae0f822be44bb304d2e8a3a9a2590e134cef4de7ca61e6222f85e2b09a87301942fa12d8755467ce88b30679f703b48a3ed973b526648c614058f734069dc99be9312b0f64b46dbfe8b5a1c7a7b2abd301417abe61cd7580f07fb23e6b6e998dbc16fa3f5e9d0489ec2867f496f79364628731c8a81159d26ba7de29f451d5b6128224f398ad7797a605b022f5a6de1b59276a339d88921218971e393ea6027f758ddffe692b11269b5ab17f72266661cc452b83f83a70818d7a37dabff75b1995952a63c45ef51695e18ce1b9b6bf0da5dbe3c9625ad090dd507dc37e610ea95e32f2109bd7ae0783484675a445cf396ca0653d4e3c71b27c0b8d2d3239ce11da02406913fbec4d973bb1ff1922a6d5d295581f635afc22cc1379afb9f5158d88baed9020b8d393f1c7157b30a36724d6efcf162fad84e20cd61fa92f99ef48a4ab4a0d93f91f06204addb1686344476605f2089defb561c4b0f8502b3dbe27395b86e8535fc4ca14c25c071880f8153e931194473b2d809dae5160125311328f69840cc36c0d68e3b135827f79076ec6cde7c94b607792485812b35437227068c0122eaa92ab9e181a7b6776b681d3c6d58d3a2d7e8462537bc06071d29fc3ec2e789d343359fd2adb5aa1e343033b6456cd9f9c9859120c39266d6032c2e4cd703c2ebaf1ed66edeaa95b8e9363190cb2e33b7c430ad2e7d9a9587ed281bf44e2305aa20c3e031ede1cfc40cfc600b09b45b0647cf1100000004269eb69c9017e0e3e8f2fefae3d2bf072ce5a9369b9dc4a949863a096b8c2fe70e53a7aaf51d02342549315f63b45527dbc1fc8b640b471337ea89ab78179c6305ba8d54c55ef56feb4623a09629f70164572993ac8ff3472753d643f325622220f4e4183a52903095d6989f5f3aaa18da0693d815e18eacab419e1306a492d42b1a34104b599e2c91e697969f4e47fb2305f5e0e587b1616c2cf53d4ac5ebba004710a9c23bfd6dfe3b0b4ce9b934fc6a6b2304640d56fc2ecdc20d37dbe352048691666bb7323b37f75cb2a3c078e05ee3d38ba666f33ee3c9cf9cbe2e1a901fcfbf411bdbf8b3e671579c29fe2da9b29cc1e99ab30801d8b9a1bd87c5cd8c -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #auth
    ${CLI} send mix config vk -c 4 -z 1473ab4328616752c00ea248309e10bce4b92e3fd4e3047e30747512821e24741a59bb2115d794d53e8a123b284f59acfccfe5fd76c87a4ee5cf72a24e415d52049eddd126746ca46f1a6b5151354844fba1cc136f56dc8d52ee087e77ee545a24793c266c6760371f66e3ddcbd2f50fdb40ebcd63327d0daeccace522636aa42525af4490cdbcba010d5d60b17b899a5d4ce32a4f028153d18ccb9b54108e35215fc55fa11835b75755f4b2e71af4787c7d8424c8fcde9b59d0283fa1a6596c287a8695a5c4fbe45979ba25930455fddeec304c1e0fe1d980730a262bd3c15a097ed012586cc746756ff350ba08c0f7e97b1e3368ad99d4b413419544a5ceea09b9dbb381264e8a786e9875c7fc9989ad119aa3a8e1b07f170dfc29b6603ac1245629f9260508d08747a7b8ddc2502be73cdc16e467d47df69aa71575d39119197ce70dbe1e1f1d8f1b45a3dddc63e6e236794fe08b42080e6746e8e22628f70d436597422407240ed94d46b32d0914451225dfeb64124f7215476148092e190010e75b4d11104f80c5259fda8778d6886fcbf789c48a13e21d7357eafd8a8308b0ae3dfcc341e715bdab7de49fb699330025da6060eb91b5a79cbe25e1cb3705f350c2d348ef3df4a6cccb11e68874fc8e75d617b3934a7c654c8c854f3752091291b80aa18de8fbd41b0ec39ebbd96ca685e6386da6cc204ae13a9184d18b044f26507b4b443e60e02162492364452a83246e9643adbd42e87316dc04382b29c76acfba7614956b9d9aed7e49be9258268e22adbba0aee818a3f8bc1c3bf40000000411f40834f563e79b2ffabc45c24797299ee34346888f34802b784b62427bd5d429f5117115d7162353b65557850a4a2f99e9834ca8ad95830ede631d897048521fabfd0dd4246c4d8e98062daf771ded13921619e76df0ae190ed0200757684e02ef7e138bf53134af024817312c725cc8a6617c8ecdd66d65e90576f59596c90358b4ce57d7b4832639096c4fc814fa067b3d4b11a4ba8626c543b26b68bac505c2ac761fb39a0b160010b4085dbe7ab2f4c940b96fa75fbcd0fefa6fbf12811e86674dd28531b9b264e6f098b91344df5755dd28a41696d1ffa878f5dd9ada0be976bdfcb9f5538df6dc9eee17babeeecaf3c93b4f317d189bb82cb0b7c264 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

}

function mix_deposit() {
    proof="0e420e0a8963195ab7fc4a610d8cd758a04be4950317c1c0cdfff701e207d03c0f72b5c669058434b3db24ef57f9a0d9e443d78a3b661e1cd2a1f6c1cdbdb855081c95f92a68286a0aa4d17615f9e61ed073ecb9aa4cbb86feaf4007dc86de8f139e1ff01ff5a6ff5e46db1ede026243c47755cfc33c3dbe4f8871fc62322a3a230ffd865b7b5bfc0e99677877517f562245cdda012e28941b4bad8a25d5edb427152ca0948940a12f489aa259377b0eb0cd1f084faf4dd9ae08c43aba780a6a0245a3a5d13253523272bc7b82ed16ecbcfa94fce8f3033fcf863240134265b70e2ff0b31a35ce0c5d0cba7efb7a4ede29e884da05ee3ede095433c9bfd6fea5"
    hash=$(${CLI} send mix deposit -m 1000000000 -p ./gnark/ -w "" -v true -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -r 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -e coins -s bty -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI31}" 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 3
    query_note "${MIX_CLI32}" 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR 3
    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 3

    echo "auth"
    authHash=$(${MIX_CLI32} mix wallet notes -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -s 3 | jq -r ".notes[0].noteHash")
    authKey=$(${MIX_CLI32} mix wallet notes -a 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -s 3 | jq -r ".notes[0].secret.returnKey")
    echo "authHash=$authHash,authKey=$authKey"
    proof="1c33f86a705d387a191d598cdf5a4d97d839e49013ec8b73fb265a62bafc03fc17e9b34a36c6f015f49d1e944c945ff16bc3c7b4886591578c780fa829f140560fea4be7ddf3a98a5a620010f7a210456158e3f20b88a5a120980c8b960d15fb1c50a4d93b78c3cb3bc729152e1cd0b1574b3cc63a4ce6fb5f8e51828320df861d70f529eab34712c213a605dda2f05eda7fc4b0aefc99adb2817eb10548489d12591f64709a5ad3dd32a1ebfcdb70c50288af1185b4fa3e9a25636b9cc9df0c22880ab3541018fa25a9fb232ed4dd47e0aa92de0c4ee2f0fab57fe95df096fc01fc054d1331088bb57efd36df79ab36686df37dcb9284943d2047df96bd7635"
    rawData=$(${MIX_CLI32} mix auth -n "$authHash" -a "$authKey" -p ./gnark/  -w "" -v true -e coins -s bty)
    signData=$(${CLI} wallet sign -d $rawData -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI31}" 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 1

    echo "transfer to 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
    transHash=$(${MIX_CLI31} mix wallet notes -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -s 1 | jq -r ".notes[0].noteHash")
    proof="1c33f86a705d387a191d598cdf5a4d97d839e49013ec8b73fb265a62bafc03fc17e9b34a36c6f015f49d1e944c945ff16bc3c7b4886591578c780fa829f140560fea4be7ddf3a98a5a620010f7a210456158e3f20b88a5a120980c8b960d15fb1c50a4d93b78c3cb3bc729152e1cd0b1574b3cc63a4ce6fb5f8e51828320df861d70f529eab34712c213a605dda2f05eda7fc4b0aefc99adb2817eb10548489d12591f64709a5ad3dd32a1ebfcdb70c50288af1185b4fa3e9a25636b9cc9df0c22880ab3541018fa25a9fb232ed4dd47e0aa92de0c4ee2f0fab57fe95df096fc01fc054d1331088bb57efd36df79ab36686df37dcb9284943d2047df96bd7635-172b267f5d1d1a629b364388c803405485aee2d5a86c4438dddba011c5ce77b42d7e5f12179642372f325af5d6fca46d169a2437e732461c4f4993a4ac3938a92c89038d4ba4434dacf7113c59e7be32a1952e0a452bd23621fa7159af74eb351a5ca048e591be47db5d45d54505c8924c7c1c5ec3d4df8e5b7c6e7f3ccf563d1acf12ad74b6985d9ba676a3db5abb69f0d4b687f8f11d019479ef22671b1cc90511e3fa819ad60f07d20329af3bfc0344f1a96b61e3f77c50885c5701be8e480482a58a06b3469b2d5821fca30e8d283c2e8e26c2106604709280e00c3df77f2ea8d9b4263ddedeab056d206c4665ed2caa28d4af67f09f32922239e1e5e572-1b66ecc95b8caf448d3bd235662c24a218a5c75ab42d41a8def3310424ac2d4515dfe9c9c77ad041e76ca2175d7e91de608cd39dea0194dc4eed7304d3126b230041306b465e74397b104db0303d87e0b02d052f6c7486d249a4dd5635f268ed1cd9180c6603351d1d74ed3f64283b0bfb75736d391892818f037b2549cb7e6d0b139b88edbf530fc3711e5c9235fa0e5b10af2fd03ac5993a6936d52ed51f250cfc84d3a379b9caa41ce9bf911259c1d7da73f8551032af8e3939c4d03e9cb125d0871b2e7a4ed474b87376188fd2fc463d0807d9bd701a6339c82489b523e022151dc964977616cc1d12176aad27d596b3e5caa4538d8716f93bfcaa2f2588"
    rawData=$(${MIX_CLI31} mix transfer -m 600000000 -n "$transHash" -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -p ./gnark/ -w "" -v true -e coins -s bty)
    signData=$(${CLI} wallet sign -d $rawData -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 1

    echo "withdraw"
    withdrawHash=$(${MIX_CLI30} mix wallet notes -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -s 1 | jq -r ".notes[0].noteHash")
    proof="152d166b3cbf1863f9da7ebb478a99d170544653d870a48533a48da8c484f2590fc795d9db67816b39a3dec2562dccfc5920b88236b766b8c3879e1d991121aa1b28801f92fac0597d397dc605b1e479742dbc3354bd1f33c8b52fac95cf047c29302a139eef9c948d148641e91d8b122785570d4a6c753c289e61bba96860c719fad87bb0954fc6ed3616f32af3f978aa26aeb03d123d312f569be901057c64123a2fed3890d5c0e59b3cff0e527dfe23b3f069227fd975a11c41133772cd311aa1b7ab4ff4849e01d7475c99e6e5fbc3fb0c1817e5f0a3ae98a7dc513af1e011aadc35e47b0359e39d89a50ff585b436fd27f5bc0bc973047c5b2fd52cd606"
    rawData=$(${MIX_CLI30} mix withdraw -m 600000000 -n "$withdrawHash" -p ./gnark/  -w "" -v true -e coins -s bty)
    signData=$(${CLI} wallet sign -d $rawData -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    hash=$(${CLI} wallet send -d "$signData")

    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 2

    ${CLI} account balance -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -e mix
    balance=$(${CLI} account balance -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -e mix | jq -r ".balance")
    if [ "${balance}" != "6.0000" ]; then
        echo "account 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k should be 6.0000, real is $balance"
        #        exit 1
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
    proof="12c840fed692d2f38e6d227e6b4dfe46f0107075035adbb035c652f9e00a662b2ce53546477aff0fc320dd8c505cab2861d494a173be2fdfb6ddb033542a8bb10826dce453a6aa7c728e821b67a7c0dac14dd76135c04d71985e6682a4a4e0ca033be19ce61562235ae9d62cb8409599331da4b11718169fdddde4b6e4642f60031f64e90a2c37f627059af4db30e92ce9d3f45a7f4683088f048ed2e34d331a0966758a7849abfe4a15d2f7df52a1be3270bd13b2c87d559cfd85c3e12aed911627413fb9e99be990e0b4ea9c61ef68ba655bf003cbc050d92febfc337c32b52e7c30942536885d0b62ea3b44261f557d12ef4c24cc3728d4ea91d39b19ebf5"
    hash=$(${CLI} send mix deposit -m 1000000000 -p ./gnark/ -w ""  -v true -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -e token -s GD -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 1
    echo "transfer to 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
    transHash=$(${MIX_CLI30} mix wallet notes -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -s 1 | jq -r ".notes[0].noteHash")
    proof="19af0372674c7a4b2ab9126abe97862a75c0359998c02846fc9e43c6a75ef93d100b12a6f0caefb55d45407b8fee49271a20be3fe1615ca535635cf58a20625922cb46d5f9c41054f6d85b4c7df1ac5d308d3c0187f2b787befe5fcf80f790351a200124cdde88299fa3b6bdf9e88969607ea4816066f51f05d77b2034b5676817a8ece901f5df1993b242d1cf2bfd0f91d4478b33ff0994d6ea70830feaecf40d12f855ccd1cf73e2c7e27f5cf1df357c25293decc91d8cc252bbb5e19cd3741a90834263874d77ca79e6e69e6711cc4d73f1045942448eb3600dd58a1e084b1960fe8bf0dbfcb54dfdbed79f3c3c9b5a22469a7aa4e195b247bfd26f1fca73-17b4593758ceed61ac96cba435d5bfe6b7cdcdc8e2e4deab88aeadf433e37542131a1c2f4f5383a57387b2de3b5767e91d301b9e42da4858513bf897c8441ce205284b694309a57dc8dcdea27922742c446c94413054de7451f6f21d0425fecb1ce6b7a52d044fdbf52f842cd8fb2c4539a0cfa7973293fa78b0c5e656df40650ad4d041ecc9f50c72c4de2db6f917a5e08dff6f59e7a48bdd9f74667193fa7218ec30659bc1409354f58ae3fc98be3b0810dc83a39a4ee84e586487747b62e61e5f07914c2e94bd86128944106239ba523377d87216cb2c91191e1ac3c1a7382a3e9ee7868da53d20fdef596008c4641b48ca2b052a3c569ff5a2364aab5aa5-234cd13277ef10f2ffe28cd56964411b70040e2bec022e00ac068c9a5c94ed510b2f04e1a4377efd2e83f543cdfdfae0b1bd70595b5b6ea0db15b6c161354dd90e0028c5ae061cb46e10cbdd58d7558798350956777981178a3a08ad4f8513af06fb315045f3f438358206e543ad7b5caa93e9b9d82f7452a7d57dbf48ebd4c42f8583c62a050bd2b143340bdcbeb641679a49a607ee13e7fde5d1eba0217a1b0384ff474d6d5f041b1eba24b70b8f807b4bc45a4751a7cba10d806fef6ba6f122fa1928ba8328e89f82c2281330460931a69324e14fe584e0a9ab851b069a6304e0178f04a0bc89ffb50e7755328ba8ef84ca830424bed6ef3a8142c567ff2b"
    rawData=$(${MIX_CLI30} mix transfer -m 600000000 -n "$transHash" -t 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -p ./gnark/ -w "" -v true -e token -s GD)
    signData=$(${CLI} wallet sign -d $rawData -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs 1

    echo "withdraw token GD"
    withdrawHash=$(${MIX_CLI30} mix wallet notes -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -s 1 | jq -r ".notes[0].noteHash")
    proof="2ae76d20889a6be8c9173bd7fa93883a4d3f018b01d083f80809e363b954fab626327fc97e34d830a8101fa6c6fce2fcbc613a8794776488a9acad7983af189f225e1b7284f40ab265092524c20219da32ccf4ac6dc2a8b2de924e43e8763c421bf55d90f6615e15aa228549b09f48f7072af6f768528da8a7b13611950b8e611513ec503c0ea0a0782106eb85dc0493a117f699465cff63d44beec2668ab368207bab40cf1867e11fbc947408339f04cd991eda2b1a250fb181080f575cfe650cc4e1144a6f02e997d49269c0406a373913bc2530fd96fa2856c459de596da71d12b0e8aef76d33d4e862879636c9f9d21ff49789d93f9f671013c9e0489774"
    rawData=$(${MIX_CLI30} mix withdraw -m 600000000 -n "$withdrawHash" -p ./gnark/ -w ""  -v true -e token -s GD)
    signData=$(${CLI} wallet sign -d $rawData -k 0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs 2

    ${CLI} account balance -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -e mix
    balance=$(${CLI} asset balance -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -e mix --asset_exec token --asset_symbol GD | jq -r ".balance")
    if [ "${balance}" != "6.0000" ]; then
        echo "account 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs should be 6.0000, real is $balance"
        #        exit 1
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

    elif
        [ "${2}" == "test" ]
    then
        mix_test "${1}"
    fi

}
