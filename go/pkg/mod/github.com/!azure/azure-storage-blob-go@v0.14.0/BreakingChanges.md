# Breaking Changes

> See the [Change Log](ChangeLog.md) for a summary of storage library changes.

## Version 0.12.0:
- Added [`ClientProvidedKeyOptions`](https://github.com/Azure/azure-storage-blob-go/blob/dev/azblob/request_common.go#L11) in function signatures. 

## Version 0.3.0:
- Removed most panics from the library. Several functions now return an error.
- Removed 2016 and 2017 service versions.