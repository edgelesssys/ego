// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

#include <elf.h>
#include <openenclave/advanced/allocator.h>
#include <openenclave/ert.h>
#include <sys/mount.h>
#include <unistd.h>
#include <cassert>
#include <cstdlib>
#include <cstring>
#include <iostream>
#include <stdexcept>
#include <string_view>
#include <thread>
#include "exception_handler.h"
#include "go_runtime_cleanup.h"

static const auto _memfs_name = "edg_memfs";

using namespace std;
using namespace ert;
static int _argc;
static char** _argv;
static int _envc;
static char** _envp;

extern "C" void ert_ego_premain(
    int* argc,
    char*** argv,
    int envc,
    char** envp,
    const char* payload_data);
static char** _merge_argv_env(int argc, char** argv, char** envp);

extern "C" __thread char ert_reserved_tls[11264];
extern "C" const char* oe_sgx_get_td();
extern "C" uint64_t oe_get_num_tcs();

extern "C" void __libc_start_main(int payload_main(...))
{
    exit(payload_main(_argc, _argv));
}

static void _log(string_view s)
{
    cout << "[ego] " << s << '\n';
}

static void _log_verbose(string_view s)
{
    static const bool verbose_enabled = []
    {
        // env is not available via libc (see comment in ert_get_args), so look
        // up manually
        for (int i = 0; i < _envc; ++i)
            if (_envp[i] == "EDG_EGO_VERBOSE=1"s)
                return true;
        return false;
    }();

    if (verbose_enabled)
        _log(s);
}

static void _set_concurrency_limits()
{
    // We must prevent the Go runtime from creating too many system threads.
    // Creating more threads than available TCS will cause OE_OUT_OF_THREADS.
    //
    // Go already knows GOMAXPROCS: "The GOMAXPROCS variable limits the number
    // of operating system threads that can execute user-level Go code
    // simultaneously. There is no limit to the number of threads that can be
    // blocked in system calls on behalf of Go code; those do not count against
    // the GOMAXPROCS limit."
    // (https://pkg.go.dev/runtime#hdr-Environment_Variables)
    //
    // As this isn't a hard limit for system threads, we added EGOMAXTHREADS to
    // achieve this.
    // GOMAXPROCS > EGOMAXTHREADS makes no sense; it will work, but produces
    // scheduling overhead. In practice, GOMAXPROCS should be a bit below
    // EGOMAXTHREADS because there can be threads (e.g., in syscalls) that are
    // not available for a Go proc.

    auto count = oe_get_num_tcs();
    if (count < 6)
        return; // can only happen if enclave was manually signed instead of
                // using `ego sign`

    count -= 2; // safety margin
    setenv("EGOMAXTHREADS", to_string(count).c_str(), false);

    // By default, GOMAXPROCS is the number of cores assigned to the process.
    // Thus, we only need to set it if number of cores come close to or are
    // above EGOMAXTHREADS.
    count -= 2;
    if (thread::hardware_concurrency() > count)
        setenv("GOMAXPROCS", to_string(count).c_str(), false);
}

int emain()
{
    _log_verbose("entered emain");

    // Assert that the variable is located at the end of the TLS block.
    if (oe_sgx_get_td() - ert_reserved_tls != sizeof ert_reserved_tls)
    {
        _log("ert_reserved_tls failure");
        return EXIT_FAILURE;
    }

    // load oe modules
    if (oe_load_module_host_epoll() != OE_OK ||
        oe_load_module_host_file_system() != OE_OK ||
        oe_load_module_host_resolver() != OE_OK ||
        oe_load_module_host_socket_interface() != OE_OK)
    {
        _log("oe_load_module_host failed");
        return EXIT_FAILURE;
    }

    // Initialize memfs
    const Memfs memfs(_memfs_name);

    // Copy potentially existing payload data into string (for null-termination)
    // and pass it to ego's premain
    const auto payload_data_pair = payload::get_data();
    const string payload_data(
        static_cast<const char*>(payload_data_pair.first),
        payload_data_pair.second);

    _log_verbose("invoking premain");
    ert_ego_premain(&_argc, &_argv, _envc, _envp, payload_data.c_str());
    _log_verbose("premain done");
    ert_init_ttls(getenv("MARBLE_TTLS_CONFIG"));

    _set_concurrency_limits();

    // get args and env
    _argv = _merge_argv_env(_argc, _argv, environ);

    // Assume environment variables & mounts were performed in ert_ego_premain

    // If user specified PWD, try to set is as current working directory
    // Otherwise we should be in / (which should be memfs by default)
    const char* const pwd = getenv("PWD");

    if (pwd && chdir(pwd) != 0)
    {
        _log("cannot set cwd to specified pwd");
        return EXIT_FAILURE;
    }
    // cleanup go runtime
    _log_verbose("cleaning up the old goruntime: go_rc_kill_threads");
    go_rc_kill_threads();
    _log_verbose("cleaning up the old goruntime: go_rc_unmap_memory");
    go_rc_unmap_memory();
    _log_verbose("cleaning up the old goruntime: done");

    if (oe_add_vectored_exception_handler(false, ego_exception_handler) !=
        OE_OK)
    {
        _log("oe_add_vectored_exception_handler failed");
        return EXIT_FAILURE;
    }

    // get payload entry point
    const auto base = static_cast<const uint8_t*>(ert::payload::get_base());
    assert(base);
    const auto& ehdr = *reinterpret_cast<const Elf64_Ehdr*>(base);
    assert(ehdr.e_entry);
    const auto entry = (void (*)())(base + ehdr.e_entry);

    _log("starting application ...");
    entry();
    abort(); // unreachable
}

ert_args_t ert_get_args()
{
    //
    // Get env vars from the host.
    //

    ert_args_t args{};
    if (ert_get_args_ocall(&args) != OE_OK || args.envc < 0 || args.argc < 0)
        abort();

    /* Don't make envp available as the environment yet, but rather store it as
     a variable so the Go premain can access the host environment with the
     supposed values (without actually setting them). This is a mitigation to
     avoid the host messing with the Go premain with GODEBUG and similar.
    */
    ert_copy_strings_from_host_to_enclave(
        args.envp, &_envp, static_cast<size_t>(args.envc));

    assert(_envp);
    _envc = args.envc;

    ert_args_t result{};

    //
    // Get args from host.
    //

    char** argv = nullptr;
    ert_copy_strings_from_host_to_enclave(
        args.argv, &argv, static_cast<size_t>(args.argc));

    assert(argv);

    result.argv = argv;
    result.argc = args.argc;

    return result;
}

static char** _merge_argv_env(int argc, char** argv, char** envp)
{
    int envc = 0;
    while (envp[envc])
    {
        envc++;
    }

    // [argv][null][env][null][auxv][null]
    char** p = (char**)oe_allocator_calloc(argc + 1 + envc + 1 + 4, sizeof *p);

    if (!p)
        abort();

    char** result = p;
    memcpy(p, argv, (size_t)argc * sizeof *argv);
    p += argc + 1;
    memcpy(p, envp, (size_t)envc * sizeof *envp);
    p += envc + 1;
    const auto pa = reinterpret_cast<intptr_t*>(p);
    pa[0] = AT_PAGESZ;
    pa[1] = OE_PAGE_SIZE;

    return result;
}
