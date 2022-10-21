# Programming model
Enclaves are execution environments isolated from the rest of the system. In the original SGX programming model, the application code is partitioned into trusted and untrusted code. The untrusted code runs in a conventional process. Within this process, one or more enclaves are created that execute the trusted code. The enclave is entered with an *ECALL*. The enclave can transfer execution to untrusted code by performing an *OCALL*.

EGo has a different programming model: The entire application runs inside the enclave. Transitions between trusted and untrusted code are hidden inside the EGo runtime and are transparent to the developer. Internally, an ECALL is performed when the enclave creates a new thread. The enclave uses OCALLs to implement some classes of syscalls, e.g., file and network I/O. (Some syscalls can be fully emulated within the enclave.) Manual ECALLs/OCALLs by application code are neither required nor possible.

Advantages of this approach are:
* Developing a confidential app is almost like developing a conventional app. Developers don't need to learn a new programming model.
* No need to partition an app, which can be effortful and error-prone
* Porting an existing app is simple
