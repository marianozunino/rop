# Run on Pod (ROP)

## Why Use Run on Pod?
- **Convenience**: Easily execute scripts or binaries directly on Kubernetes pods without manual copying or complex kubectl commands.
- **Flexibility**: Support for both script and binary execution, with automatic detection of file type and custom runners.
- **Efficiency**: Streamlined workflow for developers and operators working with Kubernetes environments.
- **Safety**: Built-in confirmation prompts and context awareness to prevent accidental executions.

## Installation
To install Run on Pod, run:
```bash
go install github.com/marianozunino/rop@latest
```

## Usage
```
 ______     ______     ______
/\  == \   /\  __ \   /\  == \
\ \  __<   \ \ \/\ \  \ \  _-/
 \ \_\ \_\  \ \_____\  \ \_\
  \/_/ /_/   \/_____/   \/_/     v1.3.0


Usage:
  rop [flags]
  rop [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of rop

Flags:
  -c, --context string     Kubernetes context (autocomplete available from kube config)
  -n, --namespace string   Kubernetes namespace (defaults to current namespace if not provided)
  -p, --pod string         The target pod name
      --container string   The container name (optional for single-container pods)
  -f, --file string        The file path to execute
  -a, --args stringArray   File arguments
  -d, --dest-path string   Destination path for the script or binary (default "/tmp")
  -r, --runner string      Custom runner for the script (e.g., 'python', 'node')
  -t, --type string        File type: 'script', 'binary', or 'auto' (default "auto")
      --no-confirm         Skip confirmation prompt
  -v, --verbose            Verbose output
  -h, --help               help for rop

Use "rop [command] --help" for more information about a command.
```

### Examples
1. Run a script on a pod:
   ```
   rop -c my-context -f ./myscript.sh -p my-pod
   ```
2. Execute a binary with arguments:
   ```
   rop -c prod-cluster -f ./myapp -p backend-pod -t binary -a "--verbose" -a "--config=/etc/myapp.conf"
   ```
3. Run in a specific container of a multi-container pod:
   ```
   rop -c dev-cluster -f ./debug.py -p monitoring-pod --container logger
   ```
4. Execute without confirmation prompt:
   ```
   rop -c staging -f ./update-db.sh -p db-pod --no-confirm
   ```
5. Use a custom runner for a script:
   ```
   rop -c dev-cluster -f ./script.js -p nodejs-pod -r node
   ```
6. Specify a custom destination path:
   ```
   rop -c prod-cluster -f ./config.yaml -p config-pod -d /app/config
   ```
7. Test out the completion:
   ```
   rop completion zsh > /tmp/completion; source /tmp/completion
   ```

## Configuration
Run on Pod doesn't require a configuration file. All options are specified via command-line flags.

## How Does Run on Pod Work?
1. **Context Awareness**: Uses the specified Kubernetes context to ensure you're operating in the correct cluster. Contexts can be auto-completed from the kube config.
2. **Namespace Handling**: The namespace can also be auto-completed, and if not provided, it defaults to the current namespace of the context.
3. **File Detection**: Automatically detects whether the file is a script or binary, with an option to override.
4. **Pod Selection**: Targets the specified pod and optionally a specific container within that pod.
5. **File Transfer**: Securely copies the file to the target pod.
6. **Execution**: Runs the file within the pod's context, capturing and displaying output.
7. **Cleanup**: Removes the transferred file from the pod after execution.

## Safety Features
- Confirmation prompt before execution (can be disabled with `--no-confirm` flag)
- Clear display of target context, pod, and container before execution
- Automatic file type detection to prevent incorrect execution methods

## Notes
- **Kubernetes Version**: This tool has been tested with Kubernetes 1.20+. If you encounter issues with other versions, please report them.
- **File Size Limit**: Be aware of potential limitations on file sizes that can be transferred to pods. Very large files may cause issues.
- **Security**: Ensure you have the necessary permissions in your Kubernetes cluster to execute files on pods.
- **Network Dependency**: Requires network access to your Kubernetes cluster. Performance may vary based on network conditions.
- **Custom Runners**: When using the `--runner` flag, ensure that the specified runner is available in the target pod's container.

## Contributing
Contributions to Run on Pod are welcome! Please feel free to submit pull requests, create issues for bugs and feature requests, or contribute to the documentation.

## License
Run on Pod is released under the MIT License. See the LICENSE file for more details.

