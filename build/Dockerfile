FROM ubuntu:16.04

WORKDIR /root
COPY chain33 chain33
COPY chain33-cli chain33-cli
COPY chain33.toml ./
COPY jvm/Chain33Loader.jar Chain33Loader.jar
COPY jvm/Dice.jar Dice.jar
COPY jvm/Guess.jar Guess.jar
COPY jvm/j2sdk-image j2sdk-image
COPY jvm/jarlib jarlib

ENV _JAVA_LAUNCHER_DEBUG=1
ENV GOTRACEBACK=crash

CMD ["/root/chain33", "-f", "/root/chain33.toml"]
