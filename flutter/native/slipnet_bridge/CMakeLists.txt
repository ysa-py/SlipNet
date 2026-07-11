cmake_minimum_required(VERSION 3.10)
project(slipnet_bridge C)

set(CMAKE_C_STANDARD 11)

# Builds libslipnet_bridge.so for Android (via Gradle externalNativeBuild) and
# can also be built standalone on the host for dart:ffi verification.
add_library(slipnet_bridge SHARED slipnet_bridge.c)

target_include_directories(slipnet_bridge PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})

set_target_properties(slipnet_bridge PROPERTIES
    C_VISIBILITY_PRESET hidden
)
