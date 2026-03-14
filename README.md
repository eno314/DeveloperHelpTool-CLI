# Developer Help Tool CLI

Welcome to the Developer Help Tool CLI! This tool is designed to help developers with simple tasks. It is easy to use and built with Go.

## Features

- **Amidakuji (Ghost Leg) Lottery**: Create a random lottery to assign goals to participants quickly and fairly.

## Installation

You can download the latest version for macOS (Apple Silicon / arm64) using `wget`. Open your terminal and run:

```bash
wget https://github.com/eno314/developer-help-tool-cli/releases/latest/download/developer-help-tool-cli-darwin-arm64
chmod +x developer-help-tool-cli-darwin-arm64
sudo mv developer-help-tool-cli-darwin-arm64 /usr/local/bin/developer-help-tool-cli
```

*(Note: The binary name might change depending on the release. Please check the [Releases](https://github.com/eno314/developer-help-tool-cli/releases) page for the exact URL.)*

## Usage

Currently, the tool has one main command: `amidakuji`.

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
