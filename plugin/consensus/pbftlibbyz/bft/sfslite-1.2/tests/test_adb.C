/* 
 * test_adb.C
 *
 *  Test the adb (asynchronous database) by inserting randomly
 *  generated key/value pairs and attempting to retrieve them from the
 *  database.
 *
 */

#include "btreeSync.h"
#include "stdlib.h"
#include "db_cxx.h"

#define Min(X, Y) ( ( (X) < (Y) ) ? (X) : (Y) )

struct argsRec {
  long key_low;
  long key_high;
  long data_low;
  long data_high;
  long numPairs;
  long node_size;
  long node_mult;
  char create;
  char *filename;
};

char **gKeys;
char **gData;
long *gKeySizes;
long *gDataSizes;
char gVerbose, gStats, gCreate;
struct timeval gStartTime, gFinishTime;
long outstandingTransactions;

void insert_done(tid_t tid, int err, record *rec);
void lookup_done(long i, tid_t tid, int err, record *rec);

char
randomChar() {
  long r = random();
  return (char)('a' + (char)('z' - 'a')*((float)(((float)(r))/RAND_MAX)));
}

void
randomString(char *rstr, int len) {
  
  for (int i = 0; i < len; i++) rstr[i] = randomChar();
  rstr[len] = 0;
}

long
boundedRand(long low, long high) {
  long r = random();
  return low + (long)((high - low)*((float)(((float)(r))/RAND_MAX)));

}

void
startTimer() {
  gettimeofday(&gStartTime, NULL);
}

void
stopTimer() {
  gettimeofday(&gFinishTime, NULL);
}

long
elapsedmsecs() {
  return (gFinishTime.tv_sec - gStartTime.tv_sec)*1000 + (gFinishTime.tv_usec - gStartTime.tv_usec)/1000;
}

void
generateInputPairs(long key_low, 
		   long key_high, 
		   long data_low, 
		   long data_high,
		   long n) {

  if (gVerbose) printf("Generating %ld input pairs\n", n);

  gKeys = (char **)malloc(sizeof(char *)*n);
  gData = (char **)malloc(sizeof(char *)*n);
  gKeySizes = (long *)malloc(sizeof(long)*n);
  gDataSizes = (long *)malloc(sizeof(long)*n);

  for (long i = 0; i < n; i++) {

    //randomly pick sizes for these keys
    gKeySizes[i] = boundedRand(key_low,key_high);
    gDataSizes[i] = boundedRand(data_low, data_high);
    
    gKeys[i] = (char *)malloc(gKeySizes[i] + 1);
    gData[i] = (char *)malloc(gDataSizes[i] + 1);

    randomString(gKeys[i], gKeySizes[i]);
    randomString(gData[i], gDataSizes[i]);
  }
  
}

void
insertPairs(btreeSync *db, long n) {

  for (long i = 0; i < n; i++) {
    if (gVerbose) printf("Inserting pair %ld: <%s, %s>\n", i, gKeys[i], gData[i]);
    bError_t err = db->insert(gKeys[i], gKeySizes[i], gData[i], gDataSizes[i] + 1);
    if (err) {
      printf("ERROR: insertion failed (%s)\n", gKeys[i]);
      exit(0);
    }
  }
 if (gVerbose) printf("Insertion Complete\n");
}

void
verifyDB(btreeSync *db, long n) {
  record *res;
  void *result;
  bSize_t len;
  
  if (gVerbose) printf("Verifying database\n");

  for (long i = 0; i < n; i++) {
    bError_t err = db->lookup(gKeys[i], gKeySizes[i], &res);
    if (err)  { 
      printf("ERROR: lookup of (%s) returned %s\n", gKeys[i], bstrerror(err));
      exit(0);
    }

    result = res->getValue(&len);
    if ( (gDataSizes[i] != len - 1) || (memcmp(result, gData[i], Min(len, gDataSizes[i])) != 0)) {
      printf("ERROR: (%ld) %s doesn't match  %s\n", i, (char *)result, gData[i]);
      exit(0);
    }
    
    if (gVerbose) printf("%ld: %s matches %s\n", i, (char *)result, gData[i]); 
    delete res;
  }

}

void
printUsage(char *appName) {
    printf("usage: %s [-vS] <number of pairs> <low_key_size> <high_key_size> <low_data_size> <high_data_size> <filename>\n", appName);
    exit(0);
}

void
parseTestFile(char *paramFile, argsRec *a) {
  
  FILE *f = fopen(paramFile, "r");
  if (f == NULL) {
    printf("couldn't open %s for reading, skipping\n", paramFile);
    return;
  }

  /*
   *
   * File Format
   *
   * # a comment
   * <number of pairs> \n
   * <low value of key size range> \n
   * <high value of key size range> \n
   * <low value of data size range> \n
   * <high value of data size range> \n
   * <node size> \n
   * <data mult> \n
   * <create/reuse> \n
   * <database filename> \n
   */
  char lineBuf[128];
  char tmp;
  int valuesParsed = 0;
  do {
    int pos = 0;
    do {
      tmp = (char)getc(f);
      lineBuf[pos++] = tmp;
    } while ((tmp != '\n') && (tmp != -1));
    lineBuf[pos - 1] = 0;
    if (lineBuf[0] != '#') {
      switch (valuesParsed) {
      case 0:
	a->numPairs = atoi(lineBuf);
	break;
      case 1:
	a->key_low =  atoi(lineBuf);
	break;
      case 2:
	a->key_high = atoi(lineBuf);
	break;
      case 3:
	a->data_low = atoi(lineBuf);
	break;
      case 4:
	a->data_high = atoi(lineBuf);
	break;
      case 5:
	a->node_size = atoi(lineBuf);
	break;
      case 6:
	a->node_mult = atoi(lineBuf);
	break;
      case 7: 
	a->create = (strcmp(lineBuf, "create") == 0) ? 1 : 0;
	break;
      case 8:
	a->filename = strdup(lineBuf);
	break;
      }
      valuesParsed++;
    }
  } while (valuesParsed < 9);
  fclose(f);
}

void
testBTree(char *paramFile, argsRec a) {

  printf(" (BTree) Running Test %s.\n"
         "   %ld pairs\n"
	 "   Key sizes: %ld-%ld\n"
	 "   Data sizes: %ld-%ld\n"
	 "   Nodes are %ld bytes (%ld data multiplier)\n"
	 "   Database %s ", paramFile, a.numPairs, a.key_low, a.key_high, a.data_low, a.data_high, a.node_size, a.node_mult, a.filename);
  if (a.create) printf("will be created\n\n");
  else printf("will be reused\n\n");
  
  btreeSync *db;
  if (a.create) {
    bError_t err = createTree(a.filename, a.node_size, a.node_mult);
    if (err) {
      printf("Error: creating database\n");
      exit(0);
    }
  }
  
  db = new btreeSync();
  db->open(a.filename, 5000000);

  srandom(1);
  
  //run the test
  startTimer();
  if (a.create) insertPairs(db, a.numPairs);
  stopTimer();

  printf("Elapsed time for insertion: %ld msec (%f keys/sec)\n", 
	 elapsedmsecs(), 1000.0*((float)a.numPairs)/(elapsedmsecs()));

  startTimer();
  verifyDB(db, a.numPairs);
  stopTimer();
  printf("Elapsed time for verification: %ld msec (%f keys/sec)\n", 
	 elapsedmsecs(), 1000.0*((float)a.numPairs)/(elapsedmsecs()));
  
  db->finalize();
  delete db;


  return;
}

void
testSleepyCat(char *paramFile, argsRec a) {

  printf(" (SleepyCat) Running Test %s.\n"
         "   %ld pairs\n"
	 "   Key sizes: %ld-%ld\n"
	 "   Data sizes: %ld-%ld\n"
	 "   Database %s ", paramFile, a.numPairs, a.key_low, a.key_high, a.data_low, a.data_high, a.filename);
  if (a.create) printf("will be created\n\n");
  else printf("will be reused\n\n");

  DB *db;
  db_create(&db, NULL, 0L);
  char filename[128];
  sprintf(filename, "%s.sc", a.filename);
  db->open(db, filename, NULL, DB_BTREE, DB_CREATE | DB_TRUNCATE, 0666);

  srandom(1);

  startTimer();
  //insertion
  DBT key, data;
  bzero(&key, sizeof(DBT));
  bzero(&data, sizeof(DBT));
  for (long i = 0; i < a.numPairs; i++) {
    key.data = gKeys[i];
    key.size = gKeySizes[i]; 
    data.data = gData[i];
    data.size = gDataSizes[i];
    int err  = db->put(db, NULL, &key, &data, 0L);
    if (err) printf("Error: %s\n", strerror(err));
  }
  stopTimer();

  printf("Elapsed time for insertion: %ld msec (%f keys/sec)\n", 
	 elapsedmsecs(), 1000.0*((float)a.numPairs)/(elapsedmsecs()));


  //verification
  startTimer();
  bzero(&key, sizeof(DBT));
  bzero(&data, sizeof(DBT));
  char tmpData[8192];
  for (long i = 0; i < a.numPairs; i++) {
    key.data = gKeys[i];
    key.size = gKeySizes[i]; 
    data.data = tmpData;
    data.size = gDataSizes[i];
    int err  = db->get(db, NULL, &key, &data, 0L);
    if (memcmp(data.data, gData[i], data.size) != 0) {
      ((char *)data.data)[gDataSizes[i]] = 0;
      printf("ERROR: data.size = %d, data.data = %s\n", data.size, (char *)data.data);
      printf("gData = %s\n", gData[i]);
    }
    if (err) printf("Error: %s\n", strerror(err));
  }
  stopTimer();

  printf("Elapsed time for verification: %ld msec (%f keys/sec)\n", 
	 elapsedmsecs(), 1000.0*((float)a.numPairs)/(elapsedmsecs()));

}

void
testAsyncBTree(char *paramFile, argsRec a) {

  printf(" (async btree) Running Test %s.\n"
         "   %ld pairs\n"
	 "   Key sizes: %ld-%ld\n"
	 "   Data sizes: %ld-%ld\n"
	 "   Node size: %ld (%ld mult)\n"
	 "   Database %s ", paramFile, a.numPairs, a.key_low, a.key_high, a.data_low, a.data_high, a.node_size, a.node_mult,  a.filename);
  if (a.create) printf("will be created\n\n");
  else printf("will be reused\n\n");

  if (a.create) {
    bError_t err = createTree(a.filename, a.node_size, a.node_mult);
    if (err) {
      printf("Error: creating database\n");
      exit(0);
    }
  }
  
  btreeDispatch *db = new btreeDispatch(a.filename, 5000000);
  db->setInsertPolicy(kOverwrite);

  startTimer();

  //insertion
  long i = 0;
  outstandingTransactions = 0;
  while (i < a.numPairs)  {
    if (outstandingTransactions < 4) {
      bError_t err = db->insert(gKeys[i], gKeySizes[i], gData[i], gDataSizes[i] + 1, wrap(&insert_done));
      if (err) printf("Error: %s\n", strerror(err));
      outstandingTransactions++;
      i++;
    }
    else
      acheck();
  }
  while (outstandingTransactions) acheck();
  stopTimer();

  printf("Elapsed time for insertion: %ld msec (%f keys/sec)\n", 
	 elapsedmsecs(), 1000.0*((float)a.numPairs)/(elapsedmsecs()));

  //verification
  startTimer();
  outstandingTransactions = 0;
  i = 0;
  while (i < a.numPairs)  {
    if (outstandingTransactions < 4) {
      bError_t err = db->lookup(gKeys[i], gKeySizes[i], wrap(&lookup_done, i));
      if (err) printf("Error: %s\n", strerror(err));
      outstandingTransactions++;
      i++;
    }
    else
      acheck();
  }
  while (outstandingTransactions) acheck();
  stopTimer();

  printf("Elapsed time for verification: %ld msec (%f keys/sec)\n", 
	 elapsedmsecs(), 1000.0*((float)a.numPairs)/(elapsedmsecs()));

}

void
insert_done(tid_t tid, int err, record *rec) {
  if (err) {
    printf("error on insert (async): %s\n", bstrerror(err));
    exit(0);
  }
  outstandingTransactions--;
}

void
lookup_done(long i, tid_t tid, int err, record *rec) {
  if (err) {
    printf("error on insert (async): %s\n", bstrerror(err));
    exit(0);
  }
  outstandingTransactions--;

  void * result;
  bSize_t len;
  result = rec->getValue(&len);
  if ( (gDataSizes[i] != len - 1) || (memcmp(result, gData[i], Min(len, gDataSizes[i])) != 0)) {
    printf("ERROR: (%ld) %s doesn't match  %s\n", i, (char *)result, gData[i]);
    exit(0);
  }
}

int
main(int argc, char **argv) {
  
#define MINARGS 2
  gVerbose = gStats = gCreate = 0;
  
  if (argc < MINARGS) printUsage(argv[0]);
  
  for (int argNum = 1; argNum < argc; argNum++) {
    if (strchr(argv[argNum], '-')) {
      if (strchr(argv[argNum], 'v')) gVerbose = true;
      if (strchr(argv[argNum], 'S')) gStats = true;
      if ((!gVerbose) && (!gStats)) printUsage(argv[0]); 
    } else {
      argsRec a;
      parseTestFile(argv[argNum], &a);
      srandom(1);
      generateInputPairs(a.key_low, a.key_high, a.data_low, a.data_high, a.numPairs);
      printf("--------------------------------\n");
      testSleepyCat(argv[argNum],a);
      testBTree(argv[argNum],a);
      testAsyncBTree(argv[argNum],a);
      printf("--------------------------------\n");

      free(gKeySizes);
      free(gDataSizes);
      for (int i=0; i < a.numPairs; i++) {
	free(gKeys[i]); free(gData[i]); 
      }
      free(gKeys); free(gData);
    }
  }
  
}




