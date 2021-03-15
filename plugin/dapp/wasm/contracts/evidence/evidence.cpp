#include "../common.h"
#include "evidence.hpp"
#include <string.h>
#define SUCC 0
//----------------------------------------------------------------------------------------------
//-m AddStateTx -v "TestKey001","TestValue001"
int AddStateTx()
{
    if(totalENV() != 2) return -1;
    char ChKey[128] = {0};
    char ChValue[128] = {0};
    //环境变量索引是从0开始的,0,1,2,3...
    getENV(0, ChKey, getENVSize(0));
    getENV(1, ChValue, getENVSize(1));
    if (string_size(ChKey) == 0) return -1;
    //1.检查存在这样的记录
    if (getStateDBSize(ChKey, string_size(ChKey)) != 0) return -1;
    //2.插入到链上数据库
    setStateDB(ChKey, string_size(ChKey), ChValue, string_size(ChValue));
    return SUCC;
}
//----------------------------------------------------------------------------------------------
int DelStateTx()
{
    if(totalENV() !=1) return -1;
    //1.从环境变量中获取变量值
    char ChKey[128] = {0};
    getENV(0, ChKey, getENVSize(0));
    if (string_size(ChKey) == 0) return -1;
    //2.获取对应的记录
    if(getStateDBSize(ChKey, string_size(ChKey)) == 0) {
        const char info[128] = "DelStateTx::getStateDBSize Not Exists\0";
        printlog(info, string_size(info));
        return -1;
    }
    //size_t getStateDB(const char* key, size_t k_len, char* value, size_t v_len);
    char ChValue[128] = {0};
    getStateDB(ChKey, string_size(ChKey), ChValue, sizeof(ChValue));
    if(string_size(ChValue) == 0) {
        //记录存在，但是之前删除过了
        const char info[128] = "DelStateTx::getStateDB Has Been Deleted\0";
        printlog(info, string_size(info));
        return -1;
    }
    char ChNull[128] = {0};
    setStateDB(ChKey, string_size(ChKey), ChNull, string_size(ChNull));
    const char info[128] = "DelStateTx::Exec setStateDB Del OK\0";
    printlog(info, string_size(info));
    return SUCC;
}
//----------------------------------------------------------------------------------------------
int ModStateTx()
{
    if(totalENV() != 2) return -1;
    char ChKey[128] = {0};
    char ChValue[128] = {0};
    getENV(0, ChKey, getENVSize(0));
    getENV(1, ChValue, getENVSize(1));
    if ((string_size(ChKey) == 0) || (string_size(ChValue) == 0)) return -1;
    //查询旧值，非空为存在旧值，为空则表示这个记录已经被删除过
    if(getStateDBSize(ChKey, string_size(ChKey)) == 0) {
        const char info[128] = "ModStateTx::getStateDBSize failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    //在判断这条记录value是否为空，如果是则为已经被删除的记录，如此修改操作不允许
    char value[128] = {0};
    getStateDB(ChKey, string_size(ChKey), value, sizeof(value));
    if(string_size(value) == 0) {
        const char info[128] = "ModStateTx::getStateDB Has Been Deleted\0";
        printlog(info, string_size(info));
        return -1;
    }
    setStateDB(ChKey, string_size(ChKey), ChValue, string_size(ChValue));
    const char info[128] = "ModStateTx::Exec setStateDB Update OK\0";
    printlog(info, string_size(info));
    return SUCC;
}
//----------------------------------------------------------------------------------------------
int AddLocalTx()
{
    if(2 != totalENV()) return -1;
    char ChKey[128] = {0};
    char ChValue[128] = {0};
    getENV(0, ChKey, getENVSize(0));
    getENV(1, ChValue, getENVSize(1));
    if (string_size(ChKey) == 0) return -1;
    //1.检查存在这样的记录
    if (getLocalDBSize(ChKey, string_size(ChKey)) != 0) return -1;
    //2.存储本地数据
    setLocalDB(ChKey, string_size(ChKey), ChValue, string_size(ChValue));
    return SUCC;
}
//----------------------------------------------------------------------------------------------
int DelLocalTx()
{
    if(totalENV() != 1) return -1;
    char ChKey[128] = {0};
    char szValue[128] = {0};
    getENV(0, ChKey, getENVSize(0));
    if (string_size(ChKey) == 0) return -1;
    if(getLocalDBSize(ChKey, string_size(ChKey)) == 0) {
        const char info[128] = "DelLocalTx::getLocalDBSize Length is 0\0";
        printlog(info, string_size(info));
        return -1;
    }
    char ChValue[128] = {0};
    getLocalDB(ChKey, string_size(ChKey), ChValue, sizeof(ChValue));
    if(string_size(ChValue) == 0) {
        //记录存在，但是之前删除过了
        const char info[128] = "DelLocalTx::getLocalDB Has Been Deleted\0";
        printlog(info, string_size(info));
        return -1;
    }
    char ChNull[128] = {0};
    setLocalDB(ChKey, string_size(ChKey), ChNull, string_size(ChNull));
    const char info[128] = "DelLocalTx::Exec Delete OK\0";
    printlog(info, string_size(info));
    return SUCC;
}
//----------------------------------------------------------------------------------------------
int ModLocalTx()
{
    if(totalENV() != 2) return -1;
    char ChKey[128] = {0};
    char ChValue[128] = {0};
    getENV(0, ChKey, getENVSize(0));
    getENV(1, ChValue, getENVSize(1));
    if (string_size(ChKey) == 0) return -1;
    //1.检查数据记录是否存在
    if(getLocalDBSize(ChKey, string_size(ChKey)) == 0) {
        const char info[128] = "ModLocalTx::getLocalDBSize Not Exist\0";
        printlog(info, string_size(info));
        return -1;
    }
    char value[128] = {0};
    getLocalDB(ChKey, string_size(ChKey), value, sizeof(value));
    if(string_size(value) == 0) {
        const char info[128] = "ModLocalTx::getLocalDB Has Been Deleted\0";
        printlog(info, string_size(info));
        return -1;
    }
    setLocalDB(ChKey, string_size(ChKey), ChValue, string_size(ChValue));
    const char info[128] = "ModLocalTx::Exec setLocalDB Update OK\0";
    printlog(info, string_size(info));
    return SUCC;
}
//----------------------------------------------------------------------------------------------
