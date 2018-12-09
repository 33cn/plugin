#!/bin/sh

docker network inspect pbft-net &>/dev/null || docker network create --subnet=172.172.0.0/16 pbft-net

gnome-terminal --tab --active -- sh -c "docker run --net pbft-net --ip 172.172.0.2 --name replica-1 -it --rm pbft 1; exec $SHELL"

gnome-terminal --tab --active -- sh -c "docker run --net pbft-net --ip 172.172.0.3 --name replica-2 -it --rm pbft 2; exec $SHELL"

gnome-terminal --tab --active -- sh -c "docker run --net pbft-net --ip 172.172.0.4 --name replica-3 -it --rm pbft 3; exec $SHELL"

gnome-terminal --tab --active -- sh -c "docker run --net pbft-net --ip 172.172.0.5 --name replica-4 -it --rm pbft 4; exec $SHELL"

gnome-terminal --tab --active -- sh -c "docker run --net pbft-net --ip 172.172.0.6 --name replica-5 -it --rm pbft 5; exec $SHELL"
