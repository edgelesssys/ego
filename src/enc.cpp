#include <elf.h>
#include <openenclave/ert.h>
#include <cassert>
#include <cstdlib>
#include <iostream>
#include <stdexcept>

using namespace std;

extern "C" __thread char ert_ego_reserved_tls[1024];
extern "C" const char* oe_sgx_get_td();

static void _start_main(int payload_main(...))
{
    exit(payload_main(ert_get_argc(), ert_get_argv()));
}

int main()
{
    // Accessing this variable makes sure that the reserved_tls lib will be
    // linked. See comment about the lib in CMakeLists for more info.
    *ert_ego_reserved_tls = 0;
    // Assert that the variable is located at the end of the TLS block.
    assert(
        oe_sgx_get_td() - ert_ego_reserved_tls == sizeof ert_ego_reserved_tls);

    // relocate
    try
    {
        ert::payload::apply_relocations(_start_main);
    }
    catch (const exception& e)
    {
        cout << "apply_relocations failed: " << e.what() << '\n';
        return EXIT_FAILURE;
    }

    // get payload entry point
    const auto base = static_cast<const uint8_t*>(ert::payload::get_base());
    assert(base);
    const auto& ehdr = *reinterpret_cast<const Elf64_Ehdr*>(base);
    assert(ehdr.e_entry);
    const auto entry = (void (*)())(base + ehdr.e_entry);

    entry();
}
