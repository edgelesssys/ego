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
#include "go_runtime_cleanup.h"

static const auto _memfs_name = "edg_memfs";
static const auto _verbose_env_key = "EDG_EGO_VERBOSE";

using namespace std;
using namespace ert;
static int _argc;
static char** _argv;

extern "C" void ert_ego_premain(
    int* argc,
    char*** argv,
    const char* payload_data);
static char** _merge_argv_env(int argc, char** argv, char** envp);

extern "C" __thread char ert_ego_reserved_tls[1024];
extern "C" const char* oe_sgx_get_td();

static void _start_main(int payload_main(...))
{
    exit(payload_main(_argc, _argv));
}

static void _log(string_view s)
{
    cout << "[ego] " << s << '\n';
}

static void _log_verbose(string_view s)
{
    static const bool verbose_enabled = [] {
        const char* const env_verbose = getenv(_verbose_env_key);
        return env_verbose && *env_verbose == '1';
    }();

    if (verbose_enabled)
        _log(s);
}

int emain()
{
    _log_verbose("entered emain");

    // Accessing this variable makes sure that the reserved_tls lib will be
    // linked. See comment about the lib in CMakeLists for more info.
    *ert_ego_reserved_tls = 0;
    // Assert that the variable is located at the end of the TLS block.
    assert(
        oe_sgx_get_td() - ert_ego_reserved_tls == sizeof ert_ego_reserved_tls);

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
    ert_ego_premain(&_argc, &_argv, payload_data.c_str());
    _log_verbose("premain done");

    // get args and env
    _argv = _merge_argv_env(_argc, _argv, environ);

    // Assume environment variables & mounts were performed in ert_ego_premain

    // If user specified PWD, try to set is as current working directory
    // Otherwise fall back to / (which should be memfs by default)
    const char* const pwd = getenv("PWD");

    if (pwd && chdir(pwd) != 0)
    {
        _log("cannot set cwd to specified pwd");
        return EXIT_FAILURE;
    }
    if (!pwd && chdir("/") != 0)
    {
        _log("cannot set cwd to root");
        return EXIT_FAILURE;
    }

    // cleanup go runtime
    _log_verbose("cleaning up the old goruntime: go_rc_kill_threads");
    go_rc_kill_threads();
    _log_verbose("cleaning up the old goruntime: go_rc_unmap_memory");
    go_rc_unmap_memory();
    _log_verbose("cleaning up the old goruntime: done");
    // relocate
    try
    {
        ert::payload::apply_relocations(_start_main);
    }
    catch (const exception& e)
    {
        _log("apply_relocations failed: "s + e.what());
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

    char** env = nullptr;
    ert_copy_strings_from_host_to_enclave(
        args.envp, &env, static_cast<size_t>(args.envc));

    assert(env);

    ert_args_t result{};
    result.envc = args.envc;
    result.envp = env;

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
    char** p = (char**)oe_allocator_calloc(argc + 1 + envc + 1 + 2, sizeof *p);

    if (!p)
        abort();

    char** result = p;
    memcpy(p, argv, (size_t)argc * sizeof *argv);
    p += argc + 1;
    memcpy(p, envp, (size_t)envc * sizeof *envp);

    return result;
}

// unsetenv+cgo is broken in Go < 1.15. This is a temporary fix until we update
// ertgo.
extern "C" void x_cgo_unsetenv(char** arg)
{
    unsetenv(arg[0]);
}
extern "C" void x_cgo_setenv(char** arg)
{
    setenv(arg[0], arg[1], 1);
}
