package azblob

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	chk "gopkg.in/check.v1"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"
)

func (s *aztestsSuite) TestSetBlobTags(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)
	blobTagsMap := BlobTagsMap{
		"azure": "blob",
		"blob":  "sdk",
		"sdk":   "go",
	}
	blockBlobUploadResp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.StatusCode(), chk.Equals, 201)
	blobSetTagsResponse, err := blobURL.SetTags(ctx, nil, nil, nil, blobTagsMap)
	c.Assert(err, chk.IsNil)
	c.Assert(blobSetTagsResponse.StatusCode(), chk.Equals, 204)

	blobGetTagsResponse, err := blobURL.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResponse.StatusCode(), chk.Equals, 200)
	c.Assert(blobGetTagsResponse.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range blobGetTagsResponse.BlobTagSet {
		c.Assert(blobTagsMap[blobTag.Key], chk.Equals, blobTag.Value)
	}
}

func (s *aztestsSuite) TestSetBlobTagsWithVID(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)
	blobTagsMap := BlobTagsMap{
		"Go":         "CPlusPlus",
		"Python":     "CSharp",
		"Javascript": "Android",
	}
	blockBlobUploadResp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.StatusCode(), chk.Equals, 201)
	versionId1 := blockBlobUploadResp.VersionID()

	blockBlobUploadResp, err = blobURL.Upload(ctx, bytes.NewReader([]byte("updated_data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.StatusCode(), chk.Equals, 201)
	versionId2 := blockBlobUploadResp.VersionID()

	blobURL1 := blobURL.WithVersionID(versionId1)
	blobSetTagsResponse, err := blobURL1.SetTags(ctx, nil, nil, nil, blobTagsMap)
	c.Assert(err, chk.IsNil)
	c.Assert(blobSetTagsResponse.StatusCode(), chk.Equals, 204)

	blobGetTagsResponse, err := blobURL1.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResponse.StatusCode(), chk.Equals, 200)
	c.Assert(blobGetTagsResponse.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range blobGetTagsResponse.BlobTagSet {
		c.Assert(blobTagsMap[blobTag.Key], chk.Equals, blobTag.Value)
	}

	blobURL2 := blobURL.WithVersionID(versionId2)
	blobGetTagsResponse, err = blobURL2.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResponse.StatusCode(), chk.Equals, 200)
	c.Assert(blobGetTagsResponse.BlobTagSet, chk.IsNil)
}

func (s *aztestsSuite) TestSetBlobTagsWithVID2(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)

	blockBlobUploadResp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.StatusCode(), chk.Equals, 201)
	versionId1 := blockBlobUploadResp.VersionID()

	blockBlobUploadResp, err = blobURL.Upload(ctx, bytes.NewReader([]byte("updated_data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.StatusCode(), chk.Equals, 201)
	versionId2 := blockBlobUploadResp.VersionID()

	blobTags1 := BlobTagsMap{
		"Go":         "CPlusPlus",
		"Python":     "CSharp",
		"Javascript": "Android",
	}

	blobURL1 := blobURL.WithVersionID(versionId1)
	blobSetTagsResponse, err := blobURL1.SetTags(ctx, nil, nil, nil, blobTags1)
	c.Assert(err, chk.IsNil)
	c.Assert(blobSetTagsResponse.StatusCode(), chk.Equals, 204)

	blobGetTagsResponse, err := blobURL1.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResponse.StatusCode(), chk.Equals, 200)
	c.Assert(blobGetTagsResponse.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range blobGetTagsResponse.BlobTagSet {
		c.Assert(blobTags1[blobTag.Key], chk.Equals, blobTag.Value)
	}

	blobTags2 := BlobTagsMap{
		"a123": "321a",
		"b234": "432b",
	}

	blobURL2 := blobURL.WithVersionID(versionId2)
	blobSetTagsResponse, err = blobURL2.SetTags(ctx, nil, nil, nil, blobTags2)
	c.Assert(err, chk.IsNil)
	c.Assert(blobSetTagsResponse.StatusCode(), chk.Equals, 204)

	blobGetTagsResponse, err = blobURL2.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResponse.StatusCode(), chk.Equals, 200)
	c.Assert(blobGetTagsResponse.BlobTagSet, chk.NotNil)
	for _, blobTag := range blobGetTagsResponse.BlobTagSet {
		c.Assert(blobTags2[blobTag.Key], chk.Equals, blobTag.Value)
	}
}

func (s *aztestsSuite) TestUploadBlockBlobWithSpecialCharactersInTags(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)
	blobTagsMap := BlobTagsMap{
		"+-./:=_ ": "firsttag",
		"tag2":     "+-./:=_",
		"+-./:=_1": "+-./:=_",
	}
	blockBlobUploadResp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.StatusCode(), chk.Equals, 201)

	blobGetTagsResponse, err := blobURL.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResponse.StatusCode(), chk.Equals, 200)
	c.Assert(blobGetTagsResponse.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range blobGetTagsResponse.BlobTagSet {
		c.Assert(blobTagsMap[blobTag.Key], chk.Equals, blobTag.Value)
	}
}

func (s *aztestsSuite) TestStageBlockWithTags(c *chk.C) {
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

	blobTagsMap := BlobTagsMap{
		"azure": "blob",
		"blob":  "sdk",
		"sdk":   "go",
	}
	commitResp, err := blobURL.CommitBlockList(ctx, base64BlockIDs, BlobHTTPHeaders{}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(commitResp.VersionID(), chk.NotNil)
	versionId := commitResp.VersionID()

	contentResp, err := blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	contentData, err := ioutil.ReadAll(contentResp.Body(RetryReaderOptions{}))
	c.Assert(contentData, chk.DeepEquals, []uint8(strings.Join(data, "")))

	blobURL1 := blobURL.WithVersionID(versionId)
	blobGetTagsResp, err := blobURL1.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResp, chk.NotNil)
	c.Assert(blobGetTagsResp.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range blobGetTagsResp.BlobTagSet {
		c.Assert(blobTagsMap[blobTag.Key], chk.Equals, blobTag.Value)
	}

	blobGetTagsResp, err = blobURL.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResp, chk.NotNil)
	c.Assert(blobGetTagsResp.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range blobGetTagsResp.BlobTagSet {
		c.Assert(blobTagsMap[blobTag.Key], chk.Equals, blobTag.Value)
	}
}

func (s *aztestsSuite) TestStageBlockFromURLWithTags(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 8 * 1024 * 1024 // 8MB
	r, sourceData := getRandomDataAndReader(testSize)
	ctx := ctx // Use default Background context
	srcBlob := container.NewBlockBlobURL("sourceBlob")
	destBlob := container.NewBlockBlobURL("destBlob")

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

	blockList, err := destBlob.GetBlockList(ctx, BlockListAll, LeaseAccessConditions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockList.Response().StatusCode, chk.Equals, 200)
	c.Assert(blockList.CommittedBlocks, chk.HasLen, 0)
	c.Assert(blockList.UncommittedBlocks, chk.HasLen, 2)

	listResp, err := destBlob.CommitBlockList(ctx, []string{blockID1, blockID2}, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(listResp.Response().StatusCode, chk.Equals, 201)
	//versionId := listResp.VersionID()

	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)

	blobGetTagsResp, err := destBlob.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResp.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range blobGetTagsResp.BlobTagSet {
		c.Assert(blobTagsMap[blobTag.Key], chk.Equals, blobTag.Value)
	}
}

func (s *aztestsSuite) TestCopyBlockBlobFromURLWithTags(c *chk.C) {
	bsu := getBSU()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	testSize := 1 * 1024 * 1024 // 1MB
	r, sourceData := getRandomDataAndReader(testSize)
	sourceDataMD5Value := md5.Sum(sourceData)
	srcBlob := container.NewBlockBlobURL("srcBlob")
	destBlob := container.NewBlockBlobURL("destBlob")

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

	downloadResp, err := destBlob.BlobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	destData, err := ioutil.ReadAll(downloadResp.Body(RetryReaderOptions{}))
	c.Assert(err, chk.IsNil)
	c.Assert(destData, chk.DeepEquals, sourceData)

	c.Assert(len(downloadResp.NewMetadata()), chk.Equals, 1)

	_, badMD5 := getRandomDataAndReader(16)
	_, err = destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, badMD5, DefaultAccessTier, blobTagsMap)
	c.Assert(err, chk.NotNil)

	resp, err = destBlob.CopyFromURL(ctx, srcBlobURLWithSAS, Metadata{}, ModifiedAccessConditions{}, BlobAccessConditions{}, nil, DefaultAccessTier, blobTagsMap)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Response().StatusCode, chk.Equals, 202)
	c.Assert(resp.XMsContentCrc64(), chk.Not(chk.Equals), "")
}

func (s *aztestsSuite) TestGetPropertiesReturnsTagsCount(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := getBlockBlobURL(c, containerURL)
	blobTagsMap := BlobTagsMap{
		"azure": "blob",
		"blob":  "sdk",
		"sdk":   "go",
	}
	blockBlobUploadResp, err := blobURL.Upload(ctx, bytes.NewReader([]byte("data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(blockBlobUploadResp.StatusCode(), chk.Equals, 201)

	getPropertiesResponse, err := blobURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(getPropertiesResponse.TagCount(), chk.Equals, int64(3))

	downloadResp, err := blobURL.Download(ctx, 0, CountToEnd, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(downloadResp, chk.NotNil)
	c.Assert(downloadResp.r.rawResponse.Header.Get("x-ms-tag-count"), chk.Equals, "3")
}

func (s *aztestsSuite) TestSetBlobTagForSnapshot(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, _ := createNewBlockBlob(c, containerURL)
	blobTagsMap := BlobTagsMap{
		"Microsoft Azure": "Azure Storage",
		"Storage+SDK":     "SDK/GO",
		"GO ":             ".Net",
	}
	_, err := blobURL.SetTags(ctx, nil, nil, nil, blobTagsMap)
	c.Assert(err, chk.IsNil)

	resp, err := blobURL.CreateSnapshot(ctx, nil, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	snapshotURL := blobURL.WithSnapshot(resp.Snapshot())
	resp2, err := snapshotURL.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp2.TagCount(), chk.Equals, int64(3))
}

func (s *aztestsSuite) TestCreatePageBlobWithTags(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	blobTagsMap := BlobTagsMap{
		"azure": "blob",
		"blob":  "sdk",
		"sdk":   "go",
	}
	blob, _ := createNewPageBlob(c, container)
	putResp, err := blob.UploadPages(ctx, 0, getReaderToRandomBytes(1024), PageBlobAccessConditions{}, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(putResp.Response().StatusCode, chk.Equals, 201)
	c.Assert(putResp.LastModified().IsZero(), chk.Equals, false)
	c.Assert(putResp.ETag(), chk.Not(chk.Equals), ETagNone)
	c.Assert(putResp.Version(), chk.Not(chk.Equals), "")
	c.Assert(putResp.rawResponse.Header.Get("x-ms-version-id"), chk.NotNil)

	setTagResp, err := blob.SetTags(ctx, nil, nil, nil, blobTagsMap)
	c.Assert(err, chk.IsNil)
	c.Assert(setTagResp.StatusCode(), chk.Equals, 204)

	gpResp, err := blob.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(gpResp, chk.NotNil)
	c.Assert(gpResp.rawResponse.Header.Get("x-ms-tag-count"), chk.Equals, "3")

	modifiedBlobTags := BlobTagsMap{
		"a0z1u2r3e4": "b0l1o2b3",
		"b0l1o2b3":   "s0d1k2",
	}

	setTagResp, err = blob.SetTags(ctx, nil, nil, nil, modifiedBlobTags)
	c.Assert(err, chk.IsNil)
	c.Assert(setTagResp.StatusCode(), chk.Equals, 204)

	gpResp, err = blob.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(gpResp, chk.NotNil)
	c.Assert(gpResp.rawResponse.Header.Get("x-ms-tag-count"), chk.Equals, "2")
}

func (s *aztestsSuite) TestSetTagOnPageBlob(c *chk.C) {
	bsu := getBSU()
	container, _ := createNewContainer(c, bsu)
	defer delContainer(c, container)

	blob, _ := getPageBlobURL(c, container)
	blobTagsMap := BlobTagsMap{
		"azure": "blob",
		"blob":  "sdk",
		"sdk":   "go",
	}
	resp, err := blob.Create(ctx, PageBlobPageBytes*10, 0, BlobHTTPHeaders{}, nil, BlobAccessConditions{}, DefaultPremiumBlobAccessTier, blobTagsMap, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.StatusCode(), chk.Equals, 201)

	gpResp, err := blob.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(gpResp, chk.NotNil)
	c.Assert(gpResp.rawResponse.Header.Get("x-ms-tag-count"), chk.Equals, "3")

	modifiedBlobTags := BlobTagsMap{
		"a0z1u2r3e4": "b0l1o2b3",
		"b0l1o2b3":   "s0d1k2",
	}

	setTagResp, err := blob.SetTags(ctx, nil, nil, nil, modifiedBlobTags)
	c.Assert(err, chk.IsNil)
	c.Assert(setTagResp.StatusCode(), chk.Equals, 204)

	gpResp, err = blob.GetProperties(ctx, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	c.Assert(gpResp, chk.NotNil)
	c.Assert(gpResp.rawResponse.Header.Get("x-ms-tag-count"), chk.Equals, "2")
}

func (s *aztestsSuite) TestCreateAppendBlobWithTags(c *chk.C) {
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

func (s *aztestsSuite) TestListBlobReturnsTags(c *chk.C) {
	bsu := getBSU()
	containerURL, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL)
	blobURL, blobName := createNewBlockBlob(c, containerURL)
	blobTagsMap := BlobTagsMap{
		"+-./:=_ ": "firsttag",
		"tag2":     "+-./:=_",
		"+-./:=_1": "+-./:=_",
	}
	resp, err := blobURL.SetTags(ctx, nil, nil, nil, blobTagsMap)
	c.Assert(err, chk.IsNil)
	c.Assert(resp.StatusCode(), chk.Equals, 204)

	listBlobResp, err := containerURL.ListBlobsFlatSegment(ctx, Marker{}, ListBlobsSegmentOptions{Details: BlobListingDetails{Tags: true}})

	c.Assert(err, chk.IsNil)
	c.Assert(listBlobResp.Segment.BlobItems[0].Name, chk.Equals, blobName)
	c.Assert(listBlobResp.Segment.BlobItems[0].BlobTags.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range listBlobResp.Segment.BlobItems[0].BlobTags.BlobTagSet {
		c.Assert(blobTagsMap[blobTag.Key], chk.Equals, blobTag.Value)
	}
}

func (s *aztestsSuite) TestFindBlobsByTags(c *chk.C) {
	bsu := getBSU()
	containerURL1, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL1)
	containerURL2, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL2)
	containerURL3, _ := createNewContainer(c, bsu)
	defer deleteContainer(c, containerURL3)

	blobTagsMap1 := BlobTagsMap{
		"tag2": "tagsecond",
		"tag3": "tagthird",
	}
	blobTagsMap2 := BlobTagsMap{
		"tag1": "firsttag",
		"tag2": "secondtag",
		"tag3": "thirdtag",
	}
	blobURL11, _ := getBlockBlobURL(c, containerURL1)
	_, err := blobURL11.Upload(ctx, bytes.NewReader([]byte("random data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, blobTagsMap1, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	blobURL12, _ := getBlockBlobURL(c, containerURL1)
	_, err = blobURL12.Upload(ctx, bytes.NewReader([]byte("another random data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, blobTagsMap2, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	blobURL21, _ := getBlockBlobURL(c, containerURL2)
	_, err = blobURL21.Upload(ctx, bytes.NewReader([]byte("random data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)
	blobURL22, _ := getBlockBlobURL(c, containerURL2)
	_, err = blobURL22.Upload(ctx, bytes.NewReader([]byte("another random data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, blobTagsMap2, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	blobURL31, _ := getBlockBlobURL(c, containerURL3)
	_, err = blobURL31.Upload(ctx, bytes.NewReader([]byte("random data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	c.Assert(err, chk.IsNil)

	where := "\"tag4\"='fourthtag'"
	lResp, err := bsu.FindBlobsByTags(ctx, nil, nil, &where, Marker{}, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(lResp.Blobs, chk.HasLen, 0)

	//where = "\"tag1\"='firsttag'AND\"tag2\"='secondtag'AND\"@container\"='"+ containerName1 + "'"
	//TODO: Figure out how to do a composite query based on container.
	where = "\"tag1\"='firsttag'AND\"tag2\"='secondtag'"

	lResp, err = bsu.FindBlobsByTags(ctx, nil, nil, &where, Marker{}, nil)
	c.Assert(err, chk.IsNil)

	for _, blob := range lResp.Blobs {
		containsTag := false

		for _, tag := range blob.Tags.BlobTagSet {
			if tag.Value == "firsttag" {
				containsTag = true
			}
		}

		c.Assert(containsTag, chk.Equals, true)
	}
}

func (s *aztestsSuite) TestFilterBlobsUsingAccountSAS(c *chk.C) {
	accountName, accountKey := accountInfo()
	credential, err := NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		c.Fail()
	}

	sasQueryParams, err := AccountSASSignatureValues{
		Protocol:      SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(48 * time.Hour),
		Permissions:   AccountSASPermissions{Read: true, List: true, Write: true, DeletePreviousVersion: true, Tag: true, FilterByTags: true, Create: true}.String(),
		Services:      AccountSASServices{Blob: true}.String(),
		ResourceTypes: AccountSASResourceTypes{Service: true, Container: true, Object: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		log.Fatal(err)
	}

	qp := sasQueryParams.Encode()
	urlToSendToSomeone := fmt.Sprintf("https://%s.blob.core.windows.net?%s", accountName, qp)
	u, _ := url.Parse(urlToSendToSomeone)
	serviceURL := NewServiceURL(*u, NewPipeline(NewAnonymousCredential(), PipelineOptions{}))

	containerName := generateContainerName()
	containerURL := serviceURL.NewContainerURL(containerName)
	_, err = containerURL.Create(ctx, Metadata{}, PublicAccessNone)
	defer containerURL.Delete(ctx, ContainerAccessConditions{})
	if err != nil {
		c.Fatal(err)
	}

	blobURL := containerURL.NewBlockBlobURL("temp")
	_, err = blobURL.Upload(ctx, bytes.NewReader([]byte("random data")), BlobHTTPHeaders{}, basicMetadata, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	if err != nil {
		c.Fail()
	}

	blobTagsMap := BlobTagsMap{"tag1": "firsttag", "tag2": "secondtag", "tag3": "thirdtag"}
	setBlobTagsResp, err := blobURL.SetTags(ctx, nil, nil, nil, blobTagsMap)
	c.Assert(err, chk.IsNil)
	c.Assert(setBlobTagsResp.StatusCode(), chk.Equals, 204)

	blobGetTagsResp, err := blobURL.GetTags(ctx, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(blobGetTagsResp.StatusCode(), chk.Equals, 200)
	c.Assert(blobGetTagsResp.BlobTagSet, chk.HasLen, 3)
	for _, blobTag := range blobGetTagsResp.BlobTagSet {
		c.Assert(blobTagsMap[blobTag.Key], chk.Equals, blobTag.Value)
	}

	time.Sleep(30 * time.Second)
	where := "\"tag1\"='firsttag'AND\"tag2\"='secondtag'AND@container='" + containerName + "'"
	_, err = serviceURL.FindBlobsByTags(ctx, nil, nil, &where, Marker{}, nil)
	c.Assert(err, chk.IsNil)
}
