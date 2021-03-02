#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>
#include <stdint.h>

void setStateDB(const char* key, size_t k_len, const char* value, size_t v_len);
size_t getStateDBSize(const char* key, size_t k_len);
size_t getStateDB(const char* key, size_t k_len, char* value, size_t v_len);

void setLocalDB(const char* key, size_t k_len, const char* value, size_t v_len);
size_t getLocalDBSize(const char* key, size_t k_len);
size_t getLocalDB(const char* key, size_t k_len, char* value, size_t v_len);

int64_t getBalance(const char* addr, size_t addr_len, const char* exec_addr, size_t exec_len);
int64_t getFronzen(const char* addr, size_t addr_len, const char* exec_addr, size_t exec_len);

void execAddress(const char* name, size_t name_len, const char* addr, size_t addr_len);

int transfer(const char* from_addr, size_t from_len, const char* to_addr, size_t to_len, int64_t amount);
int transferToExec(const char* from_addr, size_t from_len, const char* exec_addr, size_t exec_len, int64_t amount);
int transferWithdraw(const char* from_addr, size_t from_len, const char* exec_addr, size_t exec_len, int64_t amount);

int execFrozen(const char* addr, size_t addr_len, int64_t amount);
int execActive(const char* addr, size_t addr_len, int64_t amount);
int execTransfer(const char* from_addr, size_t from_len, const char* to_addr, size_t to_len, int64_t amount);
int execTransferFrozen(const char* from_addr, size_t from_len, const char* to_addr, size_t to_len, int64_t amount);

void getFrom(const char* from_addr, size_t from_len);
int64_t getHeight();
int64_t getRandom();
void sha256(const char* data, size_t data_len, char* sum, size_t sum_len);
void printlog(const char* log, size_t len);
void printint(int64_t n);
size_t getENVSize(int64_t n);
size_t getENV(int64_t n, char* value, size_t v_len);

#ifdef __cplusplus
}
#endif


inline size_t string_size(const char* s) {
    size_t l = 0;
    for (const char* tmp=s;*tmp!='\0';tmp++) {
        l++;
    }
    return l;
}





