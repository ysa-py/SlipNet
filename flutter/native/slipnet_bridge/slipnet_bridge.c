/*
 * slipnet_bridge implementation.
 *
 * Phase 1: self-contained, deterministic behaviour so the full
 * Flutter -> dart:ffi -> .so pipeline is verifiable end-to-end today.
 *
 * Phase 2+: the marked TODOs forward to the real cores' exported symbols
 * (e.g. the Rust slipstream start/stop entry points). The cores themselves
 * are never modified.
 */
#include "slipnet_bridge.h"

#include <stdio.h>
#include <string.h>

#define SLIPNET_BRIDGE_ABI 1
#define SLIPNET_BRIDGE_VERSION_STR "slipnet-bridge/0.1.0 (phase1)"

/* Simple process-global state. Guarded lightly; the Dart side serialises calls. */
static int32_t g_state = SLIPNET_STATE_DISCONNECTED;
static char g_last_error[256] = {0};
static char g_stats_buf[256] = {0};
static uint64_t g_bytes_up = 0;
static uint64_t g_bytes_down = 0;

static void set_error(const char *msg) {
    if (msg == NULL) {
        g_last_error[0] = '\0';
        return;
    }
    snprintf(g_last_error, sizeof(g_last_error), "%s", msg);
}

int32_t slipnet_bridge_abi_version(void) {
    return SLIPNET_BRIDGE_ABI;
}

const char *slipnet_bridge_version(void) {
    return SLIPNET_BRIDGE_VERSION_STR;
}

int32_t slipnet_bridge_connect(const char *config_json) {
    if (config_json == NULL || config_json[0] == '\0') {
        set_error("empty config");
        g_state = SLIPNET_STATE_ERROR;
        return SLIPNET_ERR_BAD_CONFIG;
    }
    if (g_state == SLIPNET_STATE_CONNECTING ||
        g_state == SLIPNET_STATE_CONNECTED) {
        set_error("already connecting or connected");
        return SLIPNET_ERR_BUSY;
    }
    set_error(NULL);
    /* TODO(phase2): forward config_json to the Rust slipstream core start entry
     * point and drive real state via the JNI-equivalent callbacks. */
    g_bytes_up = 0;
    g_bytes_down = 0;
    g_state = SLIPNET_STATE_CONNECTED;
    return SLIPNET_OK;
}

int32_t slipnet_bridge_disconnect(void) {
    /* TODO(phase2): forward to the core's stop entry point. */
    g_state = SLIPNET_STATE_DISCONNECTED;
    set_error(NULL);
    return SLIPNET_OK;
}

int32_t slipnet_bridge_state(void) {
    return g_state;
}

const char *slipnet_bridge_stats_json(void) {
    /* TODO(phase2): pull real counters from the core (nativeGetTrafficStats
     * equivalent). Phase 1 emits a stable, well-formed shape the UI can bind. */
    if (g_state == SLIPNET_STATE_CONNECTED) {
        g_bytes_up += 1024;
        g_bytes_down += 4096;
    }
    snprintf(g_stats_buf, sizeof(g_stats_buf),
             "{\"bytesUp\":%llu,\"bytesDown\":%llu,\"rttMs\":%d,"
             "\"packetLossPct\":%.1f,\"protocol\":\"%s\"}",
             (unsigned long long)g_bytes_up,
             (unsigned long long)g_bytes_down,
             g_state == SLIPNET_STATE_CONNECTED ? 42 : 0,
             0.0,
             g_state == SLIPNET_STATE_CONNECTED ? "slipstream" : "none");
    return g_stats_buf;
}

const char *slipnet_bridge_last_error(void) {
    return g_last_error;
}
