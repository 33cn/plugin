#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>

main(int argc, char **argv) {
  if (argc < 3) {
    fprintf(stderr, "usage: %s fs_name fs_size (in 8K pages) \n", argv[0]);
    exit(-1);
  }

  int fd = open(argv[1], O_WRONLY | O_EXCL | O_CREAT, 0777);
  int size = atoi(argv[2]);
  char page[8192];
  int i;

  bzero(page, 8192);
  for (i = 0; i < size; i++) {
    write(fd, page, 8192);
  }
}
