# Run on Pod (rop)

## Why Use Run on Pod?

- **Convenience**: Easily execute scripts or binaries directly on Kubernetes pods without manual copying or complex kubectl commands.
- **Flexibility**: Support for both script and binary execution, with automatic detection of file type.
- **Efficiency**: Streamlined workflow for developers and operators working with Kubernetes environments.
- **Safety**: Built-in confirmation prompts and context awareness to prevent accidental executions.

## Installation

To install Run on Pod, run:

```bash
go install github.com/marianozunino/rop@latest
```

## Usage

```
Usage:
  rop [flags]

Flags:
  -a, --args string       Arguments to pass to the script or binary
      --container string  The container name (optional for single-container pods)
  -c, --context string    Kubernetes context
  -f, --file string       The file path to execute
  -h, --help              help for rop
  -n, --no-confirm        Skip confirmation prompt
  -p, --pod string        The target pod name
  -t, --type string       File type: 'script', 'binary', or 'auto' (default "auto")
```

### Examples

1. Run a script on a pod:
   ```
   rop -c my-context -f ./myscript.sh -p my-pod
   ```

2. Execute a binary with arguments:
   ```
   rop -c prod-cluster -f ./myapp -p backend-pod -t binary -a "--verbose --config=/etc/myapp.conf"
   ```

3. Run in a specific container of a multi-container pod:
   ```
   rop -c dev-cluster -f ./debug.py -p monitoring-pod --container logger
   ```

4. Execute without confirmation prompt:
   ```
   rop -c staging -f ./update-db.sh -p db-pod -n
   ```

## Configuration

Run on Pod doesn't require a configuration file. All options are specified via command-line flags.

## How Does Run on Pod Work?

1. **Context Awareness**: Uses the specified Kubernetes context to ensure you're operating in the correct cluster.
2. **File Detection**: Automatically detects whether the file is a script or binary, with an option to override.
3. **Pod Selection**: Targets the specified pod and optionally a specific container within that pod.
4. **File Transfer**: Securely copies the file to the target pod.
5. **Execution**: Runs the file within the pod's context, capturing and displaying output.
6. **Cleanup**: Removes the transferred file from the pod after execution.

## Safety Features

- Confirmation prompt before execution (can be disabled with `-n` flag)
- Clear display of target context, pod, and container before execution
- Automatic file type detection to prevent incorrect execution methods

## Notes

- **Kubernetes Version**: This tool has been tested with Kubernetes 1.20+. If you encounter issues with other versions, please report them.
- **File Size Limit**: Be aware of potential limitations on file sizes that can be transferred to pods. Very large files may cause issues.
- **Security**: Ensure you have the necessary permissions in your Kubernetes cluster to execute files on pods.
- **Network Dependency**: Requires network access to your Kubernetes cluster. Performance may vary based on network conditions.

## Contributing

Contributions to Run on Pod are welcome! Please feel free to submit pull requests, create issues for bugs and feature requests, or contribute to the documentation.

## License

Run on Pod is released under the MIT License. See the LICENSE file for more details.
