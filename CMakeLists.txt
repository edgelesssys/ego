cmake_minimum_required(VERSION 3.11)
project(ego VERSION 0.1.0)

if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE Debug)
endif()
if(CMAKE_INSTALL_PREFIX_INITIALIZED_TO_DEFAULT)
  set(CMAKE_INSTALL_PREFIX /opt/ego CACHE PATH "" FORCE)
endif()

include(GNUInstallDirs)
find_package(OpenEnclave CONFIG REQUIRED)

execute_process(
  COMMAND git submodule update --init _ertgo
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_EXTENSIONS OFF)
add_compile_options(-Wall -Wextra -pedantic -Werror)
if(TIDY)
  set(CMAKE_CXX_CLANG_TIDY clang-tidy-10)
endif()

add_library(ego-enclave-lib
  src/enc.cpp
  src/go_runtime_cleanup.cpp)
target_link_libraries(ego-enclave-lib PRIVATE openenclave::oe_includes)


# Both ego-enclave and the payload use the local-exec TLS model. This means that TLS memory
# will overlap. Quick fix is to reserve space in ego-enclave that we won't touch.
# The Go language does not have TLS and the implementation only stores one pointer in TLS.
# We still reserve a bit more.
# The reserved space must be at the end of the TLS block, so the lib must be linked last.
add_library(reserved_tls src/reserved_tls.c)
add_library(reserved_tls_last INTERFACE)
target_link_libraries(reserved_tls_last INTERFACE reserved_tls)

add_executable(ego-enclave src/gcc_libinit.c src/gcc_mmap.c)
target_link_libraries(ego-enclave
  openenclave::oeenclave
  openenclave::ertcalls
  ego-enclave-lib
  openenclave::oehostepoll
  openenclave::oehostfs
  openenclave::oehostresolver
  openenclave::oehostsock
  openenclave::ertmeshpremain
  openenclave::ertlibc
  -Wl,--whole-archive
  openenclave::oelibc
  -Wl,--no-whole-archive
  reserved_tls_last)

add_custom_command(
  OUTPUT ego
  DEPENDS cmd/ego/main.go ${CMAKE_SOURCE_DIR}/internal/cli/*.go
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
install(FILES ${OpenEnclave_DIR}/../host/liboehostverify.a DESTINATION ${CMAKE_INSTALL_LIBDIR})
install(FILES ${OpenEnclave_DIR}/../../../include/openenclave/host_verify.h DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/openenclave)
install(
  FILES
  ${OpenEnclave_DIR}/../../../include/openenclave/bits/defs.h
  ${OpenEnclave_DIR}/../../../include/openenclave/bits/evidence.h
  ${OpenEnclave_DIR}/../../../include/openenclave/bits/report.h
  ${OpenEnclave_DIR}/../../../include/openenclave/bits/result.h
  ${OpenEnclave_DIR}/../../../include/openenclave/bits/types.h
  DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/openenclave/bits)
install(DIRECTORY ${OpenEnclave_DIR}/../debugger DESTINATION ${CMAKE_INSTALL_LIBDIR}/openenclave)
install(DIRECTORY _ertgo/ DESTINATION go USE_SOURCE_PERMISSIONS)

set(CPACK_PACKAGE_CONTACT "contact@edgeless.systems​")
set(CPACK_PACKAGING_INSTALL_PREFIX ${CMAKE_INSTALL_PREFIX})
set(CPACK_DEBIAN_PACKAGE_DEPENDS
  "libsgx-enclave-common (>=2.3.100.46354-1), libsgx-dcap-ql (>=1.0.100.46460-1.0), libsgx-dcap-ql-dev (>=1.0.100.46460-1.0)")
set(CPACK_DEBIAN_FILE_NAME DEB-DEFAULT)
set(CPACK_DEBIAN_ARCHIVE_TYPE gnutar)
include(CPack)
