
SVCXPRT *svcbyz_create(char* in, int in_size, char *out, int out_size);
void svcbyz_recycle(SVCXPRT *xpt, char* in, int in_size, char *out, int out_size);
int svcbyz_reply_bytes(SVCXPRT *xpt);
enum auth_stat svcbyz_authenticate(struct svc_req *rqst, struct rpc_msg *msg);
