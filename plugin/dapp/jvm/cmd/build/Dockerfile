FROM ubuntu:16.04

WORKDIR /root
COPY chain33 chain33
COPY chain33-cli chain33-cli
COPY chain33.toml ./
COPY Chain33Loader.jar Chain33Loader.jar
COPY Dice.jar Dice.jar
COPY Guess.jar Guess.jar
COPY j2sdk-image j2sdk-image
COPY jarlib jarlib

ENV _JAVA_LAUNCHER_DEBUG=1

CMD ["/root/chain33", "-f", "/root/chain33.toml"]