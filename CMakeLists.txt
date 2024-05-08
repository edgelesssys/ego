cmake_minimum_required(VERSION 3.11)
project(ego VERSION 1.5.1)

if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE Debug)
endif()
if(CMAKE_BUILD_TYPE STREQUAL Debug)
  # disable optimizations to improve debugging with gdb
  set(PREMAIN_DEBUG_FLAGS -gcflags=all='-N -l')
else()
  set(TRIMPATH -trimpath)
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
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
add_compile_options(-Wall -Wextra -pedantic -Werror)
if(TIDY)
  set(CMAKE_CXX_CLANG_TIDY clang-tidy-11)
endif()

#
# ego-enclave
#

add_library(ego-enclave-lib
  src/enc.cpp
  src/exception_handler.cpp
  src/go_runtime_cleanup.cpp)
target_link_libraries(ego-enclave-lib PRIVATE openenclave::oe_includes)

add_custom_command(
  OUTPUT premain.a
  DEPENDS ego/premain/main.go ego/premain/core/core.go
  COMMAND ertgo build -buildmode=c-archive -o ${CMAKE_BINARY_DIR} ${TRIMPATH} ${PREMAIN_DEBUG_FLAGS}
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/ego/premain)
add_custom_target(premainbuild DEPENDS premain.a)

# Both ego-enclave and the payload use the local-exec TLS model. This means that TLS memory
# will overlap. Quick fix is to reserve space in ego-enclave that we won't touch.
# The Go language does not have TLS and the implementation only stores one pointer in TLS.
# We still reserve a bit more.
# The reserved space must be at the end of the TLS block, so the lib must be linked last.
add_library(reserved_tls src/reserved_tls.c)
add_library(reserved_tls_last INTERFACE)
target_link_libraries(reserved_tls_last INTERFACE reserved_tls)

add_executable(ego-enclave src/gcc_libinit.c src/gcc_mmap.c)
add_dependencies(ego-enclave premainbuild)
target_link_libraries(ego-enclave
  openenclave::oeenclave
  openenclave::ertcalls
  ego-enclave-lib
  openenclave::oehostepoll
  openenclave::oehostfs
  openenclave::oehostresolver
  openenclave::oehostsock
  ${CMAKE_BINARY_DIR}/premain.a
  openenclave::ertlibc
  openenclave::ertttls
  -Wl,--whole-archive
  openenclave::oelibc
  -Wl,--no-whole-archive
  reserved_tls_last)

#
# ego cli
#

add_custom_command(
  OUTPUT ego
  DEPENDS ${CMAKE_SOURCE_DIR}/ego/*/*.go ${CMAKE_SOURCE_DIR}/ego/*/*/*.go
  COMMAND ${CMAKE_SOURCE_DIR}/src/build_ego.sh ${CMAKE_BINARY_DIR} ${PROJECT_VERSION} ${TRIMPATH}
  COMMAND ${CMAKE_BINARY_DIR}/ego completion bash > ${CMAKE_BINARY_DIR}/ego_completion
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/ego/ego)
add_custom_target(egobuild ALL DEPENDS ego)

#
# ego-bundle - the loader executable for bundled ego enclaves
#
add_custom_command(
  OUTPUT ego-bundle
  DEPENDS ${CMAKE_SOURCE_DIR}/ego/*/*.go ${CMAKE_SOURCE_DIR}/ego/*/*/*.go
  COMMAND ${CMAKE_SOURCE_DIR}/src/build_ego.sh ${CMAKE_BINARY_DIR}/ego-bundle ${PROJECT_VERSION} ${TRIMPATH}
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/ego/cmd/bundle)
add_custom_target(egobundle ALL DEPENDS ego-bundle)

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
install(FILES ${CMAKE_BINARY_DIR}/ego-bundle DESTINATION ${CMAKE_INSTALL_DATADIR})
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
install(FILES ${OpenEnclave_DIR}/../../../include/openenclave/attestation/verifier.h DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/openenclave/attestation)
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
  install(FILES ${CMAKE_BINARY_DIR}/ego_completion DESTINATION /${CMAKE_INSTALL_SYSCONFDIR}/bash_completion.d)
endif()

set(CPACK_PACKAGE_CONTACT "contact@edgeless.systems​")
set(CPACK_PACKAGING_INSTALL_PREFIX ${CMAKE_INSTALL_PREFIX})
set(CPACK_DEBIAN_PACKAGE_DEPENDS
  "libsgx-enclave-common (>=2.3.100.46354-1), libsgx-dcap-ql (>=1.0.100.46460-1.0)")
set(CPACK_DEBIAN_FILE_NAME DEB-DEFAULT)
set(CPACK_DEBIAN_ARCHIVE_TYPE gnutar)
include(CPack)

#
# tests
#

add_custom_command(
  OUTPUT test_private.pem
  COMMAND openssl genrsa -out test_private.pem -3 3072)

# build concurrency-test with ertgo and sign it with oesign
add_custom_command(
  OUTPUT concurrency-test
  DEPENDS ego-enclave test_private.pem ego/cmd/concurrency-test/enclave.conf ego/cmd/concurrency-test/main.go
  COMMAND ${CMAKE_COMMAND} -E env GOROOT=${CMAKE_SOURCE_DIR}/_ertgo ${CMAKE_SOURCE_DIR}/_ertgo/bin/go build -o ${CMAKE_BINARY_DIR}
  COMMAND oesign sign
    -e ${CMAKE_BINARY_DIR}/ego-enclave
    -c ${CMAKE_SOURCE_DIR}/ego/cmd/concurrency-test/enclave.conf
    -k ${CMAKE_BINARY_DIR}/test_private.pem
    --payload ${CMAKE_BINARY_DIR}/concurrency-test
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/ego/cmd/concurrency-test)

# build the test marble with ertgo and sign it with oesign
add_custom_command(
  OUTPUT test-marble
  DEPENDS ego-enclave test_private.pem ego/cmd/test-marble/enclave.conf ego/cmd/test-marble/main.go ego/test/t.go
  COMMAND ${CMAKE_COMMAND} -E env GOROOT=${CMAKE_SOURCE_DIR}/_ertgo ${CMAKE_SOURCE_DIR}/_ertgo/bin/go build -o ${CMAKE_BINARY_DIR}
  COMMAND oesign sign
    -e ${CMAKE_BINARY_DIR}/ego-enclave
    -c ${CMAKE_SOURCE_DIR}/ego/cmd/test-marble/enclave.conf
    -k ${CMAKE_BINARY_DIR}/test_private.pem
    --payload ${CMAKE_BINARY_DIR}/test-marble
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/ego/cmd/test-marble)

add_custom_command(
  OUTPUT marble-test
  DEPENDS ego/cmd/marble-test/main.go ego/test/t.go
  COMMAND CGO_ENABLED=0 go build -o ${CMAKE_BINARY_DIR}
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/ego/cmd/marble-test)

add_custom_target(testexes ALL DEPENDS concurrency-test test-marble marble-test)

enable_testing()
add_test(NAME api-unit-tests COMMAND go test -race --count=3 ./... WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME ego-unit-tests COMMAND go test -race --count=3 ./... WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/ego)
add_test(integration ${CMAKE_SOURCE_DIR}/src/integration_test.sh)
add_test(concurrency erthost ego-enclave:concurrency-test)
add_test(marble marble-test)
