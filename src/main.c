#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <time.h>
#include <pthread.h>
#include <unistd.h>
#include <fcntl.h>

typedef struct
{
    int *count;
    time_t start;
    int *score;
    char *testword;
    pthread_mutex_t *lock;
} myarg_t;

int score = 0;
pthread_mutex_t lock;

unsigned int true_random_seed()
{
    unsigned int seed;
    int fd = open("/dev/urandom", O_RDONLY);
    if (fd < 0)
    {
        perror("open");
        return (unsigned int)time(NULL);
    }
    read(fd, &seed, sizeof(seed));
    close(fd);
    return seed;
}

// void generator(const char *word, int *score, time_t *start)
void *generator(void *arg)
{
    myarg_t *args = (myarg_t *)arg;
    unsigned int seed = true_random_seed();
    srand(seed);
    while ((time(NULL) - args->start) < (60.0 * 5))
    {

        int i = 0;
        int tmpscore = 0;

        while (args->testword[i] != '\0')
        {
            char randomChar = 'a' + (rand() % 26);

            if (randomChar == args->testword[i])
            {
                tmpscore++;
            }
            else
            {
                break;
            }
            i++;
        }
        if (tmpscore > *args->score)
        {

            time_t elapsed = time(NULL) - args->start;
            printf("%02ld:%02ld New High Score! %d - %.*s\n", elapsed / 60, elapsed % 60, tmpscore, tmpscore, args->testword);
            *args->score = tmpscore;
        }
        pthread_mutex_lock(args->lock);
        (*args->count)++;
        pthread_mutex_unlock(args->lock);

    }
    return NULL;
}

int main()
{
    char *testword = strdup("shakespeare"); // allocate on heap
    int *score = malloc(sizeof(int));
    int *count = malloc(sizeof(int));
    long numCores = sysconf(_SC_NPROCESSORS_ONLN);
    if (numCores < 1)
    {
        perror("sysconf");
        return 1;
    }
    printf("Available cores: %ld\n", numCores);
    *count = 0;
    *score = 0;
    time_t start = time(NULL);
    pthread_mutex_init(&lock, NULL);
    myarg_t args = {
        .count = count,
        .start = start,
        .score = score,
        .testword = testword,
        .lock = &lock};
    pthread_t threads[numCores];
    // generator(testword, score, start);
    for (int i = 0; i < numCores; i++)
    {
        int rc = pthread_create(&threads[i], NULL, generator, &args);

        if (rc == 0)
        {
            printf("Thread %d started\n", i + 1);
        }
    }

    for (int i = 0; i < numCores; i++)
    {
        pthread_join(threads[i], NULL);
    }

    printf("Final score is: %d. Word is: %.*s. Total of %d iterations\n", *score, *score, testword, *count);
    free(testword);
    free(score);
    free(count);
    pthread_mutex_destroy(&lock);
    return 0;
}
