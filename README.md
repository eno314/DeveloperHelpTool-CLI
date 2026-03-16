# Developer Help Tool CLI

Welcome to the Developer Help Tool CLI! This tool is designed to help developers with simple tasks. It is easy to use and built with Go.

## Features

- **Amidakuji (Ghost Leg) Lottery**: Create a random lottery to assign goals to participants quickly and fairly.
- **HTTP Diff**: Compare HTTP responses from two different hosts to easily find differences between environments.

## Installation

**Note: Currently, only macOS (Apple Silicon / arm64) is supported.**

You can install the latest version for macOS using the provided installation script with `wget` or `curl`. Since the script installs the binary to `/usr/local/bin`, you will need to run it with `sudo`. Open your terminal and run:

```bash
wget -qO- https://raw.githubusercontent.com/eno314/DeveloperHelpTool-CLI/refs/heads/main/install.sh | sudo bash
```

Alternatively, if you prefer `curl`:
```bash
curl -sL https://raw.githubusercontent.com/eno314/DeveloperHelpTool-CLI/refs/heads/main/install.sh | sudo bash
```

### Installing a Specific Release Version

By default, the script installs the `latest` release. If you need to install a specific released version, you can pass the version string as an argument to the script. Only official release versions (e.g., `v1.0.0`) are supported.

Using `wget`:
```bash
wget -qO- https://raw.githubusercontent.com/eno314/DeveloperHelpTool-CLI/refs/heads/main/install.sh | sudo bash -s -- v1.0.0
```

Using `curl`:
```bash
curl -sL https://raw.githubusercontent.com/eno314/DeveloperHelpTool-CLI/refs/heads/main/install.sh | sudo bash -s -- v1.0.0
```

*(Note: You can check the [Releases](https://github.com/eno314/developer-help-tool-cli/releases) page for more details on available binaries and versions.)*

## Usage

Currently, the tool has one main command: `amidakuji`.

### HTTP Diff

Use this command to compare the HTTP response bodies of two different hosts. It requests the same paths on both hosts and shows you what is different. It will also warn you if the status codes (like 200 OK or 404 Not Found) are not the same.

**Command:**

```bash
developer-help-tool-cli httpdiff --host1 https://api-staging.example.com --host2 https://api-prod.example.com --paths /users,/posts
```

**Example Output:**

```
Comparing https://api-staging.example.com/users vs https://api-prod.example.com/users
Differences found:
  {
-   "version": "v1.1",
+   "version": "v1.0",
    "users": []
  }
--------------------------------------------------
Comparing https://api-staging.example.com/posts vs https://api-prod.example.com/posts
Warning: Status codes differ. host1: 200, host2: 404
Differences found:
...
--------------------------------------------------
```

### Amidakuji

Use this command to create a Ghost Leg lottery. You need to provide a list of participants and a list of goals. The number of participants must equal the number of goals.

**Command:**

```bash
developer-help-tool-cli amidakuji --participants A,B,C --goals X,Y,Z
```

**Example Output:**

```
 A      B      C
 |      |      |
 |      |------|
 |      |      |
 |------|      |
 |      |      |
 |      |      |
 |      |      |
 X      Y      Z

Results:
  A -> Y
  B -> Z
  C -> X
```

### Help

If you need help or want to see all available commands, run:

```bash
developer-help-tool-cli --help
```
