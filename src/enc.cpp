#include <elf.h>
#include <openenclave/ert.h>
#include <cassert>
#include <cstdlib>
#include <iostream>
#include <stdexcept>

using namespace std;

static void _start_main(int payload_main(...))
{
    exit(payload_main(ert_get_argc(), ert_get_argv()));
}

int main()
{
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
