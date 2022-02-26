package azblob

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"time"

	"crypto/md5"

	"bytes"
	"strings"

	guuid "github.com/google/uuid"
	chk "gopkg.in/check.v1" // go get gopkg.in/check.v1
)

func (s *aztestsSuite) TestStageGetBlocks(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	blob := container.NewBlockBlobURL(generateBlobName())

	blockID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%6d", 0)))

	putResp, err := blob.StageBlock(context.Background(), blockID, getReaderToRandomBytes(1024), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(putResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(putResp.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(putResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(putResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(putResp.Date().IsZero(), chk.Equals, false)

	blockList, err := blob.GetBlockList(context.Background(), BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.LastModified().IsZero(), chk.Equals, true)
	c.Assert(blockList.ETag(), chk.Equals, ETagNone)
	c.Assert(blockList.ContentType(), chk.Not(chk.Equals), "")
	c.Assert(blockList.BlobContentLength(), chk.Equals, int64(-1))
	c.Assert(blockList.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(blockList.Version(), chk.Not(chk.Equals), "")
	c.Assert(blockList.Date().IsZero(), chk.Equals, false)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 0)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 1)

	listResp, err := blob.CommitBlockList(context.Background(), []string{blockID}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(listResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(listResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(listResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(listResp.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(listResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(listResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(listResp.Date().IsZero(), chk.Equals, false)

	blockList, err = blob.GetBlockList(context.Background(), BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.LastModified().IsZero(), chk.Equals, false)
	c.Assert(blockList.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(blockList.ContentType(), chk.Not(chk.Equals), "")
	c.Assert(blockList.BlobContentLength(), chk.Equals, int64(1024))
	c.Assert(blockList.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(blockList.Version(), chk.Not(chk.Equals), "")
	c.Assert(blockList.Date().IsZero(), chk.Equals, false)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 1)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 0)
}

func (s *aztestsSuite) TestStageBlockFromURL(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 8 * 1024 * 1024 // 8MB
	r, sourceData := getRandomDataAndReader(testSize)
	ctx := context.Background() // Use default Background context
	srcBlob := container.NewBlockBlobURL(generateBlobName())
	destBlob := container.NewBlockBlobURL(generateBlobName())

	// Prepare source blob for copy.
	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	// Get source blob URL with SAS for StageFromURL.
	srcBlobParts := NewBlobURLParts(srcBlob.URL())

	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,                     // Users MUST use HTTPS (not HTTP)
		ExpiryTime:    time.Now().UTC().Add(48 * time.Hour), // 48-hours before expiration
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()

	// Stage blocks from URL.
	blockID1, blockID2 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%6d", 0))), base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%6d", 1)))
	stageResp1, err := destBlob.StageBlockFromURL(ctx, blockID1, srcBlobURLWithSAS, 0, 4*1024*1024, LeaseAccessConditions{}, ModifiedAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(stageResp1.Response().StatusCode, chk.Equals, 201)
	c.Assert(stageResp1.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.Version(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.Date().IsZero(), chk.Equals, false)

	stageResp2, err := destBlob.StageBlockFromURL(ctx, blockID2, srcBlobURLWithSAS, 4*1024*1024, CountToEnd, LeaseAccessConditions{}, ModifiedAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(stageResp2.Response().StatusCode, chk.Equals, 201)
	c.Assert(stageResp2.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.Version(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.Date().IsZero(), chk.Equals, false)

	// Check block list.
	blockList, err := destBlob.GetBlockList(context.Background(), BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 0)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 2)

	// Commit block list.
	listResp, err := destBlob.CommitBlockList(context.Background(), []string{blockID1, blockID2}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(listResp.Response().StatusCode, chk.Equals, 201)

	// Check data integrity through downloading.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)
}

func (s *aztestsSuite) TestCopyBlockBlobFromURL(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 8 * 1024 * 1024 // 8MB
	r, sourceData := getRandomDataAndReader(testSize)
	sourceDataMD5Value := md5.Sum(sourceData)
	ctx := context.Background() // Use default Background context
	srcBlob := container.NewBlockBlobURL(generateBlobName())
	destBlob := container.NewBlockBlobURL(generateBlobName())

	// Prepare source blob for copy.
	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	// Get source blob URL with SAS for StageFromURL.
	srcBlobParts := NewBlobURLParts(srcBlob.URL())

	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,                     // Users MUST use HTTPS (not HTTP)
		ExpiryTime:    time.Now().UTC().Add(48 * time.Hour), // 48-hours before expiration
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()

	// Invoke copy blob from URL.
	resp, err := destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{"foo": "bar"}, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], DefaultAccessTier, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 202)
	c.Assert(resp.ETag(), chk.Not(chk.Equals), "")
	c.Assert(resp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(resp.Version(), chk.Not(chk.Equals), "")
	c.Assert(resp.Date().IsZero(), chk.Equals, false)
	c.Assert(resp.CopyID(), chk.Not(chk.Equals), "")
	c.Assert(resp.ContentMD5(), chk.DeepEquals, sourceDataMD5Value[:])
	c.Assert(string(resp.CopyStatus()), chk.DeepEquals, "success")

	// Check data integrity through downloading.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)

	// Make sure the metadata got copied over
	c.Assert(len(downloadResp.NewMetadata()), chk.Equals, 1)

	// Edge case 1: Provide bad MD5 and make sure the copy fails
	_, badMD5 := getRandomDataAndReader(16)
	_, err = destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, badMD5, DefaultAccessTier, nil)
	c.Assert(err, chk.NotNil)

	// Edge case 2: Not providing any source MD5 should see the CRC getting returned instead
	resp, err = destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, nil, DefaultAccessTier, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 202)
	c.Assert(resp.XMsContentCrc64(), chk.Not(chk.Equals), "")
}

func (s *aztestsSuite) TestBlobSASQueryParamOverrideResponseHeaders(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 8 * 1024 * 1024 // 8MB
	r, _ := getRandomDataAndReader(testSize)
	ctx := context.Background() // Use default Background context
	blob := container.NewBlockBlobURL(generateBlobName())

	uploadResp, err := blob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadResp.Response().StatusCode, chk.Equals, 201)

	// Get blob URL with SAS.
	blobParts := NewBlobURLParts(blob.URL())

	cacheControlVal := "cache-control-override"
	contentDispositionVal := "content-disposition-override"
	contentEncodingVal := "content-encoding-override"
	contentLanguageVal := "content-language-override"
	contentTypeVal := "content-type-override"

	blobParts.SAS, err = BlobSASSignatureValues{
		Protocol:           SASProtocolHTTPS,                     // Users MUST use HTTPS (not HTTP)
		ExpiryTime:         time.Now().UTC().Add(48 * time.Hour), // 48-hours before expiration
		ContainerName:      blobParts.ContainerName,
		BlobName:           blobParts.BlobName,
		Permissions:        BlobSASPermissions{Read: true}.String(),
		CacheControl:       cacheControlVal,
		ContentDisposition: contentDispositionVal,
		ContentEncoding:    contentEncodingVal,
		ContentLanguage:    contentLanguageVal,
		ContentType:        contentTypeVal,
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	blobURL := NewBlobURL(blobParts.URL(), NewPipeline(NewAnonymousCredential(), PipelineOptions{}))

	gResp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(gResp.CacheControl(), chk.Equals, cacheControlVal)
	c.Assert(gResp.ContentDisposition(), chk.Equals, contentDispositionVal)
	c.Assert(gResp.ContentEncoding(), chk.Equals, contentEncodingVal)
	c.Assert(gResp.ContentLanguage(), chk.Equals, contentLanguageVal)
	c.Assert(gResp.ContentType(), chk.Equals, contentTypeVal)
}

func (s *aztestsSuite) TestStageBlockWithMD5(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	blob := container.NewBlockBlobURL(generateBlobName())
	blockID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%6d", 0)))

	// test put block with valid MD5 value
	readerToBody, body := getRandomDataAndReader(1024)
	md5Value := md5.Sum(body)
	putResp, err := blob.StageBlock(context.Background(), blockID, readerToBody, LeaseAccessConditions{}, md5Value[:], ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(putResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(putResp.ContentMD5(), chk.DeepEquals, md5Value[:])
	c.Assert(putResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(putResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(putResp.Date().IsZero(), chk.Equals, false)

	// test put block with bad MD5 value
	readerToBody, body = getRandomDataAndReader(1024)
	_, badMD5 := getRandomDataAndReader(16)
	putResp, err = blob.StageBlock(context.Background(), blockID, readerToBody, LeaseAccessConditions{}, badMD5[:], ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeMd5Mismatch)
}

func (s *aztestsSuite) TestBlobPutBlobNonEmptyBody(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.Upload(ctx, strings.NewReader(blockBlobDefaultData), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.Download(ctx, 0, 0, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	data, err := ioutil.ReadAll(resp.Response().Body)
	c.Assert(string(data), chk.Equals, blockBlobDefaultData)
}

func (s *aztestsSuite) TestBlobPutBlobHTTPHeaders(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.Upload(ctx, bytes.NewReader(nil), basicHeaders, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	h := resp.NewHTTPHeaders()
	h.ContentMD5 = nil // the service generates a MD5 value, omit before comparing
	c.Assert(h, chk.DeepEquals, basicHeaders)
}

func (s *aztestsSuite) TestBlobPutBlobMetadataNotEmpty(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.NewMetadata(), chk.DeepEquals, basicMetadata)
}

func (s *aztestsSuite) TestBlobPutBlobMetadataEmpty(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.NewMetadata(), chk.HasLen, 0)
}

func (s *aztestsSuite) TestBlobPutBlobMetadataInvalid(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.Upload(ctx, nil, BlobHTTPHeaders{}, Metadata{"In valid!": "bar"}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(strings.Contains(err.Error(), validationErrorSubstring), chk.Equals, true)
}

func (s *aztestsSuite) TestBlobPutBlobIfModifiedSinceTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(-10)

	_, err := blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfModifiedSince: currentTime}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateUpload(c, blobURL)
}

func (s *aztestsSuite) TestBlobPutBlobIfModifiedSinceFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(10)

	_, err := blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfModifiedSince: currentTime}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobPutBlobIfUnmodifiedSinceTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(10)

	_, err := blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfUnmodifiedSince: currentTime}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateUpload(c, blobURL)
}

func (s *aztestsSuite) TestBlobPutBlobIfUnmodifiedSinceFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(-10)

	_, err := blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfUnmodifiedSince: currentTime}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobPutBlobIfMatchTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: resp.ETag()}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateUpload(c, blobURL)
}

func (s *aztestsSuite) TestBlobPutBlobIfMatchFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)

	_, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: ETag("garbage")}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobPutBlobIfNoneMatchTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)

	_, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfNoneMatch: ETag("garbage")}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateUpload(c, blobURL)
}

func (s *aztestsSuite) TestBlobPutBlobIfNoneMatchFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.Upload(ctx, bytes.NewReader(nil), BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfNoneMatch: resp.ETag()}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

var blockID string // a single blockID used in tests when only a single ID is needed

func init() {
	u := [64]byte{}
	binary.BigEndian.PutUint32(u[len(guuid.UUID{}):], math.MaxUint32)
	blockID = base64.StdEncoding.EncodeToString(u[:])
}

func (s *aztestsSuite) TestBlobGetBlockListNone(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetBlockList(ctx, BlockListNone, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks, chk.HasLen, 0)
	c.Assert(resp.UncommittedBlocks, chk.HasLen, 0) // Not specifying a block list type should default to only returning committed blocks
}

func (s *aztestsSuite) TestBlobGetBlockListUncommitted(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetBlockList(ctx, BlockListUncommitted, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks, chk.HasLen, 0)
	c.Assert(resp.UncommittedBlocks, chk.HasLen, 1)
}

func (s *aztestsSuite) TestBlobGetBlockListCommitted(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.CommitBlockList(ctx, []string{blockID}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})

	resp, err := blobURL.GetBlockList(ctx, BlockListCommitted, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks, chk.HasLen, 1)
	c.Assert(resp.UncommittedBlocks, chk.HasLen, 0)
}

func (s *aztestsSuite) TestBlobGetBlockListCommittedEmpty(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetBlockList(ctx, BlockListCommitted, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks, chk.HasLen, 0)
	c.Assert(resp.UncommittedBlocks, chk.HasLen, 0)
}

func (s *aztestsSuite) TestBlobGetBlockListBothEmpty(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	validateStorageError(c, err, ServiceCodeBlobNotFound)
}

func (s *aztestsSuite) TestBlobGetBlockListBothNotEmpty(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	id := newID()

	// Put and commit two blocks
	_, err := blobURL.StageBlock(ctx, id.next(), strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.StageBlock(ctx, id.next(), strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.CommitBlockList(ctx, id.issued(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	// Put two uncommitted blocks
	_, err = blobURL.StageBlock(ctx, id.next(), strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.StageBlock(ctx, id.next(), strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks[0].Name, chk.Equals, id.issued()[0])
	c.Assert(resp.CommittedBlocks[1].Name, chk.Equals, id.issued()[1])   // Committed blocks are returned in the order they are committed (in the commit list)
	c.Assert(resp.UncommittedBlocks[0].Name, chk.Equals, id.issued()[2]) // Uncommitted blocks are returned in alphabetical order
	c.Assert(resp.UncommittedBlocks[1].Name, chk.Equals, id.issued()[3])
}

func (s *aztestsSuite) TestBlobGetBlockListInvalidType(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.GetBlockList(ctx, BlockListType("garbage"), LeaseAccessConditions{})
	validateStorageError(c, err, ServiceCodeInvalidQueryParameterValue)
}

func (s *aztestsSuite) TestBlobGetBlockListSnapshot(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.CommitBlockList(ctx, []string{blockID}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.CreateSnapshot(ctx, nil, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	snapshotURL := blobURL.WithSnapshot(resp.Snapshot())

	resp2, err := snapshotURL.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp2.CommittedBlocks, chk.HasLen, 1)
}

func (s *aztestsSuite) TestBlobPutBlockIDInvalidCharacters(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, "!!", strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeInvalidQueryParameterValue)
}

func (s *aztestsSuite) TestBlobPutBlockIDInvalidLength(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.StageBlock(ctx, "00000000", strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeInvalidBlobOrBlock)
}

func (s *aztestsSuite) TestBlobPutBlockEmptyBody(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(""), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeInvalidHeaderValue)
}

func setupPutBlockListTest(c *chk.C) (containerURL ContainerURL, blobURL BlockBlobURL, id string) {
	bsu := getBSU()
	containerURL, _ = createNewContainer(c, bsu)
	blobURL, _ = getBlockBlobURL(c, containerURL)

	_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	return containerURL, blobURL, blockID
}

func (s *aztestsSuite) TestBlobPutBlockListInvalidID(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{id[:2]}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeInvalidBlockID)
}

func (s *aztestsSuite) TestBlobPutBlockListDuplicateBlocks(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{id, id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks, chk.HasLen, 2)
}

func (s *aztestsSuite) TestBlobPutBlockListEmptyList(c *chk.C) {
	containerURL, blobURL, _ := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{}, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks, chk.HasLen, 0)
}

func (s *aztestsSuite) TestBlobPutBlockListMetadataEmpty(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.NewMetadata(), chk.HasLen, 0)
}

func (s *aztestsSuite) TestBlobPutBlockListMetadataNonEmpty(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.NewMetadata(), chk.DeepEquals, basicMetadata)
}

func (s *aztestsSuite) TestBlobPutBlockListHTTPHeaders(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{id}, basicHeaders, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, _ := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	h := resp.NewHTTPHeaders()
	c.Assert(h, chk.DeepEquals, basicHeaders)
}

func (s *aztestsSuite) TestBlobPutBlockListHTTPHeadersEmpty(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{ContentDisposition: "my_disposition"}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.ContentDisposition(), chk.Equals, "")
}

func validateBlobCommitted(c *chk.C, blobURL BlockBlobURL) {
	resp, err := blobURL.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks, chk.HasLen, 1)
}

func (s *aztestsSuite) TestBlobPutBlockListIfModifiedSinceTrue(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)
	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{}) // The blob must actually exist to have a modifed time
	c.Assert(err, chk.IsNil)

	currentTime := getRelativeTimeGMT(-10)

	_, err = blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfModifiedSince: currentTime}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlobCommitted(c, blobURL)
}

func (s *aztestsSuite) TestBlobPutBlockListIfModifiedSinceFalse(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	currentTime := getRelativeTimeGMT(10)

	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfModifiedSince: currentTime}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobPutBlockListIfUnmodifiedSinceTrue(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)
	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{}) // The blob must actually exist to have a modifed time
	c.Assert(err, chk.IsNil)

	currentTime := getRelativeTimeGMT(10)

	_, err = blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfUnmodifiedSince: currentTime}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlobCommitted(c, blobURL)
}

func (s *aztestsSuite) TestBlobPutBlockListIfUnmodifiedSinceFalse(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{}) // The blob must actually exist to have a modifed time
	defer deleteContainer(c, containerURL)

	currentTime := getRelativeTimeGMT(-10)

	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfUnmodifiedSince: currentTime}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})

	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobPutBlockListIfMatchTrue(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)
	resp, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{}) // The blob must actually exist to have a modifed time
	c.Assert(err, chk.IsNil)

	_, err = blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: resp.ETag()}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlobCommitted(c, blobURL)
}

func (s *aztestsSuite) TestBlobPutBlockListIfMatchFalse(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)
	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{}) // The blob must actually exist to have a modifed time
	c.Assert(err, chk.IsNil)

	_, err = blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: ETag("garbage")}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})

	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobPutBlockListIfNoneMatchTrue(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)
	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{}) // The blob must actually exist to have a modifed time
	c.Assert(err, chk.IsNil)

	_, err = blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfNoneMatch: ETag("garbage")}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlobCommitted(c, blobURL)
}

func (s *aztestsSuite) TestBlobPutBlockListIfNoneMatchFalse(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)
	resp, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{}) // The blob must actually exist to have a modifed time
	c.Assert(err, chk.IsNil)

	_, err = blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfNoneMatch: resp.ETag()}}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})

	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobPutBlockListValidateData(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})

	resp, err := blobURL.Download(ctx, 0, 0, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	data, _ := ioutil.ReadAll(resp.Response().Body)
	c.Assert(string(data), chk.Equals, blockBlobDefaultData)
}

func (s *aztestsSuite) TestBlobPutBlockListModifyBlob(c *chk.C) {
	containerURL, blobURL, id := setupPutBlockListTest(c)
	defer deleteContainer(c, containerURL)

	_, err := blobURL.CommitBlockList(ctx, []string{id}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.StageBlock(ctx, "0001", bytes.NewReader([]byte("new data")), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.StageBlock(ctx, "0010", bytes.NewReader([]byte("new data")), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.StageBlock(ctx, "0011", bytes.NewReader([]byte("new data")), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.StageBlock(ctx, "0100", bytes.NewReader([]byte("new data")), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	_, err = blobURL.CommitBlockList(ctx, []string{"0001", "0011"}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.CommittedBlocks, chk.HasLen, 2)
	c.Assert(resp.CommittedBlocks[0].Name, chk.Equals, "0001")
	c.Assert(resp.CommittedBlocks[1].Name, chk.Equals, "0011")
	c.Assert(resp.UncommittedBlocks, chk.HasLen, 0)
}

func (s *aztestsSuite) TestSetTierOnBlobUpload(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	for _, tier := range []AccessTierType{AccessTierArchive, AccessTierCool, AccessTierHot} {
		blobURL, _ := getBlockBlobURL(c, containerURL)

		_, err := blobURL.Upload(ctx, strings.NewReader("Test Data"), basicHeaders, nil, BlobAccessConditions{}, tier, nil, ClientProvidedKeyOptions{})
		c.Assert(err, chk.IsNil)

		resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
		c.Assert(err, chk.IsNil)
		c.Assert(resp.AccessTier(), chk.Equals, string(tier))
	}
}

func (s *aztestsSuite) TestBlobSetTierOnCommit(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)

	for _, tier := range []AccessTierType{AccessTierCool, AccessTierHot} {
		blobURL, _ := getBlockBlobURL(c, containerURL)

		_, err := blobURL.StageBlock(ctx, blockID, strings.NewReader(blockBlobDefaultData), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
		c.Assert(err, chk.IsNil)

		_, err = blobURL.CommitBlockList(ctx, []string{blockID}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, tier, nil, ClientProvidedKeyOptions{})

		resp, err := blobURL.GetBlockList(ctx, BlockListCommitted, LeaseAccessConditions{})
		c.Assert(err, chk.IsNil)
		c.Assert(resp.CommittedBlocks, chk.HasLen, 1)
		c.Assert(resp.UncommittedBlocks, chk.HasLen, 0)
	}
}

func (s *aztestsSuite) TestSetTierOnCopyBlockBlobFromURL(c *chk.C) {
	bsu := getBSU()

	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 1 * 1024 * 1024
	r, sourceData := getRandomDataAndReader(testSize)
	sourceDataMD5Value := md5.Sum(sourceData)
	ctx := context.Background()
	srcBlob := container.NewBlockBlobURL(generateBlobName())

	// Setting blob tier as "cool"
	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, AccessTierCool, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	// Get source blob URL with SAS for StageFromURL.
	srcBlobParts := NewBlobURLParts(srcBlob.URL())

	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(2 * time.Hour),
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()
	for _, tier := range []AccessTierType{AccessTierArchive, AccessTierCool, AccessTierHot} {
		destBlob := container.NewBlockBlobURL(generateBlobName())
		resp, err := destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{"foo": "bar"}, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], tier, nil)
		c.Assert(err, chk.IsNil)
		c.Assert(resp.Response().StatusCode, chk.Equals, 202)
		c.Assert(string(resp.CopyStatus()), chk.DeepEquals, "success")

		destBlobPropResp, err := destBlob.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
		c.Assert(err, chk.IsNil)
		c.Assert(destBlobPropResp.AccessTier(), chk.Equals, string(tier))

	}
}

func (s *aztestsSuite) TestSetTierOnStageBlockFromURL(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 8 * 1024 * 1024 // 8MB
	r, sourceData := getRandomDataAndReader(testSize)
	ctx := context.Background() // Use default Background context
	srcBlob := container.NewBlockBlobURL("src" + generateBlobName())
	destBlob := container.NewBlockBlobURL("dst" + generateBlobName())
	tier := AccessTierCool

	// Prepare source blob for copy.
	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, tier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	// Get source blob URL with SAS for StageFromURL.
	srcBlobParts := NewBlobURLParts(srcBlob.URL())

	srcBlobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,                     // Users MUST use HTTPS (not HTTP)
		ExpiryTime:    time.Now().UTC().Add(48 * time.Hour), // 48-hours before expiration
		ContainerName: srcBlobParts.ContainerName,
		BlobName:      srcBlobParts.BlobName,
		Permissions:   BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	srcBlobURLWithSAS := srcBlobParts.URL()

	// Stage blocks from URL.
	blockID1, blockID2 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%6d", 0))), base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%6d", 1)))
	stageResp1, err := destBlob.StageBlockFromURL(ctx, blockID1, srcBlobURLWithSAS, 0, 4*1024*1024, LeaseAccessConditions{}, ModifiedAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(stageResp1.Response().StatusCode, chk.Equals, 201)
	c.Assert(stageResp1.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.Version(), chk.Not(chk.Equals), "")
	c.Assert(stageResp1.Date().IsZero(), chk.Equals, false)

	stageResp2, err := destBlob.StageBlockFromURL(ctx, blockID2, srcBlobURLWithSAS, 4*1024*1024, CountToEnd, LeaseAccessConditions{}, ModifiedAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(stageResp2.Response().StatusCode, chk.Equals, 201)
	c.Assert(stageResp2.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.Version(), chk.Not(chk.Equals), "")
	c.Assert(stageResp2.Date().IsZero(), chk.Equals, false)

	// Check block list.
	blockList, err := destBlob.GetBlockList(context.Background(), BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 0)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 2)

	// Commit block list.
	listResp, err := destBlob.CommitBlockList(context.Background(), []string{blockID1, blockID2}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, tier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(listResp.Response().StatusCode, chk.Equals, 201)

	// Check data integrity through downloading.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)

	// Get properties to validate the tier
	destBlobPropResp, err := destBlob.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(destBlobPropResp.AccessTier(), chk.Equals, string(tier))
}
