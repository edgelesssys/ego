// Copyright (c) Open Enclave SDK contributors.
// Licensed under the MIT License.

// Based on attestation/sgx/evidence.h 

// SGX specific claims
// Required: SGX report body fields that every SQX Quote verification should
// output.
// 1 boolean flag indicated by "sgx_misc_select_t"
#define OE_CLAIM_SGX_PF_GP_EXINFO_ENABLED "sgx_pf_gp_exit_info_enabled"
#define OE_CLAIM_SGX_ISV_EXTENDED_PRODUCT_ID "sgx_isv_extended_product_id"
// 4 boolean flags indicated by "sgx_attributes_t"
#define OE_CLAIM_SGX_IS_MODE64BIT "sgx_is_mode64bit"
#define OE_CLAIM_SGX_HAS_PROVISION_KEY "sgx_has_provision_key"
#define OE_CLAIM_SGX_HAS_EINITTOKEN_KEY "sgx_has_einittoken_key"
#define OE_CLAIM_SGX_USES_KSS "sgx_uses_kss"
#define OE_CLAIM_SGX_CONFIG_ID "sgx_config_id"
#define OE_CLAIM_SGX_CONFIG_SVN "sgx_config_svn"
#define OE_CLAIM_SGX_ISV_FAMILY_ID "sgx_isv_family_id"
#define OE_CLAIM_SGX_CPU_SVN "sgx_cpu_svn"
#define OE_SGX_REQUIRED_CLAIMS_COUNT 10

/*
 * Optional: SQX Quote data
 */
// SQX quote verification collaterals.
#define OE_CLAIM_SGX_TCB_INFO "sgx_tcb_info"
#define OE_CLAIM_SGX_TCB_ISSUER_CHAIN "sgx_tcb_issuer_chain"
#define OE_CLAIM_SGX_PCK_CRL "sgx_pck_crl"
#define OE_CLAIM_SGX_ROOT_CA_CRL "sgx_root_ca_crl"
#define OE_CLAIM_SGX_CRL_ISSUER_CHAIN "sgx_crl_issuer_chain"
#define OE_CLAIM_SGX_QE_ID_INFO "sgx_qe_id_info"
#define OE_CLAIM_SGX_QE_ID_ISSUER_CHAIN "sgx_qe_id_issuer_chain"
#define OE_SGX_OPTIONAL_CLAIMS_SGX_COLLATERALS_COUNT 7
// SGX PCESVN.
#define OE_CLAIM_SGX_PCE_SVN "sgx_pce_svn"
#define OE_SGX_OPTIONAL_CLAIMS_COUNT 8

// Additional SGX specific claim: for the report data embedded in the SGX quote.

#define OE_CLAIM_SGX_REPORT_DATA "sgx_report_data"
