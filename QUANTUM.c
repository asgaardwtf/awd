#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include <time.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>
#include <arpa/inet.h>

#define HTTP_PORT 80
#define MAX_REQUEST_SIZE 1024

// Struct to pass arguments to threads
typedef struct {
    char *target;
    int duration;
    int concurrency;
} thread_args;

volatile int running = 1;

void *make_requests(void *arg) {
    thread_args *args = (thread_args *) arg;
    char *target = args->target;

    // Craft a basic GET request
    char request[MAX_REQUEST_SIZE];
    snprintf(request, sizeof(request), "GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", target);

    struct sockaddr_in server_addr;
    struct hostent *server;

    server = gethostbyname(target);
    if (server == NULL) {
        fprintf(stderr, "Error: No such host\n");
        return NULL;
    }

    // Fill in server address structure
    bzero((char *) &server_addr, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    bcopy((char *) server->h_addr, (char *)&server_addr.sin_addr.s_addr, server->h_length);
    server_addr.sin_port = htons(HTTP_PORT);

    while (running) {
        // Create a socket
        int sockfd = socket(AF_INET, SOCK_STREAM, 0);
        if (sockfd < 0) {
            perror("Error opening socket");
            continue;
        }

        // Connect to the server
        if (connect(sockfd, (struct sockaddr *) &server_addr, sizeof(server_addr)) < 0) {
            perror("Error connecting to the server");
            close(sockfd);
            continue;
        }

        // Send the HTTP GET request
        if (send(sockfd, request, strlen(request), 0) < 0) {
            perror("Error sending request");
            close(sockfd);
            continue;
        }

        // Receive the response
        char response[4096];
        bzero(response, 4096);
        recv(sockfd, response, 4095, 0);

        // Close the socket after receiving response
        close(sockfd);
    }

    return NULL;
}

void start_stress_test(char *target, int duration, int concurrency) {
    pthread_t threads[concurrency];
    thread_args args = {target, duration, concurrency};

    // Start the timer
    time_t start_time = time(NULL);

    // Create threads to make concurrent requests
    for (int i = 0; i < concurrency; i++) {
        pthread_create(&threads[i], NULL, make_requests, (void *)&args);
    }

    // Wait for the duration of the test
    while (time(NULL) - start_time < duration) {
        sleep(1);
    }

    // Stop all threads
    running = 0;

    // Wait for all threads to finish
    for (int i = 0; i < concurrency; i++) {
        pthread_join(threads[i], NULL);
    }

    printf("Stress test finished!\n");
}

int main(int argc, char *argv[]) {
    if (argc != 4) {
        fprintf(stderr, "Usage: %s <target> <duration> <concurrency>\n", argv[0]);
        exit(1);
    }

    char *target = argv[1];
    int duration = atoi(argv[2]);
    int concurrency = atoi(argv[3]);

    start_stress_test(target, duration, concurrency);

    return 0;
}
