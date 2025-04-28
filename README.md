# Terraform Moved Directive Remover

A Go tool that recursively scans Terraform files and removes all `moved` blocks.

## Overview

This tool helps clean up Terraform configurations by removing all `moved` directives, which are typically used during refactoring but may not be needed after changes are applied.

## Features

- Recursively scans directories for `.tf` files
- Identifies and removes all `moved` blocks
- Modifies files in-place
- Reports detailed statistics about the changes made
- Uses Terraform's HCL parser for accurate syntax handling

## Requirements

- Go 1.24 or later

## Installation

### From Source

```bash
git clone https://github.com/mkusaka/terraform-moved-remover.git
cd terraform-moved-remover
go build -o terraform-moved-remover cmd/terraform-moved-remover/main.go
```

### Using Go Install

```bash
go install github.com/mkusaka/terraform-moved-remover/cmd/terraform-moved-remover@latest
```

This will install the binary in your `$GOPATH/bin` directory.

## Usage

```bash
./terraform-moved-remover [options] <directory>
```

### Options

- `-help`: Display help information
- `-version`: Display version information
- `-dry-run`: Run without modifying files
- `-verbose`: Enable verbose output

### Example

```bash
./terraform-moved-remover -verbose ./terraform
```

This will:
1. Scan all `.tf` files in the `./terraform` directory and its subdirectories
2. Remove all `moved` blocks from these files
3. Display statistics about the changes made

## Example Output

```
Scanning directory: ./terraform
Found 15 Terraform files

Statistics:
Files processed: 15
Files modified: 7
Moved blocks removed: 12
Processing time: 235.412ms
```

## How It Works

The tool uses HashiCorp's HCL library to parse Terraform files and manipulate the Abstract Syntax Tree (AST). This ensures proper handling of Terraform's syntax and maintains formatting of the files.

## License

MIT
