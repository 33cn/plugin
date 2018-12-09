#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>

main(int argc, char **argv) {
  //
  // Create a file of zeros for BFS.
  //
  if (argc < 3) {
    fprintf(stderr, "usage: %s fs_name fs_size (in KB ) \n", argv[0]);
    exit(-1);
  }

  int fd = open(argv[1], O_WRONLY | O_CREAT, 0777);
  int size = atoi(argv[2]);
  char page[1024];
  int i;

  bzero(page, 1024);
  for (i = 0; i < size; i++) {
    write(fd, page, 1024);
  }
}
