#!/usr/bin/env bash
NODE3=build_chain33_1
PARA_CLI="docker exec ${NODE3} /root/chain33-para-cli"

#docker exec build_chain33_1 ./chain33-para-cli coins transfer -a 8 -n transfer8 -t 1CvLe1qNaC7tCf5xmfAqJ9UJkMhtmhUKNg
#docker exec build_chain33_1 ./chain33-para-cli wallet sign -a 1CvLe1qNaC7tCf5xmfAqJ9UJkMhtmhUKNg -d
#docker exec build_chain33_1 ./chain33-para-cli para privacy create -d
#docker exec build_chain33_1 ./chain33-para-cli wallet sign -d  -a 1qpAv7H4C5JBgVQffDRbQKti7ibdM2TfU
#docker exec build_chain33_1 ./chain33-para-cli wallet send -d
#docker exec build_chain33_1 grep --binary-files=text -nr "Failed to do util.ExecBlock due to:" /root/logs/chain33.para.log

#0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b

function buildAndSendPrivacyTx() {
    rawtxdata=$(${PARA_CLI} coins transfer -a 3 -n transfer8 -t 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4)
    echo "rawtxdata:${rawtxdata}"
    echo "                 "

    signdata=$(${PARA_CLI} wallet sign -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -d "${rawtxdata}")
    echo "signdata:${signdata}"
    echo "                 "

    priavacyData=$(${PARA_CLI} para privacy create -d "${signdata}")
    echo "priavacyData:${priavacyData}"
    echo "                 "

    signData2=$(${PARA_CLI} wallet sign -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -d "${priavacyData}")
    echo "signData2:${signData2}"
    echo "                 "

    send=$(${PARA_CLI} wallet send -d "${signData2}")
    echo "tx hash:${send}"
}
echo "Build and Send PrivacyTx and get tx hash:"
buildAndSendPrivacyTx