#include <dlfcn.h>
#include <elf.h>
#include <cassert>
#include <cstdio>
#include <cstdlib>
#include <cstring>

extern "C" uint64_t _payload_reloc_rva;
extern "C" uint64_t _payload_reloc_size;
extern "C" const void* __oe_get_enclave_base();
extern "C" const void* __oe_get_reloc_end();
extern "C" int ert_get_argc();
extern "C" char** ert_get_argv();

static void _start_main(void my_main(...))
{
    my_main(ert_get_argc(), ert_get_argv());
}

int main()
{
    const auto base = static_cast<const uint8_t*>(__oe_get_reloc_end());
    const auto& ehdr = *reinterpret_cast<const Elf64_Ehdr*>(base);
    const auto phdr = reinterpret_cast<const Elf64_Phdr*>(base + ehdr.e_phoff);

    // find address of dynamic section in program header
    const Elf64_Dyn* dyn = nullptr;
    for (int i = 0; i < ehdr.e_phnum; ++i)
        if (phdr[i].p_type == PT_DYNAMIC)
        {
            assert(phdr[i].p_vaddr);
            dyn = reinterpret_cast<const Elf64_Dyn*>(base + phdr[i].p_vaddr);
            break;
        }
    assert(dyn);

    const char* strtab = nullptr;
    const Elf64_Sym* symtab = nullptr;
    const Elf64_Rela* jmprel = nullptr;
    size_t jmprel_size = 0;

    for (; dyn->d_tag != DT_NULL; ++dyn)
        switch (dyn->d_tag)
        {
            case DT_STRTAB:
                strtab = reinterpret_cast<const char*>(base + dyn->d_un.d_ptr);
                break;
            case DT_SYMTAB:
                symtab =
                    reinterpret_cast<const Elf64_Sym*>(base + dyn->d_un.d_ptr);
                break;
            case DT_JMPREL:
                jmprel =
                    reinterpret_cast<const Elf64_Rela*>(base + dyn->d_un.d_ptr);
                break;
            case DT_PLTRELSZ:
                jmprel_size = dyn->d_un.d_val;
                break;
        }

    const auto rela = reinterpret_cast<const Elf64_Rela*>(
        static_cast<const uint8_t*>(__oe_get_enclave_base()) +
        _payload_reloc_rva);
    for (size_t i = 0; i < _payload_reloc_size / sizeof *rela; ++i)
    {
        const auto& rel = rela[i];
        const auto type = ELF64_R_TYPE(rel.r_info);
        switch (type)
        {
            case R_X86_64_NONE:
                break;
            case R_X86_64_RELATIVE:
                *(const uint8_t**)(base + rel.r_offset) = base + rel.r_addend;
                break;
            case R_X86_64_GLOB_DAT:
                if (strcmp(
                        strtab + symtab[ELF64_R_SYM(rel.r_info)].st_name,
                        "__libc_start_main") == 0)
                    *(void**)(base + rel.r_offset) = (void*)_start_main;
                break;
            default:
                printf("unsupported relocation type %lu\n", type);
                abort();
        }
    }

    bool failed = false;
    for (size_t i = 0; i < jmprel_size / sizeof *jmprel; ++i)
    {
        const auto& rel = jmprel[i];
        const auto type = ELF64_R_TYPE(rel.r_info);
        switch (type)
        {
            case R_X86_64_JUMP_SLOT:
            {
                const auto symbol =
                    strtab + symtab[ELF64_R_SYM(rel.r_info)].st_name;
                const auto addr = dlsym(0, symbol);
                if (addr)
                    *(void**)(base + rel.r_offset) = addr;
                else
                {
                    printf("symbol not found: %s\n", symbol);
                    failed = true;
                }
            }
            break;
            default:
                printf("unsupported relocation type %lu\n", type);
                abort();
        }
    }
    if (failed)
        abort();

    const auto entry = (void (*)())(base + ehdr.e_entry);
    entry();
}

extern const auto ert_ref_dlopen = dlopen;
