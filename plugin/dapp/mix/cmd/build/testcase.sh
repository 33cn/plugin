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
    ${CLI} send mix config vk -c 0 -z 2425d4f298f2aab6860bcdeb2fd5950fcb760476b2c2a32037c8737830ea1c8825c5d73667637ecc4a7643875d8cabd971d9c57c69bd818a679a40924e28609503dadd353cf404a397841a4f00692e0d5a34e746a89ddc9fef7552f55700fec416e28dbf86af56885f3ebc279cd49e730d3547b71e8e006589641e996203678d08224a611db2912584bbb91f5ffd71971b3772e9cfdfe4f864ef522bad85892b19c255f040a31f0e7da003efcda71002f6dd8cfca86e9bc300478eb668d6f15d2350316a826ba299708d083c9cf35e103c8ce1f34028beb5d428b6a433c225d41efaa5538a5eb57d0bbc234b9309fed04a3b1a43567eb2f19dcc096405b392562f52200536ec51bc85caa502b6d552884e7ccfceabf2de72c4abef49a996ea672e0dbb7da858f82593a0fd47dc06e4a7f53257da5f3baf8ff5cf23be5d9429cc260472e430640ccbceaba7f3741ea111eabdbf8432cb7c2f1352ca42d378829f14ae701f0e608cda5d6cfa32d000998fe9b8ef8bed2cd9eb44333fccf60503930aaf1a3b67e68f17a4765c1515f3b8b45aba8585fac2c813e038d79eef1bf6f20313ae84c8d629cf58e0b9f6b0db45b597e3edd1c3658040b73b595b8620365825ca691baefdadde2d5b379027da9baf3553ecd599738e24a3c329d3b8fa86221ba69978ad0c9622a5ee591cf71b8661114e31383912eb7e5e44391fce35071b06cf78d192c390b5134b0c781f7b4f30ef2668692409fd1d0927cf746ad1e1cf19d67dccde9f93589c7aa877804134d289cdf80c82efe011973a9e0a331114720000000319ba82d3cb41a6acf9a6e1b1d8322c4f7e83f2f7bbc56abcff2de3c8f922f6ff02260ebe94c2e33a47cddebfe0429f702bf4dab3725ee3df478c8a0fd45549252944b5ece51555379e390deb6f9b0fd4da42584f949ed03cf5391a1380efdbce02731d64b001e8e1d9c8c9420ac5ae85e85716b93ff0daf5c0bf16d9d761b57422b3f642892fe009cb5bf84685c52435b49bc4bdbc607ef191c545ee16890bae1dc75ec82ded043a6fb1690c63a0c03ee2d53f48478491c39f83d7e888698810 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    ##config withdraw vk
    ${CLI} send mix config vk -c 1 -z 04341f617c344849cd2291457b1abb6cc7d1059069dbdff84b043eab2b0cabd91a5a125dc49697c14fe7f0996b69444486d0deae624256e82b244fabab51a5471058f7cfa1368152789b07cd3600a48e2e5f8b3a782907a8aeb3f434c882ff1a10c214140874bb7327cc0f9d34f3897a70e8d60318ecf31dcbe98a5c30147c290dfa175600fbf5c5acd747cc82132de2bda9c9df76d3f2bde694ffff7f332058144235cb7fbf703b7f4916fed7cb8a441e4502e25d1da056a7e2d15d77163d501d251b98505bde1db434a3a2886b34ece7b3e3d97583a291ca8409809329c8992c7ef844f38bc9c47d628cde6000f0d3061fc3bff3675d25c90faa3d17a0c61615d06a2e8d60c0253bbe13480ac5881fa19eac79269b8d9e97955a0d2bce67092803b0d7737abfc5e1b9c9d4face06de02ea4c924626bb37fff6c292997068b11e8704ec01f56794e081c3fad8980eb32d8ebce45191e2697acd9fc1a94da8201854ce9db3852acf34cbab4d911ddf3fc08b56f1ef7de87721556ee85274ec9502f37e926a87c41aa10b4c8fd4b299b0df3afa29087b4efbada306768f3d62ce11db3856932992a4c6ea21872dbd87c9bc01cc9152d53e3ad16c7950524a71a91eee0463fdd8e71cf8cb5b91679aa0e36127744718b5ed9bc6a1845dfc27b58b2bbcb64a7a00f50a5cf1378f35c2259fc26797da2397e1719ee2fc64cbc14b1405b3308d1e872a4fbbd05cede9b135745d7c12508cb639cc000f552fcb0904db240fedbfe406e121c052818ed87c1c8a2ff0aefae8a13f9734e7a5a1cebd0aa4000000050cbeb6b76b9ee0ae6bc4795cfa69adfb289145d687aee63fb0e726374e563c0b14fbade78cb76081e79381f4d4f7eed891d21a6badada94acf8734659f9c353220b981ee9f22131a79da5a1b44cc020b5c04c6b559a1025d23f310db57a236b1190b2a64f4f2cae26d0dc5ecfd21c28429b047879c788e1d0c464949886610b200355f91905e96f30ad4de673c2d978557acd082fac9a0705b6f30b117f4f65612e12c8721123fa438f1dda1acabe06392c0823c5ec884e6ce55ccf450b7939a0358a5fd6aa2ff58e3f508293ff7b28185b3d82576c4bf151d7de7061717002411293b1d26321c8eddfa78eba4f58f50fdbd8de80736b599e00d786cf0004cd625bcb470a6e00cb84ab081086a8cedf20df2d82f0e8f3a06af30bc0a13301104153e0e9cc2161b927df5cb90b2fd22e4fbd86d80ded7281ca3fa789ff093658c -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #transferInput
    ${CLI} send mix config vk -c 2 -z 0111b7fcacb62f2c6d22a788b3779663bac96ce4a37c74361411968bfd168520240085f9456d59e4d3eac4569378370a0fa51d7c14a7b84028b25120ba0b15e920370f07be954a45667a37d5aa5a900a5975aebcd9f57e34fd1bbab7be602b910bc0d2aff6160363eaf3893d3ae40d08bb71d2d31b04c85f08182432c0bc231400c0e21208b91ef3e1b395633f91507a3cf30fd7e6ebb442873a7eaf13da896b0e5b85c970ca453b0e86347f89fc03c50d9e0cec71f0aa505669c2ca8437b6c91f98e2a55152148104094d614c358ee7c5f03c40406fc71d36f79a7cd248f1152e54492f273d19c5c148388e7d9ac11f8769bb2f12667d7cf1556add8df8407f0da42c14d84b6eb10e8ca08b26112111ca6c13b3ad1b73cf4b0e3014d33d83340e2560235ac74383053d4ccab4bd3903bcaad01bf04523a3687689345b7dad400ab68e6aaa7078e7acf0ca2592f7f27790ebcce57f008b16db343bb053e442dd20544a543b65aeb6c255a453496db2bc688ffd34e6d63ffa340054a131437afd001f5c365fef8d7488dc3375bccb92dbee1b948a5ccf83e5d067cfd4bf1d5bd124a965df3c28425ed0cb9e6ba4f3678f9347739b055fd5c8b666ff9dd4d2f9ca1512fc1dce41d07bcab1b3798881983953564e5696e28187750fdc6ba6ee6da5236cfaaff47e4210617af1d8b98519f72b29909b22ae91891c365ea784941a2f20105f97cf87ae0205100fecbb1b0ba7e8aa8cdf47c97b06de6e40d1ec44992e2d52c29662deb78ea528d18ce4db6f2d89b06ccdb0122dd135f0da6fffe46a4e0000000809f895acdaa98bd6fe5a1dfd25af68867ca1758e04a9068c98773302b483109b288992ac201ae69d5a1306fa43d72ed252b681ea9611e5a7653d7953e951f8a70bc937ea5513c88d89d0c483912a780cc5b0503726e7c3c9f556a789f74ac36c2f6116490c9cb714f81659a9589740a044804b7a2b3e61c62ccc7711e697151d1d6df790ca8180e434efac68e26db552fbed9fd5c1b16771a2d9889eeec330ff0bf99c4ea64cf6f07b63d7540279852b4bdaccdbb40d1fac96443ec3f7def90418ec6d52ea52c4c209b489bb49109bd9b07ac2b1df2daf8e85f04244e0ccb19f2b246b7fa7a1e885430c9e3553df5eedab9e3bdd8651fcf0e8c8695c8b47d3631811a6c5c20cc7ab7ce5421beadb183a779ee4433c8b24c6032ad8ac1660c47615ee86f4b8f9d741409b419dfe37511f7ed49b9ec0504b6526c9a442e7670ec90bf30b85fe97695c93093abcced2d0a78c16f13260818392a3a8b0378c986b46226443479cad09b6d3d37f110756438cc1e8613261eb492bf2237d1723d484bc2c5ec0734b1d23f701e09a16f2f3d4f0d3efe118961c883fce4c6f444dfa2ab00b75a3661e9efc083101e5262123170a57c067138f9abef693df1293cafecfe226a9e45b15d451d39f34e31a8d1edd3f4578fa446de2d08a5b879a315e111b7b25f4b695a4e777f19ab6c82de5160ae2ac79436af216919c571e1dcb09d6b7e2 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #transferOutput
    ${CLI} send mix config vk -c 3 -z 02dafaed1571b1234a592675fbb16a3140bbcad2e31e88abbb4181f6b70efab1147a7bc9ee30ce82abdc3001e2264ac68150c2cb0a5f1230082d9005b99d0bde3000ff64204ed9c4709902ff7123a217ae69e98783be72a02f28764ca609d43b2bc5a3aa647c8e41d9654cf679e7353e8a795395da1145129b275490e09d00f124d76a586b7fe8f473fcf0c1e0598b84df77da95ac1abef7a7258ff533a731b61e3839e9c5ec69fdceb75a5cc78f91e19579bbbeec02d6e0a59fe798fe0a6c242a231e0592d19146b5770affd5cbf490f40d4938a9d1722dd0b1b24f51d357d31e94b43be1f75c8ebc3035a5d2f31426f67f7ebf9495b1f6ad06f88a70bb445f0b0905bf5b0c1e32d501ae613e6010eb5edf25762e73edad9366972cf7cd0cb51743e3cc2bf843c47506b728384f42987cf68743cca3697541690237d030d2fe1aefd92ecab0c3233212658d6e53b523874ee4747eedc18e09b4e819ca3e6d320d24481d815bb1cc7b01f5adaea85d21470da3da8fea0f79b813993943908abe07a2394f6366a0f2f38baa8a8d52efc01b5da581c2e0aedb689eebc281a105e02e4360d8a2c92aaf49ce27ca9017fcc82ae3ff4bc3a05dd16bf2afe14b0a27651cab1cd341f4dadee74545cf9236c880aac2ae869c8fd2e16c851f83a2c6a1292debcffccb626179ec453cb7a74a8a9d9e384a479591035b5d672138531124fe0cf2a07696e57a4759bf8f81536f3e7f1153f27ee51694cb3d6e4e3668717f8b0baa2ab69d6ffb9e756dcbaab6bd60cdd0c281eff0233a237bfeede78ae97026000000060dafc96b88469d4f70792cf983536c26824d239173be2dbdb90bcce7a8be33bd2cc827569aa1aa5059300aab06a4c6011ef5266b79e1684797ca864873b918e50e46e4d44f7bed7e4875acaee9e0febe577fc69a0c61fd36e5abe0a88d04ced002e9425d72db8b6f6d666ecb5b2ab661e73d916652bce28bcab3ca76ae443f2c00ea3b57437f3dd22fd9335c72c7e6f609100232088c0942158a9abfb847ea0a053e959665156da6f28788400019b7fbca16f78eb54b1540234f6123672508650ddd4381720f50561be0527501aeae35786a819a0f5fcad7588e8fbee5e7fec90329d2a7b10d861f2fb45d7c0aadf79da21f7a4c081c21615850eb906a32452114b95a5a694fdad8e42810345b9a9879dcd70a390f55475ba028d9ee3a1dd63621e287e9e21a4a4bb6267a86bbe3747c4bf6764499b518ec4bbf469b7ced235b301e97f1e362bbbb1c7beceb208efef310933beaae0efe3a936ecba9aa2daca31ca3dc9691fbda151e5a8fac40b4441ac153fd8c523f256c632a3d29b7c555e1 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #auth
    ${CLI} send mix config vk -c 4 -z 0db4b7e02581d2b361d0f411edb9fd8dc0d6c58ae10134ea425941d028a9759307865795cf7e24688b6d70eef4f320ba042444ea6a211b20dbda898ea41d4c781a56a244f52282f91a4a5b13d26b5174514ed8385d51ff4f56d9a726a3276b142588ee88b7686ebb0cef768200cf0edd92599966b5d854752e3ef43bc86a3c662ce168aa0831d262ffd7b15e6cc6280661cc6b9833124b0ea0e3ba5510bf0cbe0dc82b244001da5304e0c420d46b773caec740211b7382aa10830c5d6f906eb81d9014bfbc2664f96b0c0bd222e4b55635dc6992201e8dd74b30ec0b538ae07716a4b545a0b06d081265a600784182952e6f37fafa05454c29fc41051901fc1d1b479189f10c8c60d0286373de73f6ba33b9c08ac62e6a580f56ff6901f9d4c40aca9390584cbc7e4a8b23b7205cd52cc31b401b9f65eea5c19de20d70afa2ff1f7b469a4c1b131371810c6f38f680e2e069fb30ef741663e191c62627d407c2096411a13627346c153571fbbac0d49576bcfb0fcde68d302e1b73fbb92469b90f03e8da69e6e84a33941781669461622eb3968ac5521265a71b7380dd2f89822fba5618cb49ce1cb425a3fee58dbe0cd338fa0227fc69d53ac729c4ece650fd2af1ed6835eb7e12ae31489dd44ab91d9dbe1137b23c6a34402dc60058aaf78508304b7d4c5289fdcf5ae1d2ca58c45f3c831c27d7e4bf84039972b159b131e12dcb73372759f7ce2184966c8dcfd0a2dab1a69cdf49a6569c7233a1681853782a2abd7d4b60b0008c98ba3c12de243dfd9fb95f2007b89639c4809e9dc35cc90000000428dfa273594dc969a6f93f22efb4fa97bbb65215818ed1f643ff765b7cd8bdb51d3d6e32281b33e1a6e80a7989334cc57da71d5e9238c7adf95fc0516da6036b10956bceb8e1966049bf3ba1a0f1aca457054914007465aef34c6441a639b1d205e60b4932efdf06fa9b0842fe3fd329276d4e134339aeeb09c352962554f82f2599d8ac08f5972603500f9002fae1dfce5a54b7fe62d8ff7297501e31559e031b40ae3f6ef4316856c242aafd822103b1ce20d43a0470e5a66f4bc433106c0f0d65e1dab382c77bb74bcc91cdd7a26b36cde96094699e40c0e5a3a37deda2d80d5cbcb14bb43bbe2eb3080c1901a42b3e4343305f13bc02b430e3ec97c92428 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

}

function mix_deposit() {
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
    rawData=$(${MIX_CLI32} mix auth -n "$authHash" -a "$authKey" -p ./gnark/ -w "" -v true -e coins -s bty)
    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI31}" 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 1

    echo "transfer to 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
    transHash=$(${MIX_CLI31} mix wallet notes -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -s 1 | jq -r ".notes[0].noteHash")
    rawData=$(${MIX_CLI31} mix transfer -m 600000000 -n "$transHash" -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -p ./gnark/ -w "" -v true -e coins -s bty)
    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 1

    echo "withdraw"
    withdrawHash=$(${MIX_CLI30} mix wallet notes -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -s 1 | jq -r ".notes[0].noteHash")
    rawData=$(${MIX_CLI30} mix withdraw -m 600000000 -n "$withdrawHash" -p ./gnark/ -w "" -v true -e coins -s bty)
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
    hash=$(${CLI} send mix deposit -m 1000000000 -p ./gnark/ -w "" -v true -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -e token -s GD -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k 1
    echo "transfer to 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
    transHash=$(${MIX_CLI30} mix wallet notes -a 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -s 1 | jq -r ".notes[0].noteHash")
    rawData=$(${MIX_CLI30} mix transfer -m 600000000 -n "$transHash" -t 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -p ./gnark/ -w "" -v true -e token -s GD)
    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_note "${MIX_CLI30}" 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs 1

    echo "withdraw token GD"
    withdrawHash=$(${MIX_CLI30} mix wallet notes -a 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs -s 1 | jq -r ".notes[0].noteHash")
    rawData=$(${MIX_CLI30} mix withdraw -m 600000000 -n "$withdrawHash" -p ./gnark/ -w "" -v true -e token -s GD)
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
