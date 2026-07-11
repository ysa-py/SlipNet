/*
 * slipnet_bridge — thin, stable C ABI between the Flutter (dart:ffi) UI and the
 * existing SlipNet native cores (Rust slipstream / Go libs / C tun2socks).
 *
 * This shim is the ONLY new native code in the Flutter migration. It never
 * re-implements the cores; in later phases each entry point forwards to the
 * corresponding exported core symbol. All strings returned are owned by the
 * shim and must NOT be freed by the caller.
 */
#ifndef SLIPNET_BRIDGE_H
#define SLIPNET_BRIDGE_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#if defined(_WIN32)
#define SLIPNET_EXPORT __declspec(dllexport)
#else
#define SLIPNET_EXPORT __attribute__((visibility("default")))
#endif

/* Connection state codes (kept in sync with Dart ConnectionState). */
enum {
    SLIPNET_STATE_DISCONNECTED = 0,
    SLIPNET_STATE_CONNECTING   = 1,
    SLIPNET_STATE_CONNECTED    = 2,
    SLIPNET_STATE_ERROR        = 3
};

/* Status codes returned by connect/disconnect. 0 == OK, negative == error. */
enum {
    SLIPNET_OK              = 0,
    SLIPNET_ERR_BAD_CONFIG  = -1,
    SLIPNET_ERR_BUSY        = -2,
    SLIPNET_ERR_NOT_READY   = -3
};

/* ABI compatibility guard. Bump when the contract changes. */
SLIPNET_EXPORT int32_t slipnet_bridge_abi_version(void);

/* Human-readable version string of the bridge + (later) core versions. */
SLIPNET_EXPORT const char *slipnet_bridge_version(void);

/*
 * Begin a connection using a JSON config blob (profile). Non-blocking:
 * transitions state to CONNECTING, then CONNECTED. Returns a SLIPNET_* status.
 */
SLIPNET_EXPORT int32_t slipnet_bridge_connect(const char *config_json);

/* Tear down the active connection. */
SLIPNET_EXPORT int32_t slipnet_bridge_disconnect(void);

/* Current SLIPNET_STATE_* value. */
SLIPNET_EXPORT int32_t slipnet_bridge_state(void);

/*
 * Live stats as a JSON string:
 * {"bytesUp":N,"bytesDown":N,"rttMs":N,"packetLossPct":F,"protocol":"..."}
 */
SLIPNET_EXPORT const char *slipnet_bridge_stats_json(void);

/* Last error message ("" when none). */
SLIPNET_EXPORT const char *slipnet_bridge_last_error(void);

#ifdef __cplusplus
}
#endif

#endif /* SLIPNET_BRIDGE_H */
