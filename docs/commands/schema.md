# smoke schema

Export the assertion type schema as structured JSON.

## Usage

```bash
smoke schema
```

## Description

Outputs all assertion types, their fields, and required flags as JSON. Useful for editor integrations, autocomplete providers, and tooling that needs to understand the config schema.

## Examples

```bash
smoke schema                           # Print full schema to stdout
smoke schema > schema.json             # Save to file
smoke schema | jq '.assertions[] | .name'  # List assertion type names
```
