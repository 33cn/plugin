#ifndef _JLI_H_
#define _JLI_H_

// jvm函数，创建虚拟机
int JLI_Create_JVM(const char *jdkPath, const char *jarPath);
// jvm函数，可以用来分别进行交易的执行和查询，交易的执行是顺序的,而查询可以并发进行
int JLI_Exec_Contract(int argc, char **argv, char **exceptionInfo, int jobType, char *jvmGo);

// jvm函数，当chain33停止的时候，对该虚拟机进行销毁操作
int JLI_Detroy_JVM();

/* utility functions */
extern int GetPtrSize();
extern void SetPtr(char **ptr, char *value, int index);
extern void FreeArgv(int argc, char **argv);
extern char ** GetNil2dPtr();
extern void * GetVoidPtr(char *voidPtr);

#endif /* _JAVA_H_ */
