cmake_minimum_required(VERSION 3.11)
project(runner)
find_package(OpenEnclave REQUIRED PATHS ${EDGELESSRT_INSTALL}/lib/openenclave/cmake)

add_executable(ego-enclave enc.cpp)
target_link_libraries(ego-enclave
  openenclave::oeenclave
  openenclave::ertdeventry
  openenclave::ertlibc
  -Wl,--whole-archive
  openenclave::oelibc
  -Wl,--no-whole-archive)
