// Reserve TLS space for payload. See comment about reserved_tls lib in
// CMakeLists.
__thread char ert_ego_reserved_tls[1024];
