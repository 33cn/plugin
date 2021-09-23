struct roundinfo {
    int64_t round;
    int64_t amount;
    int64_t height;
    int64_t guess_num;
    int64_t rand_num;
    char player[34];
    bool player_win;
    bool is_finished;
};

struct addrinfo {
    int64_t betting_times;
    int64_t betting_amount;
    int64_t earnings;
};

struct gamestatus {
    char game_creator[34];
    int64_t deposit;
    int64_t height;
    int64_t game_balance;
    int64_t current_round;
    int64_t finished_round;
    int64_t total_bets;
    int64_t total_player_win;
    bool is_active;
};

#ifdef __cplusplus //而这一部分就是告诉编译器，如果定义了__cplusplus(即如果是cpp文件，
extern "C" { //因为cpp文件默认定义了该宏),则采用C语言方式进行编译
#endif

int startgame(int64_t amount);
int deposit(int64_t amount);
int play(int64_t amount, int64_t number);
int draw();
int stopgame();

#ifdef __cplusplus
}
#endif

//void withdraw(char* creator);
gamestatus get_status();
void set_status(gamestatus status);
//void update_addrinfo(char* addr, int64_t amount, int64_t earnings);
void set_roundinfo(roundinfo info);
roundinfo get_roundinfo(int64_t round);
void gen_roundinfo_key(char* round_key, int64_t round);
//bool is_active();