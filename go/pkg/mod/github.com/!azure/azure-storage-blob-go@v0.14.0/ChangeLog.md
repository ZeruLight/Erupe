# Change Log

> See [BreakingChanges](BreakingChanges.md) for a detailed list of API breaks.

## Version 0.14.0:
- Updated [Get Blob Tags](https://docs.microsoft.com/en-us/rest/api/storageservices/get-blob-tags) and [Set Blob Tags](https://docs.microsoft.com/en-us/rest/api/storageservices/set-blob-tags) function signatures
- Added [Put Blob From URL](https://docs.microsoft.com/en-us/rest/api/storageservices/put-blob-from-url)
- Offer knob to disable application logging (Syslog)
- Added examples for MSI Login
- Updated go.mod to address dependency issues
- Fixed issues [#260](https://github.com/Azure/azure-storage-blob-go/issues/260) and [#257](https://github.com/Azure/azure-storage-blob-go/issues/257)

## Version 0.13.0:
- Validate echoed client request ID from the service
- Added new TransferManager option for UploadStreamToBlockBlob to fine-tune the concurrency and memory usage 

## Version 0.12.0:
- Added support for [Customer Provided Key](https://docs.microsoft.com/en-us/azure/storage/common/storage-service-encryption) which will let users encrypt their data within client applications before uploading to Azure Storage, and decrypting data while downloading to the client
    - Read here to know more about [Azure key vault](https://docs.microsoft.com/en-us/azure/key-vault/general/overview), [Encryption scope](https://docs.microsoft.com/en-us/azure/storage/blobs/encryption-scope-manage?tabs=portal), [managing encryption scope](https://docs.microsoft.com/en-us/azure/storage/blobs/encryption-scope-manage?tabs=portal), and how to [configure customer managed keys](https://docs.microsoft.com/en-us/azure/data-explorer/customer-managed-keys-portal)
- Stopped using memory-mapped files and switched to the `io.ReaderAt` and `io.WriterAt` interfaces. Please refer [this](https://github.com/Azure/azure-storage-blob-go/pull/223/commits/0e3e7a4e260c059c49a418a0f1501452d3e05a44) to know more
- Fixed issue [#214](https://github.com/Azure/azure-storage-blob-go/issues/214)
- Fixed issue [#230](https://github.com/Azure/azure-storage-blob-go/issues/230)

## Version 0.11.0:
- Added support for the service version [`2019-12-12`](https://docs.microsoft.com/en-us/rest/api/storageservices/versioning-for-the-azure-storage-services).
- Added [Get Blob Tags](https://docs.microsoft.com/en-us/rest/api/storageservices/get-blob-tags) and [Set Blob Tags](https://docs.microsoft.com/en-us/rest/api/storageservices/set-blob-tags) APIs which allow user-defined tags to be added to a blob which then act as a secondary index.
- Added [Find Blobs by Tags](https://docs.microsoft.com/en-us/rest/api/storageservices/find-blobs-by-tags) API which allow blobs to be retrieved based upon their tags.
- The maximum size of a block uploaded via [Put Block](https://docs.microsoft.com/en-us/rest/api/storageservices/put-block#remarks) has been increased to 4 GiB (4000 MiB). This means that the maximum size of a block blob is now approximately 200 TiB.
- The maximum size for a blob uploaded through [Put Blob](https://docs.microsoft.com/en-us/rest/api/storageservices/put-blob#remarks) has been increased to 5 GiB (5000 MiB).
- Added Blob APIs to support [Blob Versioning](https://docs.microsoft.com/en-us/azure/storage/blobs/versioning-overview) feature.
- Added support for setting blob tier directly at the time of blob creation instead of separate [Set Blob Tier](https://docs.microsoft.com/en-us/rest/api/storageservices/set-blob-tier) API call.
- Added [Get Page Range Diff](https://docs.microsoft.com/rest/api/storageservices/get-page-ranges) API to get the collection of page ranges that differ between a specified snapshot and this page blob representing managed disk.

## Version 0.10.0:
- Added support for CopyBlobFromURL (sync) and upgrade version to 2019-02-02.
- Provided default values for UploadStreamToBlockBlobOptions and refactored UploadStreamToBlockBlob.
- Added support for multiple start/expiry time formats.
- Added Solaris support.
- Enabled recovering from a unexpectedEOF error.

## Version 0.9.0:
- Updated go.mod to fix dependency issues.

## Version 0.8.0:
- Fixed error handling in high-level function DoBatchTransfer, and made it public for easy customization

## Version 0.7.0:
- Added the ability to obtain User Delegation Keys (UDK)
- Added the ability to create User Delegation SAS tokens from UDKs
- Added support for generating and using blob snapshot SAS tokens
- General secondary host improvements

## Version 0.3.0:
- Removed most panics from the library. Several functions now return an error.
- Removed 2016 and 2017 service versions.
- Added support for module.
- Fixed chunking bug in highlevel function uploadStream.