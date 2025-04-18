# Terraform Moved Remover Examples

This directory contains example Terraform configurations that demonstrate the functionality of the Terraform Moved Remover tool.

## Directory Structure

```
examples/
├── main.tf                    # Basic example with moved blocks
├── modules/                   # Nested directory with modules
│   ├── main.tf                # Module configuration with moved block
│   └── networking/            # Nested module
│       └── vpc.tf             # VPC configuration with moved blocks
└── edge_cases/                # Examples of edge cases
    ├── empty.tf               # Empty file
    ├── only_moved.tf          # File with only moved blocks
    └── commented.tf           # File with commented moved blocks
```

## Usage

1. Run the tool on the examples directory:

```bash
./terraform-moved-remover ./examples
```

2. Observe the output statistics showing files processed and moved blocks removed.

3. Check the modified files to see that the moved blocks have been removed.

## What to Expect

- The tool will process all `.tf` files recursively in the examples directory.
- It will remove all `moved` blocks from these files.
- It will report statistics about the files processed and blocks removed.
- The files will maintain their structure and formatting, with only the `moved` blocks removed.

## Edge Cases

The `edge_cases` directory demonstrates how the tool handles various special cases:

- `empty.tf`: An empty file (tool should process it without errors)
- `only_moved.tf`: A file containing only `moved` blocks (will become empty after processing)
- `commented.tf`: A file with commented out `moved` blocks (comments should remain untouched)
