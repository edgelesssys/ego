// Copyright (c) Edgeless Systems GmbH.
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

#include "exception_handler.h"

extern "C" oe_result_t ert_cpuid_ocall(
    unsigned int,
    unsigned int,
    unsigned int*,
    unsigned int*,
    unsigned int*,
    unsigned int*);

// This handler handles cpuid exceptions that aren't already handled by ERT/OE.
//
// Based on
// https://github.com/deislabs/mystikos/blob/v0.7.0/tools/myst/enc/enc.c
uint64_t ego_exception_handler(oe_exception_record_t* exception_context)
{
    auto& ctx = *exception_context->context;

    // give up if this isn't a cpuid exception
    const uint16_t cpuid_opcode = 0xA20F;
    if (exception_context->code != OE_EXCEPTION_ILLEGAL_INSTRUCTION ||
        *reinterpret_cast<uint16_t*>(ctx.rip) != cpuid_opcode)
        return OE_EXCEPTION_CONTINUE_SEARCH;

    const bool is_xsave_subleaf_zero = ctx.rax == 0xd && ctx.rcx == 0;

    unsigned int eax = 0;
    unsigned int ebx = 0;
    unsigned int ecx = 0;
    unsigned int edx = 0;
    if (ert_cpuid_ocall(ctx.rax, ctx.rcx, &eax, &ebx, &ecx, &edx) != OE_OK)
        return OE_EXCEPTION_CONTINUE_SEARCH;

    if (is_xsave_subleaf_zero)
    {
        /* replace XSAVE/XRSTOR save area size with fixed large value of 4096,
        to protect against spoofing attacks from untrusted host.
        If host returns smaller xsave area than required, this can cause a
        buffer overflow at context switch time.
        We believe value of 4096 should be sufficient for forseeable future. */
        if (ebx < 4096)
            ebx = 4096;
        if (ecx < 4096)
            ecx = 4096;
    }

    ctx.rax = eax;
    ctx.rbx = ebx;
    ctx.rcx = ecx;
    ctx.rdx = edx;
    ctx.rip += sizeof cpuid_opcode;

    return OE_EXCEPTION_CONTINUE_EXECUTION;
}
