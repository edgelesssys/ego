// Copyright (c) Open Enclave SDK contributors.
// Licensed under the MIT License.

#include <stdint.h>

#define OE_UNIQUE_ID_SIZE 32
#define OE_SIGNER_ID_SIZE 32
#define OE_PRODUCT_ID_SIZE 16
#define OE_REPORT_ATTRIBUTES_DEBUG 1
#define OE_REPORT_ATTRIBUTES_REMOTE 2

#define SGX_FLAGS_MODE64BIT 0x0000000000000004ULL
#define SGX_FLAGS_PROVISION_KEY 0x0000000000000010ULL
#define SGX_FLAGS_EINITTOKEN_KEY 0x0000000000000020ULL

/* Set the bits which have no security implications to 0 for sealed data
 migration */
/* Bits which have no security implications in attributes.flags:
 *    Reserved bit[55:6]  - 0xFFFFFFFFFFFFC0ULL
 *    SGX_FLAGS_MODE64BIT
 *    SGX_FLAGS_PROVISION_KEY
 *    SGX_FLAGS_EINITTOKEN_KEY */
#define SGX_FLAGS_NON_SECURITY_BITS                                        \
    (0xFFFFFFFFFFFFC0ULL | SGX_FLAGS_MODE64BIT | SGX_FLAGS_PROVISION_KEY | \
     SGX_FLAGS_EINITTOKEN_KEY)

/* bit[27:0]: have no security implications */
#define SGX_MISC_NON_SECURITY_BITS 0x0FFFFFFFU

/* OE seal key default flag masks*/
#define OE_SEALKEY_DEFAULT_FLAGSMASK (~SGX_FLAGS_NON_SECURITY_BITS)
#define OE_SEALKEY_DEFAULT_MISCMASK (~SGX_MISC_NON_SECURITY_BITS)
#define OE_SEALKEY_DEFAULT_XFRMMASK (0X0ULL)

typedef int oe_enclave_type_t;

typedef struct _oe_identity
{
    /** Version of the oe_identity_t structure */
    uint32_t id_version;

    /** Security version of the enclave. For SGX enclaves, this is the
     *  ISVN value */
    uint32_t security_version;

    /** Values of the attributes flags for the enclave -
     *  OE_REPORT_ATTRIBUTES_DEBUG: The report is for a debug enclave.
     *  OE_REPORT_ATTRIBUTES_REMOTE: The report can be used for remote
     *  attestation */
    uint64_t attributes;

    /** The unique ID for the enclave.
     * For SGX enclaves, this is the MRENCLAVE value */
    uint8_t unique_id[OE_UNIQUE_ID_SIZE];

    /** The signer ID for the enclave.
     * For SGX enclaves, this is the MRSIGNER value */
    uint8_t signer_id[OE_SIGNER_ID_SIZE];

    /** The Product ID for the enclave.
     * For SGX enclaves, this is the ISVPRODID value. */
    uint8_t product_id[OE_PRODUCT_ID_SIZE];
} oe_identity_t;

typedef struct _oe_report
{
    /** Size of the oe_report_t structure. */
    size_t size;

    /** The enclave type. Currently always OE_ENCLAVE_TYPE_SGX. */
    oe_enclave_type_t type;

    /** Size of report_data */
    size_t report_data_size;

    /** Size of enclave_report */
    size_t enclave_report_size;

    /** Pointer to report data field within the report byte-stream supplied to
     * oe_parse_report */
    uint8_t* report_data;

    /** Pointer to report body field within the report byte-stream supplied to
     * oe_parse_report. */
    uint8_t* enclave_report;

    /** Contains the IDs and attributes that are part of oe_identity_t */
    oe_identity_t identity;

    /** Contains the result reported by quote verification logic. The size is
     * determined based on OE_ENUM_MAX. */
    uint32_t verification_result;
} oe_report_t;
