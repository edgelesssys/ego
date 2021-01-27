cmake_minimum_required(VERSION 3.11)
project(ego)

if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE Debug)
endif()
if(CMAKE_INSTALL_PREFIX_INITIALIZED_TO_DEFAULT)
  set(CMAKE_INSTALL_PREFIX /opt/ego CACHE PATH "" FORCE)
endif()

include(GNUInstallDirs)
find_package(OpenEnclave CONFIG REQUIRED)

execute_process(
  COMMAND git submodule update --init ertgo
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})

add_library(ego-enclave-lib src/enc.cpp)
target_link_libraries(ego-enclave-lib PRIVATE openenclave::oe_includes)

add_executable(ego-enclave src/empty.c)
target_link_libraries(ego-enclave
  openenclave::oeenclave
  openenclave::ertdeventry
  ego-enclave-lib
  openenclave::ertlibc
  -Wl,--whole-archive
  openenclave::oelibc
  -Wl,--no-whole-archive)

add_custom_command(
  OUTPUT ego
  DEPENDS cmd/ego/main.go
  COMMAND go build -o ${CMAKE_CURRENT_BINARY_DIR}/ego
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/cmd/ego)
add_custom_target(egobuild ALL DEPENDS ego)

install(TARGETS ego-enclave DESTINATION ${CMAKE_INSTALL_DATADIR})
install(
  PROGRAMS
  src/ego-gdb
  src/ego-go
  ${CMAKE_BINARY_DIR}/ego
  DESTINATION ${CMAKE_INSTALL_BINDIR})
install(
  PROGRAMS ${OpenEnclave_DIR}/../../../bin/erthost
  RENAME ego-host
  DESTINATION ${CMAKE_INSTALL_BINDIR})
install(
  PROGRAMS ${OpenEnclave_DIR}/../../../bin/oesign
  RENAME ego-oesign
  DESTINATION ${CMAKE_INSTALL_BINDIR})
install(DIRECTORY ${OpenEnclave_DIR}/../debugger DESTINATION ${CMAKE_INSTALL_LIBDIR}/openenclave)
install(DIRECTORY ertgo/ DESTINATION go USE_SOURCE_PERMISSIONS)
