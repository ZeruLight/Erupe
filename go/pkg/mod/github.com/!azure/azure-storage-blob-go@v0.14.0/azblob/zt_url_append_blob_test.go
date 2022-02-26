package azblob

import (
	"context"
	"io/ioutil"
	"time"

	"crypto/md5"

	"bytes"
	"strings"

	chk "gopkg.in/check.v1" // go get gopkg.in/check.v1
)

func (s *aztestsSuite) TestAppendBlock(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	blob := container.NewAppendBlobURL(generateBlobName())

	resp, err := blob.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.StatusCode(), chk.Equals, 201)

	appendResp, err := blob.AppendBlock(context.Background(), getReaderToRandomBytes(1024), AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(appendResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(appendResp.BlobAppendOffset(), chk.Equals, "0")
	c.Assert(appendResp.BlobCommittedBlockCount(), chk.Equals, int32(1))
	c.Assert(appendResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(appendResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(appendResp.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.Date().IsZero(), chk.Equals, false)

	appendResp, err = blob.AppendBlock(context.Background(), getReaderToRandomBytes(1024), AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(appendResp.BlobAppendOffset(), chk.Equals, "1024")
	c.Assert(appendResp.BlobCommittedBlockCount(), chk.Equals, int32(2))
}

func (s *aztestsSuite) TestAppendBlockWithMD5(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	// set up blob to test
	blob := container.NewAppendBlobURL(generateBlobName())
	resp, err := blob.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.StatusCode(), chk.Equals, 201)

	// test append block with valid MD5 value
	readerToBody, body := getRandomDataAndReader(1024)
	md5Value := md5.Sum(body)
	appendResp, err := blob.AppendBlock(context.Background(), readerToBody, AppendBlobAccessConditions{}, md5Value[:], ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(appendResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(appendResp.BlobAppendOffset(), chk.Equals, "0")
	c.Assert(appendResp.BlobCommittedBlockCount(), chk.Equals, int32(1))
	c.Assert(appendResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(appendResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(appendResp.ContentMD5(), chk.DeepEquals, md5Value[:])
	c.Assert(appendResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.Date().IsZero(), chk.Equals, false)

	// test append block with bad MD5 value
	readerToBody, body = getRandomDataAndReader(1024)
	_, badMD5 := getRandomDataAndReader(16)
	appendResp, err = blob.AppendBlock(context.Background(), readerToBody, AppendBlobAccessConditions{}, badMD5[:], ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeMd5Mismatch)
}

func (s *aztestsSuite) TestAppendBlockFromURL(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 4 * 1024 * 1024 // 4MB
	r, sourceData := getRandomDataAndReader(testSize)
	ctx := context.Background() // Use default Background context
	srcBlob := container.NewAppendBlobURL(generateName("appendsrc"))
	destBlob := container.NewAppendBlobURL(generateName("appenddest"))

	// Prepare source blob for copy.
	cResp1, err := srcBlob.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(cResp1.StatusCode(), chk.Equals, 201)
	appendResp, err := srcBlob.AppendBlock(context.Background(), r, AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(appendResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(appendResp.BlobAppendOffset(), chk.Equals, "0")
	c.Assert(appendResp.BlobCommittedBlockCount(), chk.Equals, int32(1))
	c.Assert(appendResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(appendResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(appendResp.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.Date().IsZero(), chk.Equals, false)

	// Get source blob URL with SAS for AppendBlockFromURL.
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

	// Append block from URL.
	cResp2, err := destBlob.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(cResp2.StatusCode(), chk.Equals, 201)
	appendFromURLResp, err := destBlob.AppendBlockFromURL(ctx, srcBlobURLWithSAS, 0, int64(testSize), AppendBlobAccessConditions{}, ModifiedAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(appendFromURLResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(appendFromURLResp.BlobAppendOffset(), chk.Equals, "0")
	c.Assert(appendFromURLResp.BlobCommittedBlockCount(), chk.Equals, int32(1))
	c.Assert(appendFromURLResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(appendFromURLResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(appendFromURLResp.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(appendFromURLResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(appendFromURLResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(appendFromURLResp.Date().IsZero(), chk.Equals, false)

	// Check data integrity through downloading.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)
}

func (s *aztestsSuite) TestAppendBlockFromURLWithMD5(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 4 * 1024 * 1024 // 4MB
	r, sourceData := getRandomDataAndReader(testSize)
	md5Value := md5.Sum(sourceData)
	ctx := context.Background() // Use default Background context
	srcBlob := container.NewAppendBlobURL(generateName("appendsrc"))
	destBlob := container.NewAppendBlobURL(generateName("appenddest"))

	// Prepare source blob for copy.
	cResp1, err := srcBlob.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(cResp1.StatusCode(), chk.Equals, 201)
	appendResp, err := srcBlob.AppendBlock(context.Background(), r, AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(appendResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(appendResp.BlobAppendOffset(), chk.Equals, "0")
	c.Assert(appendResp.BlobCommittedBlockCount(), chk.Equals, int32(1))
	c.Assert(appendResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(appendResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(appendResp.ContentMD5(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(appendResp.Date().IsZero(), chk.Equals, false)

	// Get source blob URL with SAS for AppendBlockFromURL.
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

	// Append block from URL.
	cResp2, err := destBlob.Create(context.Background(), BlobHTTPHeaders{}, nil, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(cResp2.StatusCode(), chk.Equals, 201)
	appendFromURLResp, err := destBlob.AppendBlockFromURL(ctx, srcBlobURLWithSAS, 0, int64(testSize), AppendBlobAccessConditions{}, ModifiedAccessConditions{}, md5Value[:], ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(appendFromURLResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(appendFromURLResp.BlobAppendOffset(), chk.Equals, "0")
	c.Assert(appendFromURLResp.BlobCommittedBlockCount(), chk.Equals, int32(1))
	c.Assert(appendFromURLResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(appendFromURLResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(appendFromURLResp.ContentMD5(), chk.DeepEquals, md5Value[:])
	c.Assert(appendFromURLResp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(appendFromURLResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(appendFromURLResp.Date().IsZero(), chk.Equals, false)

	// Check data integrity through downloading.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)

	// Test append block from URL with bad MD5 value
	_, badMD5 := getRandomDataAndReader(16)
	_, err = destBlob.AppendBlockFromURL(ctx, srcBlobURLWithSAS, 0, int64(testSize), AppendBlobAccessConditions{}, ModifiedAccessConditions{}, badMD5, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeMd5Mismatch)
}

func (s *aztestsSuite) TestBlobCreateAppendMetadataNonEmpty(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getAppendBlobURL(c, containerURL)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.NewMetadata(), chk.DeepEquals, basicMetadata)
}

func (s *aztestsSuite) TestBlobCreateAppendMetadataEmpty(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getAppendBlobURL(c, containerURL)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.NewMetadata(), chk.HasLen, 0)
}

func (s *aztestsSuite) TestBlobCreateAppendMetadataInvalid(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getAppendBlobURL(c, containerURL)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, Metadata{"In valid!": "bar"}, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(strings.Contains(err.Error(), invalidHeaderErrorSubstring), chk.Equals, true)
}

func (s *aztestsSuite) TestBlobCreateAppendHTTPHeaders(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getAppendBlobURL(c, containerURL)

	_, err := blobURL.Create(ctx, basicHeaders, nil, BlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	h := resp.NewHTTPHeaders()
	c.Assert(h, chk.DeepEquals, basicHeaders)
}

func validateAppendBlobPut(c *chk.C, blobURL AppendBlobURL) {
	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.NewMetadata(), chk.DeepEquals, basicMetadata)
}

func (s *aztestsSuite) TestBlobCreateAppendIfModifiedSinceTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(-10)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfModifiedSince: currentTime}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateAppendBlobPut(c, blobURL)
}

func (s *aztestsSuite) TestBlobCreateAppendIfModifiedSinceFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(10)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfModifiedSince: currentTime}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobCreateAppendIfUnmodifiedSinceTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(10)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfUnmodifiedSince: currentTime}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateAppendBlobPut(c, blobURL)
}

func (s *aztestsSuite) TestBlobCreateAppendIfUnmodifiedSinceFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(-10)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfUnmodifiedSince: currentTime}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobCreateAppendIfMatchTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	resp, _ := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: resp.ETag()}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateAppendBlobPut(c, blobURL)
}

func (s *aztestsSuite) TestBlobCreateAppendIfMatchFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: ETag("garbage")}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobCreateAppendIfNoneMatchTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfNoneMatch: ETag("garbage")}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateAppendBlobPut(c, blobURL)
}

func (s *aztestsSuite) TestBlobCreateAppendIfNoneMatchFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	resp, _ := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})

	_, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfNoneMatch: resp.ETag()}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobAppendBlockNilBody(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, bytes.NewReader(nil), AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)
	validateStorageError(c, err, ServiceCodeInvalidHeaderValue)
}

func (s *aztestsSuite) TestBlobAppendBlockEmptyBody(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(""), AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeInvalidHeaderValue)
}

func (s *aztestsSuite) TestBlobAppendBlockNonExistantBlob(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getAppendBlobURL(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeBlobNotFound)
}

func validateBlockAppended(c *chk.C, blobURL AppendBlobURL, expectedSize int) {
	resp, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.ContentLength(), chk.Equals, int64(expectedSize))
}

func (s *aztestsSuite) TestBlobAppendBlockIfModifiedSinceTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(-10)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfModifiedSince: currentTime}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlockAppended(c, blobURL, len(blockBlobDefaultData))
}

func (s *aztestsSuite) TestBlobAppendBlockIfModifiedSinceFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(10)
	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfModifiedSince: currentTime}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobAppendBlockIfUnmodifiedSinceTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(10)
	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfUnmodifiedSince: currentTime}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlockAppended(c, blobURL, len(blockBlobDefaultData))
}

func (s *aztestsSuite) TestBlobAppendBlockIfUnmodifiedSinceFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	currentTime := getRelativeTimeGMT(-10)
	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfUnmodifiedSince: currentTime}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobAppendBlockIfMatchTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	resp, _ := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: resp.ETag()}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlockAppended(c, blobURL, len(blockBlobDefaultData))
}

func (s *aztestsSuite) TestBlobAppendBlockIfMatchFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: ETag("garbage")}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobAppendBlockIfNoneMatchTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfNoneMatch: ETag("garbage")}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlockAppended(c, blobURL, len(blockBlobDefaultData))
}

func (s *aztestsSuite) TestBlobAppendBlockIfNoneMatchFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	resp, _ := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfNoneMatch: resp.ETag()}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeConditionNotMet)
}

func (s *aztestsSuite) TestBlobAppendBlockIfAppendPositionMatchTrueNegOne(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{AppendPositionAccessConditions: AppendPositionAccessConditions{IfAppendPositionEqual: -1}}, nil, ClientProvidedKeyOptions{}) // This will cause the library to set the value of the header to 0
	c.Assert(err, chk.IsNil)

	validateBlockAppended(c, blobURL, len(blockBlobDefaultData))
}

func (s *aztestsSuite) TestBlobAppendBlockIfAppendPositionMatchZero(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{}) // The position will not match, but the condition should be ignored
	c.Assert(err, chk.IsNil)
	_, err = blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{AppendPositionAccessConditions: AppendPositionAccessConditions{IfAppendPositionEqual: 0}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlockAppended(c, blobURL, 2*len(blockBlobDefaultData))
}

func (s *aztestsSuite) TestBlobAppendBlockIfAppendPositionMatchTrueNonZero(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{AppendPositionAccessConditions: AppendPositionAccessConditions{IfAppendPositionEqual: int64(len(blockBlobDefaultData))}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlockAppended(c, blobURL, len(blockBlobDefaultData)*2)
}

func (s *aztestsSuite) TestBlobAppendBlockIfAppendPositionMatchFalseNegOne(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	_, err = blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{AppendPositionAccessConditions: AppendPositionAccessConditions{IfAppendPositionEqual: -1}}, nil, ClientProvidedKeyOptions{}) // This will cause the library to set the value of the header to 0
	validateStorageError(c, err, ServiceCodeAppendPositionConditionNotMet)
}

func (s *aztestsSuite) TestBlobAppendBlockIfAppendPositionMatchFalseNonZero(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{AppendPositionAccessConditions: AppendPositionAccessConditions{IfAppendPositionEqual: 12}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeAppendPositionConditionNotMet)
}

func (s *aztestsSuite) TestBlobAppendBlockIfMaxSizeTrue(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{AppendPositionAccessConditions: AppendPositionAccessConditions{IfMaxSizeLessThanOrEqual: int64(len(blockBlobDefaultData) + 1)}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	validateBlockAppended(c, blobURL, len(blockBlobDefaultData))
}

func (s *aztestsSuite) TestBlobAppendBlockIfMaxSizeFalse(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	_, err := blobURL.AppendBlock(ctx, strings.NewReader(blockBlobDefaultData), AppendBlobAccessConditions{AppendPositionAccessConditions: AppendPositionAccessConditions{IfMaxSizeLessThanOrEqual: int64(len(blockBlobDefaultData) - 1)}}, nil, ClientProvidedKeyOptions{})
	validateStorageError(c, err, ServiceCodeMaxBlobSizeConditionNotMet)
}
