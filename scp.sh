#!/bin/bash

hosts=(1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17)

for x in ${hosts[@]}; do
    scp -i ./ycc.test.pem chain33 root@ycc$x:/data/
done
