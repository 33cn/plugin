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
    ${CLI} send mix config vk -c 0 -z 0ab151c3f5ab4f02641e312dedc949fc34cef739ac7ceb72416274e27871a2b909a9e5219bcb43d1b13e87b9eea296625b0e3289b43b03cf70e3f77d7ca699c322f66bd31c2dea48018df71f2f8f5c94da9effe18563a29e74a3a1de6b4756521266aa67b4f7c9fc50fdf20fa99fce41247dde6d98501859c11dd4b72b8370d41dd5574d8f9d76ffa5389108d2577ad7fd0b934eaf44c006f1a754c7c238efa01511960ab2e37126bb6708397376fc137b384cf3f19099d4528dd41edb217eef22db1280cfae1949e5638870aaea06b90e322a4750ab3bea50d7df3f390ff1100e5db695c47db343ec04f8024f125ab210562123709852e53f1cf1ce8a2bac7800e049de7ed1196afcecca91b3163dff1f9dd13cba199c1c8a2ba12d7ac2860a12f967926f3d79e2a0d83251abeb84015b85643351b8a72c03d58c114f0c86fb2074307017b63aab2826bf02f1c08021abf46788da44bff7808ae5c80ba493f6245238707e3322f28ba7c48fe122893f6e3d76ff8bc14d2dc24688317385491d3055d5c0937f9db16495652ed40ae7bdfa223e5e14c59cdeace3f0caee452e0920096addf2cea748671bc8b2ce6eae8ae59af10ca4d202cbacd3ba269e0a5da72b99033697ff84a19789545d3fe8d7e3f7c4e26dfdb6d5db76742b89ec62fe5306fe23fa60b4a0ae7da5bedb1aabf42d483d534962122d54c6ae5853d121853c28eec11fa4c59cc4dd0787b90a3c0089fe81629722fb1a54b2043a0fc3171c650df2d4ddf86d70047524c5d856810927840656fe45b4ab9e676ce227e71273c0000000032db890127ff54710550d5af52986938be708c0ade9e8f903f67570de4a00e49411ac3492154d2ee4286b29f81d6c7a604df8b315065f9bb94eeca31ce6143b9e0c0eb9fa18c283c2145c25840a67b882dc5ca8e9f7e4295398cb20306a5d612714c40d52ad2296b2ae0e930d84cc105d91d3da03ddacc94018084eea36c7abd12b4765a421b2692bdab5c30abbdbc936f9ec6186d73acb473ea36ea0a57127b42ca754c1ffaf99a706e4830212afda0bcd7bf8f0a51ddab790df3e651000841d -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    ##config withdraw vk
    ${CLI} send mix config vk -c 1 -z 2ec2d037b66a1be7b07a8acbdc6de9fe02d9c6f3f9c620d84f1f07482da694131e922a695e61e4c976ce197ad77549b30605bb02540bcae0a05a1fe20212790e1396bea4c567c18b31f274072fc78283db4f0af154f5c711b5900060756aec171076ef15c3b0603969f70d4c99ada36e26c7d5b77002ff721023fd2edfe0b2a21569e07cae10cbefeca4d08b85da4d430b075ae7f443a473b7ec98382edf694e066abb75bb76deeca2dc1c8373561b3709c3ef4a5430a99c612bc70bb9a8a1470f18c0791e1797ee0014e8090b2e8c283789868fe93c9f5c4270bcac988a22be28df0e0cd8499778322fc465f8f08a71b8d242f47cb05691b8220a533be501f02741de93f1c787b9607811521a7695d9f3b56f77264ce3b2912f66d41605f9af1e671e76fb453a96730c573cc4672edd4f1c80677c5fdf6cd5e51f3a9ed8e5ab2795a7f7e51cabacd4795775dbc4056f15db1f790cdcd159cf27f3929de817a10806e21eb0634175fbf565a6180b5cfeca86e891e796cff5cc93f7317b44aca0164a54562e587c98cb00af51699342078d665b7cc9e36ef484890bec208a538e1a8a3d0057ddb19d5f6764a9b1b06eb9e4fc8bb8bdc81851a9e5f6e1dff78bf71b5038105a8520169948171e63054de847efd960575e91defe822a30728eb39f0bc741f5b129490d2b4c9f21121c58a82a641018c11c4ef383c58616bf334be30939cb14e8458f2c637400a5238804cae6beb35e4779b032334a86066ccbe0432f8a7e069d91d92a00129477c10cf262af0047aa74576efad1195d52e6f5f30d000000052185f49e302904514d4903b9a92b14b842164953dee02eb418a8df75a4e720fc23a10a0934c53319891fc59e233cb4c87cfe05817069ba4eacf0e0c126ad463f20bf9ee9d17070c41d18a95617108568094c74d328f98445e451fd070317d15722c4e7837112e3383c8eeba2f5c0936a74060a65df281be970306fc5e2e609c72bc86aad57eed36817e377e3c442d1c4bb5611e0b737caee7f67bae6946a7c1b19f77daf8d47d2d08bf72105ae2ac402bab78fc103ae265733fff2616dcfc09816b3c05d350fcd8d1c5cda7b1c743c43520d276310866bdde1dd1087073f86960fafbefd05d08722fad1a0aed57bbf9dc3a97cc13d4352f752985f364732c4332602b423b84741578880f7698d6b88997a6517e5f09e849e947a4fc35ec2f47a2a48f81533aed4a1460c133204ead8b72a0ef6bb10ccab4370fa68217513537a -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #transferInput
    ${CLI} send mix config vk -c 2 -z 051c202afa691740e646a47155aeac1ae11d3ee4d48de8b79ae734e12778ca800b1e907766f83a5b6c6ef9388cc4a308621405100edfb3505356d9440c51ccee1157fb810e73c92fba52b31a2d14ca659d1219d81662049ca4b07c9d5cd8089a0f37d9e2242f3d7a5400f6d8f2bc6e14600f94e215535cc04579a95b5034f2232d07eb10f6c423d1190da1869d12d008b0f315d80508c166866873f13c34b30e009bd4a44819d3e59a12ac3176ac2e38a230c0bb560f11f16facf18d06776d6d1364e734e8eb6c94fbfe60ef7c772155cb753adbfe7e41a04f3adbd19a83ef51152a2cf2796445d81a85e308551a55830617ba7721225817a54a88aa9f5805e40b968e5db1cc5f923bedd53937e8537137e2b04dff00a9cbb14840061e9ca72d226cd17379427435f1cd0814ac065290e8f86cd5da8a52b9d284445c4220c1251f2da600177ad117eaca24ef9f856448879e957eb9a7ad3bc44c3bea4618bdf324d07d82b550c2d041059e69c38132240a89f9012fea639ac95c7b4332a055162ed54236ffae5de62a0c005889f2e2b95959dd67ffce49933ccef6fe5c5586aa2f9ee1d799bf9ab26e5e3a6b6413e694ea87d2a64da143b8319f6885e569d4d2053eb82e7c1fc8b041df190139e8b3bdc46e2e43864469269b76fac014bb5e930d7389657eafee1564f3fc5103e8c910e49d25248500edcc60f9c5b588722e980d818cbb640a79fce4c5dcad482e3c59c9ad657d0e0b46c0663c9639918c999b155a86125433594e9043f9adda2d3427b48ab5585e901ef855ce21a390c9e4a3000000081cd2a326d969e1a6a79bc34e249d8cddc713347c0f55e8dd723d4ca4b3cb7afb0d01e639e44717003f53934678b422ff82acbceddbbaa668e810d3d6e24f6d321c1d8197db7e0080bdbb51f5e7bbbee992ad691d3049e1dba5808bca929427111d050684da6bf46717d1af9eda2296a06e86567769d099fc22a72a85c66eafff17e0a9db9f6502c74ad802eeb65b4250765598bd6b9a0aa0feec4fe00b7ca2230d7627b067e9b589a5e2b8f8bae8ce5dc32891aeb4c59ac220bcda1d92a38942134dd434a78f3e018f4d53d96a5efc20f03137f75e543044789e321336efa2b116cff3644e3e6c4c059da67c39b419d222cff307a0b144176aa7ff0d6d64221c2de30c6f7c679a3f5f98321f67a045f4d1f02ff7e764456cf6229d27f6ac4322222d503fec674239e329f946b3a962ebae9f0d65246c354d36aaea58bc74f91d178c223006f0f12316c364b1c48e7cd20dc08bf4434c6ec40b723f8d7560e5651a902177973149bdddb43d18ddd14bbc2a64ab3e3398dcba75ee1d39840d347422716359ac1db91cf14d1948337f776f7b5a581309b0fc728ad2e34e902cb5bf2b18a357de815af71028c8487059aa4b3148c71c2ed68e4add8a9953f817d0f21bbe6ef3224ef1e9187e2caf5d3b28d5e0db0a877361ca6eff76796ae424f87b169f9e28c3d0c05b8372d3b3d7db413756fac99ab4ad8b966d9f231742e64408 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #transferOutput
    ${CLI} send mix config vk -c 3 -z 2646cb1c4bb8a140f415eea2c77a9daf8e484f2046e2730fc65954a74e360c1e267a8008649752e825cc003ba1a09ebf8540be49ceda6d2eaeae03b5e66736ae11bbb7c2fe42b0547e9bd568bd1511bd27489263b454c86e24ca42b81880f2622c38407f9a3e30e64dc3b76d56f0bf4288cbe4750daa91582525e7ed828cb2412662b76a46c868254530f2efc58705f171cff1c6d030f6d5e8fa87070d85373711317064e1ed6e1acb1fdf58c7801abcf953214d4934b9eba7c48510d90225810a3e018d98bd7c0fd0f5b0ebc0e9536362783fada89c97f744ca531eef0a6ed0072c7b22dae0e3eff404a49a6633c4050792ecb09930fd614b60fcc4b356edb701f8d766ae2e36f7c1e63622ec5c92ba0ef273199e52085228edf5108521545b12acb48ef9f5ee1adbc7bbcda6fa55fc421d8c62357d5814c1320fba2d4ed82a1b83f4596edc9fd297ac836725d635e101d022db2056d031e2e088f7ecd32b442beb194682415b6c76654b76dc093646b43aebfb0d09ee75882eb1a4344dae0603b2c48c3e321fede6bcaa60ca27497e8a9c97a95bfbcc76ba3303a7f1bba2441371423823ca6c7b82397ee678f57e55ae26b127d7f6a086266726c1b057576b12c173471d0ff82b08eb1937f57f50630a1f145fde62e6ba80a69511565b99711cddd6f348cc530dba26f52317d3bae46cf1b7a60bb55e6f1d5f4303d7edc53914d674be4814a22eb65a1fe2831ece438a31209b0b1cbeba8f901b48dae4de770201ff2a51398254d7d8c1f459e72a336221c663677298b0a58fb3972dbc7b7b000000060ec29eca4fa56da19081893c9bf504b3117e4a19eae4f59ac3d73d2340fced9518a151d7b1c437cab7afcd3573660b707ae2cefa7f976a2d1d97514515ea78cf1515411d1baf74c4da25bddd0b403824031a6d3e84ab92fcb79f1d4df521d0af2c3c6456ca5fa049d406c71f54ab6a3fde440031fa6cdc3790b736e237d3c7cd10d3315f7fc93918996f925652eb843e2c0c00cfe66615f6c1fa1074069dcd7a02c99dccbb857786c47a6ff1aec76803960a255cfee5987792b94195b7b8666b0738e212f51c1842e7c168c60a8989e3dbb381446c71da439e22570fae2f93ab2394f2bcfa252adfede9d3e4a189b4d215ad507c0ae291f5dcf8e2892923cc771ee29277c8026dd46de818e6aed6db2e28d440a52cb52b7b4ce376e1b80a078f1f8c9fbf6d1ac90c65ecdfd0d9e00a45eac94074e89a2581f21bd9958945c72307d07433866715de28649aceb06bb22ca4536cf5c17f61a8161b68a638701b100da04c352d930e42d62c9cfa801ec9c56ec1755df8df48d35660b5764b5d0be7 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #auth
    ${CLI} send mix config vk -c 4 -z 064e30b2d67831d68cce9f2309eb09eba041a948a5b87f1fb1422be921cb988d16150150f1ce6b3c05c708a84ecbb981b4bcd0b7678d9825234da26335168b8119e9f7dff32dd0b2c8f029b1326659721c17b757fb43fc115663dbe61f0517d90ecc2f9ee148756ae207dea302394aee1ae2da9c2b2b53e57f9cc2aed7b763eb23f7f286446f938b3a30ea7f4d22696367b2e58ce34c048a6c735ba340f6cf7104eb6dcd58210758e6c79a07396725ba3b31c69d42bb65c5ccab75f60516c0ad03ca1d155048fb23dd286651eea7aec4afd3d573722d708a2872b45a70de81782261cf19f4051f6d73d5f997289f7f97be0dbf8975cc50d7619b42e0c883d4c313264e46631f13bd49b5a4ad955575f58a54932cf36d89047cd8cdfad8953eab1c500bad362114f0be85a742dd09ffeab5fa16c9ff81e3d7d5d99debb2e7f25c2d2c3431da5d210c49610f0a9731d1789810956c39d0512a25ded2bfafdbce720aabf526c01c0b181e2470fba30393785af0dc2fd21b0e4e194f9fa509a86efb2ff283e0969e9722aaf523fc7d1243bbf8c446e3006452d15f95f9630a409e6201d8af6f9919c5861a6fa6d5d55a986790cd1fee2aeedc26af16cb4477534c5a09d7910721123fdc1fa98e34d4cf67bfcc159a234e0d006202a35ccaa82d83fa031656baaa050e0ecd0aed0762347718780f3be923a975d3aa041b51cb48045e0b626704c4f975399477f33bf2e5c0200350a19aad1f05f39c30221b54a47593054ae06fd883518e0b3658b54142c64a67f6e1152ca1a93215d8c71659b7caff000000041c8e2e41136855709f53ed332d2a2b3bf793758a11f058b676cce8176051db5b1ad307808cc54fbd897eb68b9ff83266618c1172be210262d2146db0db7fd7391399b1df6dda25cadb729c83ee3f9ceca02c32f28be603634c597ca1c67edadd0e6a7a30fbe46b57a8742bcb0212ec2fc7eb0808543bd70f2de7dd9a7c8904f20f72dfe65916f22382e01f757084bdbd7ee33ea4eca28c598b51bd7cb191973602b091483ee1ed95d9d0c8e28c79a0a4ba627769b194d0c0e6b308031367cc060cdbadd39510b391adc885ff4123f0d9f4c6ccfb23fc9347bfc09a5144dc38de098a53cd28ded02023dec1baec74dd7005a7f82acf75f3eef65b3bac64fe1136 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

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
