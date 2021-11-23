// Copyright (c) Open Enclave SDK contributors.
// Licensed under the MIT License.

#include <stdint.h>

#define OE_REPORT_ATTRIBUTES_DEBUG 0x0000000000000001ULL
#define OE_REPORT_ATTRIBUTES_REMOTE 0x0000000000000002ULL
#define OE_EVIDENCE_ATTRIBUTES_SGX_DEBUG OE_REPORT_ATTRIBUTES_DEBUG
#define OE_EVIDENCE_ATTRIBUTES_SGX_REMOTE OE_REPORT_ATTRIBUTES_REMOTE

typedef struct _oe_claim
{
    char* name;
    uint8_t* value;
    size_t value_size;
} oe_claim_t;

#define OE_CLAIM_SECURITY_VERSION "security_version"
#define OE_CLAIM_ATTRIBUTES "attributes"
#define OE_CLAIM_UNIQUE_ID "unique_id"
#define OE_CLAIM_SIGNER_ID "signer_id"
#define OE_CLAIM_PRODUCT_ID "product_id"
#define OE_CLAIM_TCB_STATUS "tcb_status"
#define OE_CLAIM_SGX_REPORT_DATA "sgx_report_data"
