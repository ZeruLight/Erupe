package azblob

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"time"

	"crypto/md5"

	"bytes"
	"strings"

	chk "gopkg.in/check.v1" // go get gopkg.in/check.v1
)

func (s *aztestsSuite) TestGetBlobPropertiesUsingVID(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewAppendBlob(c, containerURL)

	blobProp, _ := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	createResp, err := blobURL.Create(ctx, BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{ModifiedAccessConditions: ModifiedAccessConditions{IfMatch: blobProp.ETag()}}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(createResp.VersionID(), chk.NotNil)
	blobProp, _ = blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(createResp.VersionID(), chk.Equals, blobProp.VersionID())
	c.Assert(createResp.LastModified(), chk.DeepEquals, blobProp.LastModified())
	c.Assert(createResp.ETag(), chk.Equals, blobProp.ETag())
	c.Assert(blobProp.IsCurrentVersion(), chk.Equals, "true")
}

func (s *aztestsSuite) TestSetBlobMetadataReturnsVID(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, blobName := createNewBlockBlob(c, containerURL)
	metadata := Metadata{"test_key_1": "test_value_1", "test_key_2": "2019"}
	resp, err := blobURL.SetMetadata(ctx, metadata, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.VersionID(), chk.NotNil)

	listBlobResp, err := containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{Details: BlobListingDetails{Metadata: true}})

	c.Assert(err, chk.IsNil)
	c.Assert(listBlobResp.Segment.BlobItems[0].Name, chk.Equals, blobName)
	c.Assert(listBlobResp.Segment.BlobItems[0].Metadata, chk.HasLen, 2)
	c.Assert(listBlobResp.Segment.BlobItems[0].Metadata, chk.DeepEquals, metadata)
}

func (s *aztestsSuite) TestCreateAndDownloadBlobSpecialCharactersWithVID(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	data := []rune("-._/()$=',~0123456789")
	for i := 0; i < len(data); i++ {
		blobName := "abc" + string(data[i])
		blobURL := containerURL.NewBlockBlobURL(blobName)
		resp, err := blobURL.Upload(ctx, strings.NewReader(string(data[i])), BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
		c.Assert(err, chk.IsNil)
		c.Assert(resp.VersionID(), chk.NotNil)

		dResp, err := blobURL.WithVersionID(resp.VersionID()).Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
		c.Assert(err, chk.IsNil)
		d1, err := ioutil.ReadAll(dResp.Body(RetryReaderOptions{}))
		c.Assert(dResp.Version(), chk.Not(chk.Equals), "")
		c.Assert(string(d1), chk.DeepEquals, string(data[i]))
		versionId := dResp.r.rawResponse.Header.Get("x-ms-version-id")
		c.Assert(versionId, chk.NotNil)
		c.Assert(versionId, chk.Equals, resp.VersionID())
	}
}

func (s *aztestsSuite) TestDeleteSpecificBlobVersion(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	blockBlobUploadResp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.VersionID(), chk.NotNil)
	versionID1 := blockBlobUploadResp.VersionID()

	blockBlobUploadResp, err = blobURL.Upload(ctx, bytes.NewReader([]byte("updated_data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.VersionID(), chk.NotNil)

	listBlobsResp, err := containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{Details: BlobListingDetails{Versions: true}})
	c.Assert(err, chk.IsNil)
	c.Assert(listBlobsResp.Segment.BlobItems, chk.HasLen, 2)

	// Deleting previous version snapshot.
	deleteResp, err := blobURL.WithVersionID(versionID1).Delete(ctx, DeleteSnapshotsOptionNone, BlobAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(deleteResp.StatusCode(), chk.Equals, 202)

	listBlobsResp, err = containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{Details: BlobListingDetails{Versions: true}})
	c.Assert(err, chk.IsNil)
	c.Assert(listBlobsResp.Segment.BlobItems, chk.NotNil)
	if len(listBlobsResp.Segment.BlobItems) != 1 {
		c.Fail()
	}
}

func (s *aztestsSuite) TestDeleteSpecificBlobVersionWithBlobSAS(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal(err)
	}
	containerURL, containerName := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, blobName := getBlockBlobURL(c, containerURL)

	resp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	versionId := resp.VersionID()
	c.Assert(versionId, chk.NotNil)

	resp, err = blobURL.Upload(ctx, bytes.NewReader([]byte("updated_data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.VersionID(), chk.NotNil)

	blobParts := NewBlobURLParts(blobURL.URL())
	blobParts.VersionID = versionId
	blobParts.SAS, err = BlobSASSignatureValues{
		Protocol:      SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(1 * time.Hour),
		ContainerName: containerName,
		BlobName:      blobName,
		Permissions:   BlobSASPermissions{Delete: true, DeletePreviousVersion: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}

	sbURL := NewBlockBlobURL(blobParts.URL(), containerURL.client.p)
	deleteResp, err := sbURL.Delete(ctx, DeleteSnapshotsOptionNone, BlobAccessConditions{})
	c.Assert(deleteResp, chk.IsNil)

	listBlobResp, err := containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{Details: BlobListingDetails{Versions: true}})
	c.Assert(err, chk.IsNil)
	for _, blob := range listBlobResp.Segment.BlobItems {
		c.Assert(blob.VersionID, chk.Not(chk.Equals), versionId)
	}
}

func (s *aztestsSuite) TestDownloadSpecificBlobVersion(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	blockBlobUploadResp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp, chk.NotNil)
	versionId1 := blockBlobUploadResp.VersionID()

	blockBlobUploadResp, err = blobURL.Upload(ctx, bytes.NewReader([]byte("updated_data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp, chk.NotNil)
	versionId2 := blockBlobUploadResp.VersionID()
	c.Assert(blockBlobUploadResp.VersionID(), chk.NotNil)

	// Download previous version of snapshot.
	blobURL = blobURL.WithVersionID(versionId1)
	blockBlobDeleteResp, err := blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	data, err := ioutil.ReadAll(blockBlobDeleteResp.Response().Body)
	c.Assert(string(data), chk.Equals, "data")

	// Download current version of snapshot.
	blobURL = blobURL.WithVersionID(versionId2)
	blockBlobDeleteResp, err = blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	data, err = ioutil.ReadAll(blockBlobDeleteResp.Response().Body)
	c.Assert(string(data), chk.Equals, "updated_data")
}

func (s *aztestsSuite) TestCreateBlobSnapshotReturnsVID(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer delContainer(c, containerURL)
	blobURL := containerURL.NewBlockBlobURL(generateBlobName())
	uploadResp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("updated_data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadResp.VersionID(), chk.NotNil)

	csResp, err := blobURL.CreateSnapshot(ctx, Metadata{}, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(csResp.VersionID(), chk.NotNil)
	lbResp, err := containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{
		Details: BlobListingDetails{Versions: true, Snapshots: true},
	})
	c.Assert(lbResp, chk.NotNil)
	if len(lbResp.Segment.BlobItems) < 2 {
		c.Fail()
	}

	_, err = blobURL.Delete(ctx, DeleteSnapshotsOptionInclude, BlobAccessConditions{})
	lbResp, err = containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{
		Details: BlobListingDetails{Versions: true, Snapshots: true},
	})
	c.Assert(lbResp, chk.NotNil)
	if len(lbResp.Segment.BlobItems) < 2 {
		c.Fail()
	}
	for _, blob := range lbResp.Segment.BlobItems {
		c.Assert(blob.Snapshot, chk.Equals, "")
	}
}

func (s *aztestsSuite) TestCopyBlobFromURLWithSASReturnsVID(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 4 * 1024 * 1024 // 4MB
	r, sourceData := getRandomDataAndReader(testSize)
	sourceDataMD5Value := md5.Sum(sourceData)
	ctx := context.Background()
	srcBlob := container.NewBlockBlobURL(generateBlobName())
	destBlob := container.NewBlockBlobURL(generateBlobName())

	uploadSrcResp, err := srcBlob.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadSrcResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(uploadSrcResp.Response().Header.Get("x-ms-version-id"), chk.NotNil)

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

	resp, err := destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{"foo": "bar"}, ModifiedAccessConditions{}, BlobAccessConditions{}, sourceDataMD5Value[:], DefaultAccessTier, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 202)
	c.Assert(resp.Version(), chk.Not(chk.Equals), "")
	c.Assert(resp.CopyID(), chk.Not(chk.Equals), "")
	c.Assert(string(resp.CopyStatus()), chk.DeepEquals, "success")
	c.Assert(resp.VersionID(), chk.NotNil)

	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)
	c.Assert(downloadResp.Response().Header.Get("x-ms-version-id"), chk.NotNil)
	c.Assert(len(downloadResp.NewMetadata()), chk.Equals, 1)
	_, badMD5 := getRandomDataAndReader(16)
	_, err = destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, badMD5, DefaultAccessTier, nil)
	c.Assert(err, chk.NotNil)

	resp, err = destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, nil, DefaultAccessTier, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 202)
	c.Assert(resp.XMsContentCrc64(), chk.Not(chk.Equals), "")
	c.Assert(resp.Response().Header.Get("x-ms-version"), chk.Equals, ServiceVersion)
	c.Assert(resp.Response().Header.Get("x-ms-version-id"), chk.NotNil)
}

func (s *aztestsSuite) TestCreateBlockBlobReturnsVID(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer delContainer(c, containerURL)

	testSize := 2 * 1024 * 1024 // 1MB
	r, _ := getRandomDataAndReader(testSize)
	ctx := context.Background() // Use default Background context
	blobURL := containerURL.NewBlockBlobURL(generateBlobName())

	// Prepare source blob for copy.
	uploadResp, err := blobURL.Upload(ctx, r, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(uploadResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(uploadResp.rawResponse.Header.Get("x-ms-version"), chk.Equals, ServiceVersion)
	c.Assert(uploadResp.Response().Header.Get("x-ms-version-id"), chk.NotNil)

	csResp, err := blobURL.CreateSnapshot(ctx, Metadata{}, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(csResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(csResp.Response().Header.Get("x-ms-version-id"), chk.NotNil)

	listBlobResp, err := containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{Details: BlobListingDetails{Snapshots: true}})
	c.Assert(err, chk.IsNil)
	c.Assert(listBlobResp.rawResponse.Header.Get("x-ms-request-id"), chk.NotNil)
	if len(listBlobResp.Segment.BlobItems) < 2 {
		c.Fail()
	}

	deleteResp, err := blobURL.Delete(ctx, DeleteSnapshotsOptionOnly, BlobAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(deleteResp.Response().StatusCode, chk.Equals, 202)
	c.Assert(deleteResp.Response().Header.Get("x-ms-version-id"), chk.NotNil)

	listBlobResp, err = containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{Details: BlobListingDetails{Snapshots: true, Versions: true}})
	c.Assert(err, chk.IsNil)
	c.Assert(listBlobResp.rawResponse.Header.Get("x-ms-request-id"), chk.NotNil)
	if len(listBlobResp.Segment.BlobItems) == 0 {
		c.Fail()
	}
	blobs := listBlobResp.Segment.BlobItems
	c.Assert(blobs[0].Snapshot, chk.Equals, "")
}

func (s *aztestsSuite) TestPutBlockListReturnsVID(c *chk.C) {
	blockIDIntToBase64 := func(blockID int) string {
		binaryBlockID := (&[4]byte{})[:]
		binary.LittleEndian.PutUint32(binaryBlockID, uint32(blockID))
		return base64.StdEncoding.EncodeToString(binaryBlockID)
	}
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer delContainer(c, containerURL)

	blobURL := containerURL.NewBlockBlobURL(generateBlobName())

	data := []string{"Azure ", "Storage ", "Block ", "Blob."}
	base64BlockIDs := make([]string, len(data))

	for index, d := range data {
		base64BlockIDs[index] = blockIDIntToBase64(index)
		resp, err := blobURL.StageBlock(ctx, base64BlockIDs[index], strings.NewReader(d), LeaseAccessConditions{}, nil, ClientProvidedKeyOptions{})
		if err != nil {
			c.Fail()
		}
		c.Assert(resp.Response().StatusCode, chk.Equals, 201)
		c.Assert(resp.Version(), chk.Not(chk.Equals), "")
	}

	commitResp, err := blobURL.CommitBlockList(ctx, base64BlockIDs, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(commitResp.VersionID(), chk.NotNil)

	contentResp, err := blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	contentData, err := ioutil.ReadAll(contentResp.Body(RetryReaderOptions{}))
	c.Assert(contentData, chk.DeepEquals, []uint8(strings.Join(data, "")))
}

func (s *aztestsSuite) TestSyncCopyBlobReturnsVID(c *chk.C) {

}

func (s *aztestsSuite) TestCreatePageBlobReturnsVID(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	blob, _ := createNewPageBlob(c, container)
	putResp, err := blob.UploadPages(context.Background(), 0, getReaderToRandomBytes(1024), PageBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(putResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(putResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(putResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(putResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(putResp.rawResponse.Header.Get("x-ms-version-id"), chk.NotNil)

	gpResp, err := blob.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(gpResp, chk.NotNil)
}
