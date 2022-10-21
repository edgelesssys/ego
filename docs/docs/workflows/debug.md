# Debug your app 🔬
EGo executables can be debugged inside as well as outside of enclaves. Depending on the task, you may prefer one or the other.

## Outside an enclave
An EGo executable can be run as a normal host process without an enclave. Thus, it can also be debugged like any other Go program. This should be your first attempt if the problem isn't related to specific enclave functionality. Use your favorite tools (e.g. the Delve debugger) as usual.

## Inside an enclave
EGo comes with `ego-gdb` that augments `gdb` with enclave support. The `console` interface is the same as `gdb`:
```bash
ego-gdb --args ./helloworld
```

The enclave may raise SIGILL signals during startup. These are expected and will be handled by EGo, so you can just `continue`. Use `handle SIGILL nostop` to do this automatically.

Setting up the `mi` interface for VSCode is a bit trickier. First, install the [C/C++ extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode.cpptools).

Use one of the following templates for your `.vscode/launch.json` file. Just replace `samples/helloworld/helloworld` with the path to your EGo executable.

### Snap
If you installed the EGo snap, use this:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "(gdb) Launch",
            "type": "cppdbg",
            "request": "launch",
            "program": "/snap/ego-dev/current/opt/ego/bin/ego-host",
            "args": [
                "/snap/ego-dev/current/opt/ego/share/ego-enclave:samples/helloworld/helloworld",
                "arg1",
                "arg2"
            ],
            "cwd": "${workspaceFolder}",
            "environment": [
                {
                    "name": "LD_LIBRARY_PATH",
                    "value": "/snap/ego-dev/current/usr/lib/x86_64-linux-gnu"
                }
            ],
            "MIMode": "gdb",
            "miDebuggerPath": "/snap/ego-dev/current/opt/ego/bin/ego-gdb",
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

### DEB package or built from source
If you installed the DEB package or built it yourself, use this:
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
