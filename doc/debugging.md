# Debugging
EGo executables can be debugged inside as well as outside of enclaves. Depending on the task, one or the other may be preferable.

## Debugging outside an enclave
An EGo executable can be run as a normal host process without an enclave. Thus, it can also be debugged like any other Go program. This should be your first attempt if the problem is not related to specific enclave functionality. Use your favorite tools (e.g., the Delve debugger) as usual.

## Debugging inside an enclave
EGo comes with `ego-gdb` that augments `gdb` with enclave support. The `console` interface is the same as `gdb`:
```sh
ego-gdb --args ./helloworld
```

Setting up the `mi` interface is a bit trickier, so we provide a VSCode template:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "(gdb) Launch",
            "type": "cppdbg",
            "request": "launch",
            "program": "/opt/ego/bin/ego-host",
            "args": [
                "/opt/ego/share/ego-enclave:samples/helloworld/helloworld",
                "arg1",
                "arg2"
            ],
            "cwd": "${workspaceFolder}",
            "environment": [],
            "MIMode": "gdb",
            "miDebuggerPath": "/opt/ego/bin/ego-gdb",
            "setupCommands": [
                {
                    "description": "Enable pretty-printing for gdb",
                    "text": "-enable-pretty-printing",
                    "ignoreFailures": true
                },
                {
                    "text": "handle SIGILL nostop"
                }
            ]
        }
    ]
}
```
Replace `samples/helloworld/helloworld` with the path to your EGo executable.
