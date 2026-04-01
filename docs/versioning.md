# Versioning Policy

ccpr follows [Semantic Versioning 2.0.0](https://semver.org/).

## Version format

`MAJOR.MINOR.PATCH`

- **MAJOR** — breaking changes to CLI interface or JSON output contract
- **MINOR** — new commands, flags, or JSON fields (backward-compatible)
- **PATCH** — bug fixes, documentation, internal improvements

## v1.x stability guarantees

Within the v1.x release series, the following are guaranteed:

1. **Existing JSON field names are not changed**
2. **Existing JSON field types are not changed**
3. **Required JSON fields are not removed**
4. **New fields are added in a backward-compatible way only**
5. **Exit codes (0, 1, 2) retain their defined semantics**
6. **Existing CLI flags are not renamed or removed**

Consumers should ignore unknown fields in JSON output. New fields may be added in any minor release.

## What counts as a breaking change

| Change | Breaking? |
|--------|-----------|
| Rename a JSON field | Yes |
| Remove a JSON field | Yes |
| Change a field's type (e.g., string → number) | Yes |
| Change exit code semantics | Yes |
| Rename or remove a CLI flag | Yes |
| Rename a command | Yes |
| Add a new JSON field | No |
| Add a new command | No |
| Add a new optional flag | No |
| Change error message wording | No |
| Change summary (non-JSON) output format | No |

## Deprecation policy

If a feature needs to be removed:

1. Mark it as deprecated in a minor release (documented in release notes)
2. Keep it functional for at least one minor release cycle
3. Remove it in the next major release

## JSON output contract

The JSON output contract is defined in [json-schema.md](json-schema.md) and enforced by golden tests in CI. Any change that would alter the golden test output requires explicit review.
