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

#
# ego-enclave
#

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

#
# ego cli
#

add_custom_command(
  OUTPUT ego
  DEPENDS cmd/ego/main.go ${CMAKE_SOURCE_DIR}/internal/cli/*.go
  COMMAND go build -o ${CMAKE_CURRENT_BINARY_DIR}/ego
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/cmd/ego)
add_custom_target(egobuild ALL DEPENDS ego)

#
# install
#

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

if(CMAKE_INSTALL_PREFIX STREQUAL /opt/ego)
  install(
    FILES
    src/symlinks/ego
    src/symlinks/ego-gdb
    src/symlinks/ego-go
    DESTINATION /usr/local/bin)
endif()

set(CPACK_PACKAGE_CONTACT "contact@edgeless.systems​")
set(CPACK_PACKAGING_INSTALL_PREFIX ${CMAKE_INSTALL_PREFIX})
set(CPACK_DEBIAN_PACKAGE_DEPENDS
  "libsgx-enclave-common (>=2.3.100.46354-1), libsgx-dcap-ql (>=1.0.100.46460-1.0), libsgx-dcap-ql-dev (>=1.0.100.46460-1.0)")
set(CPACK_DEBIAN_FILE_NAME DEB-DEFAULT)
set(CPACK_DEBIAN_ARCHIVE_TYPE gnutar)
include(CPack)

#
# tests
#

add_custom_command(
  OUTPUT test_private.pem
  COMMAND openssl genrsa -out test_private.pem -3 3072)

# build the test marble with ertgo and sign it with oesign
add_custom_command(
  OUTPUT test-marble
  DEPENDS ego-enclave test_private.pem cmd/test-marble/enclave.conf cmd/test-marble/main.go internal/test/t.go
  COMMAND ${CMAKE_COMMAND} -E env GOROOT=${CMAKE_SOURCE_DIR}/_ertgo ${CMAKE_SOURCE_DIR}/_ertgo/bin/go build -o ${CMAKE_BINARY_DIR}
  COMMAND oesign sign
    -e ${CMAKE_BINARY_DIR}/ego-enclave
    -c ${CMAKE_SOURCE_DIR}/cmd/test-marble/enclave.conf
    -k ${CMAKE_BINARY_DIR}/test_private.pem
    --payload ${CMAKE_BINARY_DIR}/test-marble
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/cmd/test-marble)

add_custom_command(
  OUTPUT marble-test
  DEPENDS cmd/marble-test/main.go internal/test/t.go
  COMMAND go build -o ${CMAKE_BINARY_DIR}
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/cmd/marble-test)

add_custom_target(marbletest ALL DEPENDS test-marble marble-test)

enable_testing()
add_test(NAME go-unit-tests COMMAND go test ./... WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(integration ${CMAKE_SOURCE_DIR}/src/integration_test.sh)
add_test(marble marble-test)
