cmake_minimum_required(VERSION 3.11)
project(ego)
if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE Debug)
endif()
if(CMAKE_INSTALL_PREFIX_INITIALIZED_TO_DEFAULT)
  set(CMAKE_INSTALL_PREFIX /opt/ego CACHE PATH "" FORCE)
endif()

include(ExternalProject)
include(GNUInstallDirs)

execute_process(
  COMMAND git submodule update --init edgelessrt ertgo
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
execute_process(
  COMMAND git submodule update --init tools/oeedger8r-cpp 3rdparty/mbedtls/mbedtls
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/edgelessrt)

ExternalProject_Add(edgelessrt
  SOURCE_DIR edgelessrt-src
  BINARY_DIR edgelessrt
  DOWNLOAD_COMMAND cp -r ${CMAKE_CURRENT_SOURCE_DIR}/edgelessrt/. ${CMAKE_CURRENT_BINARY_DIR}/edgelessrt-src
  CMAKE_ARGS -DCMAKE_BUILD_TYPE=${CMAKE_BUILD_TYPE} -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/edgelessrt-install -DBUILD_TESTS=OFF
  USES_TERMINAL_BUILD ON)

ExternalProject_Add(runner
  DEPENDS edgelessrt
  SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/runner
  BINARY_DIR runner
  DOWNLOAD_COMMAND ""
  CMAKE_ARGS -DCMAKE_BUILD_TYPE=${CMAKE_BUILD_TYPE} -DEDGELESSRT_INSTALL=${CMAKE_CURRENT_BINARY_DIR}/edgelessrt-install
  BUILD_ALWAYS ON
  INSTALL_COMMAND "")

add_custom_command(
  OUTPUT ego
  DEPENDS cmd/ego/main.go
  COMMAND go build ${CMAKE_SOURCE_DIR}/cmd/ego)
add_custom_target(egobuild ALL DEPENDS ego)

install(
  PROGRAMS
  src/ego-env
  src/ego-go
  src/ego-run
  ${CMAKE_BINARY_DIR}/ego
  DESTINATION ${CMAKE_INSTALL_BINDIR})
install(
  PROGRAMS ${CMAKE_BINARY_DIR}/edgelessrt-install/bin/erthost
  RENAME ego-host
  DESTINATION ${CMAKE_INSTALL_BINDIR})
install(
  PROGRAMS ${CMAKE_BINARY_DIR}/edgelessrt-install/bin/oesign
  RENAME ego-sign
  DESTINATION ${CMAKE_INSTALL_BINDIR})
install(
  FILES
  src/enclave.conf
  ${CMAKE_BINARY_DIR}/runner/ego-enclave
  DESTINATION ${CMAKE_INSTALL_DATADIR})
install(DIRECTORY ertgo/ DESTINATION go USE_SOURCE_PERMISSIONS)
