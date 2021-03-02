#include "../common.h"
#include "evidence.hpp"
#include <string.h>

int set(int64_t n) {
	size_t l = getENVSize(n);
	//暂不支持动态数组 char* value = new char[l]; 给定静态数组长度表示字符串长度上限
	char value[1024] = {0};
	getENV(n, value, l);
	printlog(value, l);
	return 0;
}