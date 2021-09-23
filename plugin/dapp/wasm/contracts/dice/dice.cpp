#include "../common.h"
#include "dice.hpp"
#include <string.h>

#define STATUS "dice_status\0"
#define ROUND_KEY_PREFIX "roundinfo:"
#define ROUND_KEY_PREFIX_LEN 10
#define OK 0

int startgame(int64_t amount) {
    char from[34]={0};
    getFrom(from, 34);

    printlog(from, 34);
    gamestatus status = get_status();
    if (status.is_active) {
        const char info[] = "active game\0";
        printlog(info, string_size(info));
        return -1;
    }
    if ((status.height != 0) && (strncmp(from, status.game_creator, 34) != 0)) {
        const char info[] = "game can only be restarted by the creator\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (amount <= 0) {
       return -1;
    }

    if (OK != execFrozen(from, 34, amount)) {
        const char info[] = "frozen coins failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    status.height = getHeight();
    status.is_active = true;
    status.deposit = amount;
    strcpy(status.game_creator, from);
    status.game_balance = amount;
    set_status(status);
    const char info[] = "call contract success\0";
    printlog(info, string_size(info));
    return 0;
}

int deposit(int64_t amount) {
    gamestatus status = get_status();
    set_status(status);
    if (!status.is_active) {
        const char info[] = "inactive game\0";
        printlog(info, string_size(info));
        return -1;
    }
    char from[34]={0};
    getFrom(from, 34);
    printlog(from, 34);
    printlog(status.game_creator, 34);
    if (strncmp(from, status.game_creator, 34) != 0) {
        const char info[] = "game can only be deposited by the creator\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (amount<=0) {
        return -1;
    }
    if (OK != execFrozen(from, 34, amount)) {
        const char info[] = "frozen coins failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    status.deposit += amount;
    status.game_balance += amount;
    set_status(status);
    return 0;
}

int play(int64_t amount, int64_t number) {
    gamestatus status = get_status();
    if (!status.is_active) {
        const char info[] = "inactive game\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (number<2 || number>97) {
        const char info[] = "number must be within range of [2,97]\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (amount<=0) {
        return -1;
    }
    //最大投注额为奖池的0.5%
    if (amount*200>status.game_balance) {
        const char info[] = "amount is too big\0";
        printlog(info, string_size(info));
        return -1;
    }
    char from[34]={0};
    getFrom(from, 34);
    if (OK != execFrozen(from, 34, amount)) {
        const char info[] = "frozen coins failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    status.current_round++;
    status.total_bets += amount;
    set_status(status);

    roundinfo info;
    info.round = status.current_round;
    info.height = getHeight();
    strcpy(info.player, from);
    info.amount = amount;
    info.guess_num = number;
    set_roundinfo(info);
    return 0;
}

int draw() {
    gamestatus status = get_status();
    if (!status.is_active) {
        const char info[] = "inactive game\0";
        printlog(info, string_size(info));
        return -1;
    }
    char from[34]={0};
    getFrom(from, 34);
    if (strncmp(from, status.game_creator, 34) != 0) {
        const char info[] = "game can only be drawn by the creator\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (status.current_round == status.finished_round) {
        //没有待开奖记录
        return 0;
    }

    int64_t height = getHeight();
    status.height = height;
    int64_t random = getRandom() % 100;
    printint(random);
    int64_t round=status.finished_round+1;
    for (int64_t round=status.finished_round+1;round<=status.current_round;round++) {
        roundinfo info = get_roundinfo(round);
        if (info.height == status.height) {
            break;
        }
        if (random < info.guess_num) {
            int64_t probability = info.guess_num;
            int64_t payout = info.amount *(100 - probability) / probability;
            if (OK != execTransferFrozen(status.game_creator, 34, info.player, 34, payout-info.amount)) {
                const char info[] = "transfer frozen coins from game creator failed\0";
                printlog(info, string_size(info));
                return -1;
            }
            if (OK != execActive(info.player, 34, info.amount)) {
                const char info[] = "active frozen coins failed\0";
                printlog(info, string_size(info));
                return -1;
            }
            
            status.total_player_win += payout;
            status.game_balance -= (payout - info.amount);
            info.player_win = true;
            
        } else {
            if (OK != execTransferFrozen(info.player, 34, status.game_creator, 34, info.amount)) {
                const char info[] = "transfer frozen coins from player failed\0";
                printlog(info, string_size(info));
                return -1;
            }
            if (OK != execFrozen(status.game_creator, 34, info.amount)) {
                const char info[] = "frozen coins failed\0";
                printlog(info, string_size(info));
                return -1;
            }
            info.player_win = false;
            status.game_balance += info.amount;
        }

        info.rand_num = random;
        info.is_finished = true;
        status.finished_round++;
        set_roundinfo(info);
        set_status(status);
    }
    
    return 0;
}

int stopgame() {
    gamestatus status = get_status();
    char from[34]={0};
    getFrom(from, 34);
    if (strncmp(from, status.game_creator, 34) != 0) {
        const char info[] = "game can only be stopped by the creator\0";
        printlog(info, string_size(info));
        return -1;
    }

    if (status.finished_round != status.current_round) {
        // const char info[] = "inactive game\0";
        // printlog(info, string_size(info));
        return -1;
    }

    if (!status.is_active) {
        const char info[] = "inactive game\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (OK != execActive(from, 34, status.game_balance)) {
        const char info[] = "active frozen coins failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    status.is_active = false;
    status.deposit = 0;
    status.game_balance = 0;
    set_status(status);
    return 0;
}



gamestatus get_status() {
    char status_key[] = STATUS;
    gamestatus status;
    getStateDB(status_key, string_size(status_key), (char*)(&status), sizeof(status));
    return status;
}

void set_status(gamestatus status) {
    char status_key[] = STATUS;
    setStateDB(status_key, string_size(status_key), (char*)(&status), sizeof(status));
}

roundinfo get_roundinfo(int64_t round) {
    char round_key[32];
    gen_roundinfo_key(round_key, round);
    roundinfo info;
    getStateDB(round_key, string_size(round_key), (char*)(&info), sizeof(info));
    return info;
}

void set_roundinfo(roundinfo info) {
    char round_key[32];
    gen_roundinfo_key(round_key, info.round);
    setStateDB(round_key, string_size(round_key), (char*)(&info), sizeof(info));
}

void gen_roundinfo_key(char* round_key, int64_t round) {
    strcpy(round_key, ROUND_KEY_PREFIX);
    char round_str[20] = {0};
    int index;
    for (index=0;;index++) {
        round_str[index] = char(round%10) + '0';
        round/=10;
        if (round==0) {
            break;
        }
    }
    for (int i=0;i<=index;i++) {
        round_key[ROUND_KEY_PREFIX_LEN+i] = round_str[index-i];
    }
    round_key[ROUND_KEY_PREFIX_LEN+index+1] = '\0';
}
