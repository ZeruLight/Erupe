package azblob

import (
	"bytes"
	"crypto/md5"
	chk "gopkg.in/check.v1"
	"io/ioutil"
	"net/url"
	"time"
)

func CreateBlockBlobsForTesting(c *chk.C, size int) (ContainerURL, *SharedKeyCredential, *bytes.Reader, []uint8, [16]uint8, BlockBlobURL, BlockBlobURL) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)

	testSize := size * 1024 * 1024 // 1MB
	r, sourceData := getRandomDataAndReader(testSize)
	sourceDataMD5Value := md5.Sum(sourceData)
	srcBlob := container.NewBlockBlobURL(generateBlobName())
	destBlob := container.NewBlockBlobURL(generateBlobName())

	return container, credential, r, sourceData, sourceDataMD5Value, srcBlob, destBlob
}

func (s *aztestsSuite) TestPutBlobFromURLWithIncorrectURL(c *chk.C) {
	container, _, _, _, sourceDataMD5Value, _, destBlob := CreateBlockBlobsForTesting(c, 8)
	defer delContainer(c, container)

	// Invoke put blob from URL with URL without SAS and make sure it fails
	resp, err := destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, url.URL{}, basicMetadata, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], sourceDataMD5Value[:], DefaultAccessTier, BlobTagsMap{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)
	c.Assert(resp, chk.IsNil)
}

func (s *aztestsSuite) TestPutBlobFromURLWithMissingSAS(c *chk.C) {
	container, _, r, _, sourceDataMD5Value, srcBlob, destBlob := CreateBlockBlobsForTesting(c, 8)
	defer delContainer(c, container)

	// Prepare source blob for put.
	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	// Invoke put blob from URL with URL without SAS and make sure it fails
	resp, err := destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, srcBlob.URL(), basicMetadata, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], sourceDataMD5Value[:], DefaultAccessTier, BlobTagsMap{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)
	c.Assert(resp, chk.IsNil)
}

func (s *aztestsSuite) TestSetTierOnPutBlockBlobFromURL(c *chk.C) {
	container, credential, r, _, sourceDataMD5Value, srcBlob, _ := CreateBlockBlobsForTesting(c, 1)
	defer delContainer(c, container)

	// Setting blob tier as "cool"
	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, AccessTierCool, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)

	// Get source blob URL with SAS for StageFromURL.
	srcBlobParts := NewBlobURLParts(srcBlob.URL())

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
		resp, err := destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, srcBlobURLWithSAS, basicMetadata, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], sourceDataMD5Value[:], tier, BlobTagsMap{}, ClientProvidedKeyOptions{})
		c.Assert(err, chk.IsNil)
		c.Assert(resp.Response().StatusCode, chk.Equals, 201)

		destBlobPropResp, err := destBlob.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
		c.Assert(err, chk.IsNil)
		c.Assert(destBlobPropResp.AccessTier(), chk.Equals, string(tier))
		c.Assert(destBlobPropResp.NewMetadata(), chk.DeepEquals, basicMetadata)
	}
}

func (s *aztestsSuite) TestPutBlockBlobFromURL(c *chk.C) {
	container, credential, r, sourceData, sourceDataMD5Value, srcBlob, destBlob := CreateBlockBlobsForTesting(c, 8)
	defer delContainer(c, container)

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

	// Invoke put blob from URL.
	resp, err := destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, srcBlobURLWithSAS, basicMetadata, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], sourceDataMD5Value[:], DefaultAccessTier, BlobTagsMap{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 201)
	c.Assert(resp.ETag(), chk.Not(chk.Equals), "")
	c.Assert(resp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(resp.Version(), chk.Not(chk.Equals), "")
	c.Assert(resp.Date().IsZero(), chk.Equals, false)
	c.Assert(resp.ContentMD5(), chk.DeepEquals, sourceDataMD5Value[:])

	// Check data integrity through downloading.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)

	// Make sure the metadata got copied over
	c.Assert(len(downloadResp.NewMetadata()), chk.Equals, 1)
	c.Assert(downloadResp.NewMetadata(), chk.DeepEquals, basicMetadata)
}

func (s *aztestsSuite) TestPutBlobFromURLWithSASReturnsVID(c *chk.C) {
	container, credential, r, sourceData, sourceDataMD5Value, srcBlob, destBlob := CreateBlockBlobsForTesting(c, 4)
	defer delContainer(c, container)

	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(uploadSrcResp.Response().Header.Get("x-ms-version-id"), chk.NotNil)

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

	// Invoke put blob from URL
	resp, err := destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, srcBlobURLWithSAS, basicMetadata, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], sourceDataMD5Value[:], DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 201)
	c.Assert(resp.Version(), chk.Not(chk.Equals), "")
	c.Assert(resp.VersionID(), chk.NotNil)

	// Check data integrity through downloading.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)
	c.Assert(downloadResp.Response().Header.Get("x-ms-version-id"), chk.NotNil)
	c.Assert(len(downloadResp.NewMetadata()), chk.Equals, 1)
	c.Assert(downloadResp.NewMetadata(), chk.DeepEquals, basicMetadata)

	// Edge case: Not providing any source MD5 should see the CRC getting returned instead and service version matches
	resp, err = destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, nil, nil, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 201)
	c.Assert(resp.rawResponse.Header.Get("x-mx-content-crc64"), chk.NotNil)
	c.Assert(resp.Response().Header.Get("x-ms-version"), chk.Equals, ServiceVersion)
	c.Assert(resp.Response().Header.Get("x-ms-version-id"), chk.NotNil)
}

func (s *aztestsSuite) TestPutBlockBlobFromURLWithTags(c *chk.C) {
	container, credential, r, sourceData, sourceDataMD5Value, srcBlob, destBlob := CreateBlockBlobsForTesting(c, 1)
	defer delContainer(c, container)

	blobTagsMap := BlobTagsMap{
		"Go":         "CPlusPlus",
		"Python":     "CSharp",
		"Javascript": "Android",
	}

	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
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

	// Invoke put blob from URL
	resp, err := destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, srcBlobURLWithSAS, basicMetadata, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], sourceDataMD5Value[:], DefaultAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 201)
	c.Assert(resp.ETag(), chk.Not(chk.Equals), "")
	c.Assert(resp.RequestID(), chk.Not(chk.Equals), "")
	c.Assert(resp.Version(), chk.Not(chk.Equals), "")
	c.Assert(resp.Date().IsZero(), chk.Equals, false)
	c.Assert(resp.ContentMD5(), chk.DeepEquals, sourceDataMD5Value[:])

	// Check data integrity through downloading.
	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)
	c.Assert(len(downloadResp.NewMetadata()), chk.Equals, 1)
	c.Assert(downloadResp.r.rawResponse.Header.Get("x-ms-tag-count"), chk.Equals, "3")
	c.Assert(downloadResp.NewMetadata(), chk.DeepEquals, basicMetadata)

	// Edge case 1: Provide bad MD5 and make sure the put fails
	_, badMD5 := getRandomDataAndReader(16)
	_, err = destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, badMD5, badMD5, DefaultAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
	c.Assert(err, chk.NotNil)

	// Edge case 2: Not providing any source MD5 should see the CRC getting returned instead
	resp, err = destBlob.PutBlobFromURL(ctx, BlobHTTPHeaders{}, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, nil, nil, DefaultAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 201)
	c.Assert(resp.rawResponse.Header.Get("x-mx-content-crc64"), chk.NotNil)
}
