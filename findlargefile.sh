#!/bin/bash
#set -x

# Shows you the largest objects in your repo's pack file.
# Written for osx.
#
# @see https://stubbisms.wordpress.com/2009/07/10/git-script-to-show-largest-pack-objects-and-trim-your-waist-line/
# @author Antony Stubbs

# set the internal field separator to line break, so that we can iterate easily over the verify-pack output
IFS=$'\n'

# list all objects including their size, sort by size, take top 10
objects=$(git verify-pack -v .git/objects/pack/pack-*.idx | grep -v chain | sort -k3nr | head -n 15)

echo "All sizes are in kB's. The pack column is the size of the object, compressed, inside the pack file."

# 46034   13074   650902c0f310f2f64e9ae5dfda35656ce8dadae3  chain33
# 42998   12325   5e8c0c3e6bc33034fd804d4c68b589c1eb66804b  chain33-cli
# 30440   9801    d2d4a3aa838d85738e27ee474c785331bfe8e81c  plugin/consensus/raft/tools/scripts/chain33.tgz
# 21292   10288   1e21797c3af8a4385c9169570b9b1c4072d7b3b6  plugin/dapp/exchange/test/cmd/main
# 4834    1113    9ec4f3d49403e8b9dd46885031a92e23af3828b9  vendor/golang.org/x/text/collate/tables.go
# 3468    1767    825659f96c308cd79ed2b32860d45d510dff6cce  vendor/github.com/33cn/chain33/doc/golang/Go的50度灰：Golang新开发者要注意的陷阱和常见错误  .pdf
# 2556    863     9df156a7f0c9570431161873e3f605c7e4bb7ba9  vendor/github.com/haltingstate/secp256k1-go/secp256k1-go2/z_consts.go
# 2432    2159    0304d27f62317d2216c3288047a1a2a8bf37d94a  vendor/github.com/33cn/chain33/doc/PBFT/pbft.pdf

history="650902c0f310f2f64e9ae5dfda35656ce8dadae3 5e8c0c3e6bc33034fd804d4c68b589c1eb66804b d2d4a3aa838d85738e27ee474c785331bfe8e81c \
         1e21797c3af8a4385c9169570b9b1c4072d7b3b6 9ec4f3d49403e8b9dd46885031a92e23af3828b9 825659f96c308cd79ed2b32860d45d510dff6cce  \
         9df156a7f0c9570431161873e3f605c7e4bb7ba9 0304d27f62317d2216c3288047a1a2a8bf37d94a 0304d27f62317d2216c3288047a1a2a8bf37d94a"

oversize="false"

output="size,pack,SHA,location"
allObjects=$(git rev-list --all --objects)
for y in $objects; do
    # extract the size in bytes
    size=$(($(echo "$y" | cut -f 5 -d ' ') / 1024))
    # extract the compressed size in bytes
    compressedSize=$(($(echo "$y" | cut -f 6 -d ' ') / 1024))
    # extract the SHA
    sha=$(echo "$y" | cut -f 1 -d ' ')
    if [[ ! $history =~ $sha ]]; then
        if [ $size -ge 2000 ]; then
            echo "over size file = $sha"
            oversize="true"
        fi
    fi
    # find the objects location in the repository tree
    other=$(echo "${allObjects}" | grep "$sha")
    #lineBreak=`echo -e "\n"`
    output="${output}\n${size},${compressedSize},${other}"
done

echo -e "$output" | column -t -s ', '
if [ "$oversize" == "true" ]; then
    echo "there are files over 2M size committed!!!"
    exit 1
fi
