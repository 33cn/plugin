#ifdef __cplusplus
extern "C" {
#endif  
  ssize_t readvfd (int fd, const struct iovec *iov, int iovcnt, int *rfdp);
  ssize_t writevfd (int fd, const struct iovec *iov, int iovcnt, int wfd);
  ssize_t readfd (int fd, void *buf, size_t len, int *rfdp);
  ssize_t writefd (int fd, const void *buf, size_t len, int wfd);
#ifdef __cplusplus
}
#endif
