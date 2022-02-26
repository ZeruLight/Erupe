package azblob

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	chk "gopkg.in/check.v1" // go get gopkg.in/check.v1
)

/*
Azure Storage supports following operations support of sending customer-provided encryption keys on a request:
Put Blob, Put Block List, Put Block, Put Block from URL, Put Page, Put Page from URL, Append Block,
Set Blob Properties, Set Blob Metadata, Get Blob, Get Blob Properties, Get Blob Metadata, Snapshot Blob.
*/
var testEncryptedKey = "MDEyMzQ1NjcwMTIzNDU2NzAxMjM0NTY3MDEyMzQ1Njc="
var testEncryptedHash = "3QFFFpRA5+XANHqwwbT4yXDmrT/2JaLt/FKHjzhOdoE="
var testEncryptedScope = ""
var testCPK = NewClientProvidedKeyOptions(&testEncryptedKey, &testEncryptedHash, &testEncryptedScope)

var testEncryptedScope1 = "blobgokeytestscope"
var testCPK1 = ClientProvidedKeyOptions{EncryptionScope: &testEncryptedScope1}

func blockIDBinaryToBase64(blockID []byte) string {
	return base64.StdEncoding.EncodeToString(blockID)
}

func blockIDBase64ToBinary(blockID string) []byte {
	binary, _ := base64.StdEncoding.DecodeString(blockID)
	return binary
}

// blockIDIntToBase64 functions convert an int block ID to a base-64 string and vice versa
func blockIDIntToBase64(blockID int) string {
	binaryBlockID := (&[4]byte{})[:] // All block IDs are 4 bytes long
	binary.LittleEndian.PutUint32(binaryBlockID, uint32(blockID))
	return blockIDBinaryToBase64(binaryBlockID)
}

//func blockIDBase64ToInt(blockID string) int {
//	blockIDBase64ToBinary(blockID)
//	return int(binary.LittleEndian.Uint32(blockIDBase64ToBinary(blockID)))
//}

func (s *aztestsSuite) TestPutBlockAndPutBlockListWithCPK(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	blobURL := container.NewBlockBlobURL(generateBlobName())

	words := []string{"AAA ", "BBB ", "CCC "}
	base64BlockIDs := make([]string, len(words))
	for index, word := range words {
		base64BlockIDs[index] = blockIDIntToBase64(index)
		_, err := blobURL.StageBlock(ctx, base64BlockIDs[index], strings.NewReader(word), LeaseAccessConditions{}, nil, testCPK)
		c.Assert(err, chk.IsNil)
	}

	resp, err := blobURL.CommitBlockList(ctx, base64BlockIDs, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, testCPK)
	c.Assert(err, chk.IsNil)

	c.Assert(resp.ETag(), chk.NotNil)
	c.Assert(resp.LastModified(), chk.NotNil)
	c.Assert(resp.IsServerEncrypted(), chk.Equals, "true")
	c.Assert(resp.EncryptionKeySha256(), chk.DeepEquals, *(testCPK.EncryptionKeySha256))

	// Get blob content without encryption key should fail the request.
	_, err = blobURL.Download(ctx, 0, 0, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	// Download blob to do data integrity check.
	getResp, err := blobURL.Download(ctx, 0, 0, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)
	b := bytes.Buffer{}
	reader := getResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK})
	b.ReadFrom(reader)
	reader.Close() // The client must close the response body when finished with it
	c.Assert(b.String(), chk.Equals, "AAA BBB CCC ")
	c.Assert(getResp.ETag(), chk.Equals, resp.ETag())
	c.Assert(getResp.LastModified(), chk.DeepEquals, resp.LastModified())
}

func (s *aztestsSuite) TestPutBlockAndPutBlockListWithCPKByScope(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	blobURL := container.NewBlockBlobURL(generateBlobName())

	words := []string{"AAA ", "BBB ", "CCC "}
	base64BlockIDs := make([]string, len(words))
	for index, word := range words {
		base64BlockIDs[index] = blockIDIntToBase64(index)
		_, err := blobURL.StageBlock(ctx, base64BlockIDs[index], strings.NewReader(word), LeaseAccessConditions{}, nil, testCPK1)
		c.Assert(err, chk.IsNil)
	}

	resp, err := blobURL.CommitBlockList(ctx, base64BlockIDs, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.ETag(), chk.NotNil)
	c.Assert(resp.LastModified(), chk.NotNil)
	c.Assert(resp.IsServerEncrypted(), chk.Equals, "true")
	c.Assert(resp.EncryptionScope(), chk.Equals, *(testCPK1.EncryptionScope))

	getResp, err := blobURL.Download(ctx, 0, 0, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.NotNil)
	serr := err.(StorageError)
	c.Assert(serr.Response().StatusCode, chk.Equals, 409)
	c.Assert(serr.ServiceCode(), chk.Equals, ServiceCodeFeatureEncryptionMismatch)

	getResp, err = blobURL.Download(ctx, 0, 0, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	b := bytes.Buffer{}
	reader := getResp.Body(RetryReaderOptions{})
	b.ReadFrom(reader)
	reader.Close() // The client must close the response body when finished with it
	c.Assert(b.String(), chk.Equals, "AAA BBB CCC ")
	c.Assert(getResp.ETag(), chk.Equals, resp.ETag())
	c.Assert(getResp.LastModified(), chk.DeepEquals, resp.LastModified())
	c.Assert(getResp.LastModified(), chk.DeepEquals, resp.LastModified())
	c.Assert(getResp.r.rawResponse.Header.Get("x-ms-encryption-scope"), chk.Equals, *(testCPK1.EncryptionScope))

	// Download blob to do data integrity check.
	getResp, err = blobURL.Download(ctx, 0, 0, BlobAccessConditions{}, false, testCPK1)
	c.Assert(err, chk.IsNil)
	b = bytes.Buffer{}
	reader = getResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK1})
	b.ReadFrom(reader)
	reader.Close() // The client must close the response body when finished with it
	c.Assert(b.String(), chk.Equals, "AAA BBB CCC ")
	c.Assert(getResp.ETag(), chk.Equals, resp.ETag())
	c.Assert(getResp.LastModified(), chk.DeepEquals, resp.LastModified())
	c.Assert(getResp.r.rawResponse.Header.Get("x-ms-encryption-scope"), chk.Equals, *(testCPK1.EncryptionScope))
}

func (s *aztestsSuite) TestPutBlockFromURLAndCommitWithCPK(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 2 * 1024 // 2KB
	r, srcData := getRandomDataAndReader(testSize)
	ctx := context.Background()
	blobURL := container.NewBlockBlobURL(generateBlobName())

	uploadSrcResp, err := blobURL.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	srcBlobParts := NewBlobURLParts(blobURL.URL())

	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(1 * time.Hour),
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()
	destBlob := container.NewBlockBlobURL(generateBlobName())
	blockID1, blockID2 := blockIDIntToBase64(0), blockIDIntToBase64(1)
	stageResp1, err := destBlob.StageBlockFromURL(ctx, blockID1, srcBlobURLWithSAS, 0, 1*1024, LeaseAccessConditions{}, ModifiedAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(stageResp1.Response().StatusCode, chk.Equals, 201)
	c.Assert(stageResp1.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.Version(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.Date().IsZero(), chk.Equals, false)
	c.Assert(stageResp1.IsServerEncrypted(), chk.Equals, "true")

	stageResp2, err := destBlob.StageBlockFromURL(ctx, blockID2, srcBlobURLWithSAS, 1*1024, CountToEnd, LeaseAccessConditions{}, ModifiedAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(stageResp2.Response().StatusCode, chk.Equals, 201)
	c.Assert(stageResp2.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.Version(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.Date().IsZero(), chk.Equals, false)
	c.Assert(stageResp2.IsServerEncrypted(), chk.Equals, "true")

	blockList, err := destBlob.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 2)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 0)

	listResp, err := destBlob.CommitBlockList(ctx, []string{blockID1, blockID2}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(listResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(listResp.IsServerEncrypted(), chk.Equals, "true")

	blockList, err = destBlob.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 0)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 2)

	// Get blob content without encryption key should fail the request.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	// Download blob to do data integrity check.
	downloadResp, err = destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, srcData)
}

func (s *aztestsSuite) TestPutBlockFromURLAndCommitWithCPKWithScope(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 2 * 1024 // 2KB
	r, srcData := getRandomDataAndReader(testSize)
	ctx := context.Background()
	blobURL := container.NewBlockBlobURL(generateBlobName())

	uploadSrcResp, err := blobURL.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	srcBlobParts := NewBlobURLParts(blobURL.URL())

	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(1 * time.Hour),
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()
	destBlob := container.NewBlockBlobURL(generateBlobName())
	blockID1, blockID2 := blockIDIntToBase64(0), blockIDIntToBase64(1)
	stageResp1, err := destBlob.StageBlockFromURL(ctx, blockID1, srcBlobURLWithSAS, 0, 1*1024, LeaseAccessConditions{}, ModifiedAccessConditions{}, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(stageResp1.Response().StatusCode, chk.Equals, 201)
	c.Assert(stageResp1.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.Version(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.Date().IsZero(), chk.Equals, false)
	c.Assert(stageResp1.IsServerEncrypted(), chk.Equals, "true")

	stageResp2, err := destBlob.StageBlockFromURL(ctx, blockID2, srcBlobURLWithSAS, 1*1024, CountToEnd, LeaseAccessConditions{}, ModifiedAccessConditions{}, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(stageResp2.Response().StatusCode, chk.Equals, 201)
	c.Assert(stageResp2.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.Version(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.Date().IsZero(), chk.Equals, false)
	c.Assert(stageResp2.IsServerEncrypted(), chk.Equals, "true")

	blockList, err := destBlob.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 2)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 0)

	listResp, err := destBlob.CommitBlockList(ctx, []string{blockID1, blockID2}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(listResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(listResp.IsServerEncrypted(), chk.Equals, "true")
	c.Assert(listResp.EncryptionScope(), chk.Equals, *(testCPK1.EncryptionScope))

	blockList, err = destBlob.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 0)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 2)

	// Download blob to do data integrity check.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK1)
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, srcData)
	c.Assert(downloadResp.r.rawResponse.Header.Get("x-ms-encryption-scope"), chk.Equals, *(testCPK1.EncryptionScope))
}

func (s *aztestsSuite) TestUploadBlobWithMD5WithCPK(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 1 * 1024 * 1024
	r, srcData := getRandomDataAndReader(testSize)
	md5Val := md5.Sum(srcData)
	blobURL := container.NewBlockBlobURL(generateBlobName())

	uploadSrcResp, err := blobURL.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	// Get blob content without encryption key should fail the request.
	downloadResp, err := blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	// Download blob to do data integrity check.
	downloadResp, err = blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(downloadResp.ContentMD5(), chk.DeepEquals, md5Val[:])
	data, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(data, chk.DeepEquals, srcData)
}

func (s *aztestsSuite) TestAppendBlockWithCPK(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	appendBlobURL := container.NewAppendBlobURL(generateBlobName())

	resp, err := appendBlobURL.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.StatusCode(), chk.Equals, 201)

	words := []string{"AAA ", "BBB ", "CCC "}
	for index, word := range words {
		resp, err := appendBlobURL.AppendBlock(context.Background(), strings.NewReader(word), AppendBlobAccessConditions{}, nil, testCPK)
		c.Assert(err, chk.IsNil)
		c.Assert(err, chk.IsNil)
		c.Assert(resp.Response().StatusCode, chk.Equals, 201)
		c.Assert(resp.BlobAppendOffset(), chk.Equals, strconv.Itoa(index*4))
		c.Assert(resp.BlobCommittedBlockCount(), chk.Equals, int32(index+1))
		c.Assert(resp.ETag(), chk.Not(chk.Equals), ETagNone)
		c.Assert(resp.LastModified().IsZero(), chk.Equals, false)
		c.Assert(resp.ContentMD5(), chk.Not(chk.Equals), "")
		c.Assert(resp.RequestID(), chk.Not(chk.Equals), "")
		c.Assert(resp.Version(), chk.Not(chk.Equals), "")
		c.Assert(resp.Date().IsZero(), chk.Equals, false)
		c.Assert(resp.IsServerEncrypted(), chk.Equals, "true")
		c.Assert(resp.EncryptionKeySha256(), chk.Equals, *(testCPK.EncryptionKeySha256))
	}

	// Get blob content without encryption key should fail the request.
	_, err = appendBlobURL.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	// Download blob to do data integrity check.
	downloadResp, err := appendBlobURL.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)

	data, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(string(data), chk.DeepEquals, "AAA BBB CCC ")
}

func (s *aztestsSuite) TestAppendBlockWithCPKByScope(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	appendBlobURL := container.NewAppendBlobURL(generateBlobName())

	resp, err := appendBlobURL.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.StatusCode(), chk.Equals, 201)

	words := []string{"AAA ", "BBB ", "CCC "}
	for index, word := range words {
		resp, err := appendBlobURL.AppendBlock(context.Background(), strings.NewReader(word), AppendBlobAccessConditions{}, nil, testCPK1)
		c.Assert(err, chk.IsNil)
		c.Assert(err, chk.IsNil)
		c.Assert(resp.Response().StatusCode, chk.Equals, 201)
		c.Assert(resp.BlobAppendOffset(), chk.Equals, strconv.Itoa(index*4))
		c.Assert(resp.BlobCommittedBlockCount(), chk.Equals, int32(index+1))
		c.Assert(resp.ETag(), chk.Not(chk.Equals), ETagNone)
		c.Assert(resp.LastModified().IsZero(), chk.Equals, false)
		c.Assert(resp.ContentMD5(), chk.Not(chk.Equals), "")
		c.Assert(resp.RequestID(), chk.Not(chk.Equals), "")
		c.Assert(resp.Version(), chk.Not(chk.Equals), "")
		c.Assert(resp.Date().IsZero(), chk.Equals, false)
		c.Assert(resp.IsServerEncrypted(), chk.Equals, "true")
		c.Assert(resp.EncryptionScope(), chk.Equals, *(testCPK1.EncryptionScope))
	}

	// Download blob to do data integrity check.
	downloadResp, err := appendBlobURL.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(downloadResp.IsServerEncrypted(), chk.Equals, "true")

	data, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK1}))
	c.Assert(err, chk.IsNil)
	c.Assert(string(data), chk.DeepEquals, "AAA BBB CCC ")
	c.Assert(downloadResp.r.rawResponse.Header.Get("x-ms-encryption-scope"), chk.Equals, *(testCPK1.EncryptionScope))
}

func (s *aztestsSuite) TestAppendBlockFromURLWithCPK(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 2 * 1024 * 1024 // 2MB
	r, srcData := getRandomDataAndReader(testSize)
	ctx := context.Background() // Use default Background context
	blobURL := container.NewAppendBlobURL(generateName("src"))
	destBlob := container.NewAppendBlobURL(generateName("dest"))

	cResp1, err := blobURL.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(cResp1.StatusCode(), chk.Equals, 201)

	resp, err := blobURL.AppendBlock(context.Background(), r, AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(resp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(resp.ContentMD5(), chk.Not(chk.Equals), "")

	srcBlobParts := NewBlobURLParts(blobURL.URL())

	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(1 * time.Hour),
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()

	cResp2, err := destBlob.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(cResp2.StatusCode(), chk.Equals, 201)

	appendResp, err := destBlob.AppendBlockFromURL(ctx, srcBlobURLWithSAS, 0, int64(testSize), AppendBlobAccessConditions{}, ModifiedAccessConditions{}, nil, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(appendResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(appendResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(appendResp.IsServerEncrypted(), chk.Equals, "true")

	// Get blob content without encryption key should fail the request.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	// Download blob to do data integrity check.
	downloadResp, err = destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, srcData)
}

func (s *aztestsSuite) TestPageBlockWithCPK(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 1 * 1024 * 1024
	r, srcData := getRandomDataAndReader(testSize)
	blobURL, _ := createNewPageBlobWithCPK(c, container, int64(testSize), testCPK)

	uploadResp, err := blobURL.UploadPages(ctx, 0, r, PageBlobAccessConditions{}, nil, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(uploadResp.Response().StatusCode, chk.Equals, 201)

	// Get blob content without encryption key should fail the request.
	downloadResp, err := blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	// Download blob to do data integrity check.
	downloadResp, err = blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, srcData)
}

func (s *aztestsSuite) TestPageBlockWithCPKByScope(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	// defer delContainer(c, container)

	testSize := 1 * 1024 * 1024
	r, srcData := getRandomDataAndReader(testSize)
	blobURL, _ := createNewPageBlobWithCPK(c, container, int64(testSize), testCPK1)

	uploadResp, err := blobURL.UploadPages(ctx, 0, r, PageBlobAccessConditions{}, nil, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(uploadResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(uploadResp.EncryptionScope(), chk.Equals, *(testCPK1.EncryptionScope))

	// Download blob to do data integrity check.
	downloadResp, err := blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK1)
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK1}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, srcData)
	c.Assert(downloadResp.r.rawResponse.Header.Get("x-ms-encryption-scope"), chk.Equals, *(testCPK1.EncryptionScope))
}

func (s *aztestsSuite) TestPageBlockFromURLWithCPK(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 1 * 1024 * 1024 // 1MB
	r, srcData := getRandomDataAndReader(testSize)
	ctx := context.Background() // Use default Background context
	blobURL, _ := createNewPageBlobWithSize(c, container, int64(testSize))
	destBlob, _ := createNewPageBlobWithCPK(c, container, int64(testSize), testCPK)

	uploadResp, err := blobURL.UploadPages(ctx, 0, r, PageBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadResp.Response().StatusCode, chk.Equals, 201)
	srcBlobParts := NewBlobURLParts(blobURL.URL())

	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(1 * time.Hour),
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()

	resp, err := destBlob.UploadPagesFromURL(ctx, srcBlobURLWithSAS, 0, 0, int64(testSize), nil, PageBlobAccessConditions{}, ModifiedAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.ETag(), chk.NotNil)
	c.Assert(resp.LastModified(), chk.NotNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 201)
	c.Assert(resp.IsServerEncrypted(), chk.Equals, "true")

	// Download blob to do data integrity check.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(downloadResp.r.EncryptionKeySha256(), chk.Equals, *(testCPK.EncryptionKeySha256))
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, srcData)
}

func (s *aztestsSuite) TestUploadPagesFromURLWithMD5WithCPK(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 1 * 1024 * 1024
	r, srcData := getRandomDataAndReader(testSize)
	md5Value := md5.Sum(srcData)
	srcBlob, _ := createNewPageBlobWithSize(c, container, int64(testSize))

	uploadSrcResp1, err := srcBlob.UploadPages(ctx, 0, r, PageBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp1.Response().StatusCode, chk.Equals, 201)

	srcBlobParts := NewBlobURLParts(srcBlob.URL())

	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(1 * time.Hour),
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()
	destBlob, _ := createNewPageBlobWithCPK(c, container, int64(testSize), testCPK)
	uploadResp, err := destBlob.UploadPagesFromURL(ctx, srcBlobURLWithSAS, 0, 0, int64(testSize), md5Value[:], PageBlobAccessConditions{}, ModifiedAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(uploadResp.ETag(), chk.NotNil)
	c.Assert(uploadResp.LastModified(), chk.NotNil)
	c.Assert(uploadResp.EncryptionKeySha256(), chk.Equals, *(testCPK.EncryptionKeySha256))
	c.Assert(uploadResp.ContentMD5(), chk.DeepEquals, md5Value[:])
	c.Assert(uploadResp.BlobSequenceNumber(), chk.Equals, int64(0))

	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(downloadResp.r.EncryptionKeySha256(), chk.Equals, *(testCPK.EncryptionKeySha256))
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{ClientProvidedKeyOptions: testCPK}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, srcData)

	_, badMD5 := getRandomDataAndReader(16)
	_, err = destBlob.UploadPagesFromURL(ctx, srcBlobURLWithSAS, 0, 0, int64(testSize), badMD5[:], PageBlobAccessConditions{}, ModifiedAccessConditions{}, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeMd5Mismatch)
}

func (s *aztestsSuite) TestGetSetBlobMetadataWithCPK(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlobWithCPK(c, containerURL, testCPK)

	metadata := Metadata{"key": "value", "another_key": "1234"}

	// Set blob metadata without encryption key should fail the request.
	_, err := blobURL.SetMetadata(ctx, metadata, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	resp, err := blobURL.SetMetadata(ctx, metadata, BlobAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.EncryptionKeySha256(), chk.Equals, *(testCPK.EncryptionKeySha256))

	// Get blob properties without encryption key should fail the request.
	getResp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	getResp, err = blobURL.GetProperties(ctx, BlobAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(getResp.NewMetadata(), chk.HasLen, 2)
	c.Assert(getResp.NewMetadata(), chk.DeepEquals, metadata)

	_, err = blobURL.SetMetadata(ctx, Metadata{}, BlobAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)

	getResp, err = blobURL.GetProperties(ctx, BlobAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(getResp.NewMetadata(), chk.HasLen, 0)
}

func (s *aztestsSuite) TestGetSetBlobMetadataWithCPKByScope(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlobWithCPK(c, containerURL, testCPK1)

	metadata := Metadata{"key": "value", "another_key": "1234"}

	// Set blob metadata without encryption key should fail the request.
	_, err := blobURL.SetMetadata(ctx, metadata, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	_, err = blobURL.SetMetadata(ctx, metadata, BlobAccessConditions{}, testCPK1)
	c.Assert(err, chk.IsNil)

	getResp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(getResp.NewMetadata(), chk.HasLen, 2)
	c.Assert(getResp.NewMetadata(), chk.DeepEquals, metadata)

	_, err = blobURL.SetMetadata(ctx, Metadata{}, BlobAccessConditions{}, testCPK1)
	c.Assert(err, chk.IsNil)

	getResp, err = blobURL.GetProperties(ctx, BlobAccessConditions{}, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(getResp.NewMetadata(), chk.HasLen, 0)
}

func (s *aztestsSuite) TestBlobSnapshotWithCPK(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlobWithCPK(c, containerURL, testCPK)
	_, err := blobURL.Upload(ctx, strings.NewReader("113333555555"), BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, testCPK)

	// Create Snapshot of an encrypted blob without encryption key should fail the request.
	resp, err := blobURL.CreateSnapshot(ctx, nil, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	resp, err = blobURL.CreateSnapshot(ctx, nil, BlobAccessConditions{}, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.IsServerEncrypted(), chk.Equals, "false")
	snapshotURL := blobURL.WithSnapshot(resp.Snapshot())

	dResp, err := snapshotURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK)
	c.Assert(err, chk.IsNil)
	c.Assert(dResp.r.EncryptionKeySha256(), chk.Equals, *(testCPK.EncryptionKeySha256))
	_, err = snapshotURL.Delete(ctx, DeleteSnapshotsOptionNone, BlobAccessConditions{})
	c.Assert(err, chk.IsNil)

	// Get blob properties of snapshot without encryption key should fail the request.
	_, err = snapshotURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)
	c.Assert(err.(StorageError).Response().StatusCode, chk.Equals, 404)
}

func (s *aztestsSuite) TestBlobSnapshotWithCPKByScope(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlobWithCPK(c, containerURL, testCPK)
	_, err := blobURL.Upload(ctx, strings.NewReader("113333555555"), BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, testCPK1)

	// Create Snapshot of an encrypted blob without encryption key should fail the request.
	resp, err := blobURL.CreateSnapshot(ctx, nil, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	resp, err = blobURL.CreateSnapshot(ctx, nil, BlobAccessConditions{}, testCPK1)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.IsServerEncrypted(), chk.Equals, "false")
	snapshotURL := blobURL.WithSnapshot(resp.Snapshot())

	_, err = snapshotURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, testCPK1)
	c.Assert(err, chk.IsNil)
	_, err = snapshotURL.Delete(ctx, DeleteSnapshotsOptionNone, BlobAccessConditions{})
	c.Assert(err, chk.IsNil)

	// Get blob properties of snapshot without encryption key should fail the request.
	_, err = snapshotURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)
	c.Assert(err.(StorageError).Response().StatusCode, chk.Equals, 404)
}
