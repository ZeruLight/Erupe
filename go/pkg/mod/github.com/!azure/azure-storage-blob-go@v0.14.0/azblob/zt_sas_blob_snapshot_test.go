package azblob

import (
	"bytes"
	"strings"
	"time"

	chk "gopkg.in/check.v1"
)

func (s *aztestsSuite) TestSnapshotSAS(c *chk.C) {
	//Generate URLs ----------------------------------------------------------------------------------------------------
	bsu := getBSU()
	containerURL, containerName := getContainerURL(c, bsu)
	blobURL, blobName := getBlockBlobURL(c, containerURL)

	_, err := containerURL.Create(ctx, Metadata{}, PublicAccessNone)
	defer containerURL.Delete(ctx, ContainerAccessConditions{})
	if err != nil {
		c.Fatal(err)
	}

	//Create file in container, download from snapshot to test. --------------------------------------------------------
	burl := containerURL.NewBlockBlobURL(blobName)
	data := "Hello world!"

	_, err = burl.Upload(ctx, strings.NewReader(data), BlobHTTPHeaders{ContentType: "text/plain"}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	if err != nil {
		c.Fatal(err)
	}

	//Create a snapshot & URL
	createSnapshot, err := burl.CreateSnapshot(ctx, Metadata{}, BlobAccessConditions{}, ClientProvidedKeyOptions{})
	if err != nil {
		c.Fatal(err)
	}

	//Format snapshot time
	snapTime, err := time.Parse(SnapshotTimeFormat, createSnapshot.Snapshot())
	if err != nil {
		c.Fatal(err)
	}

	//Get credentials & current time
	currentTime := time.Now().UTC()
	credential, err := getGenericCredential("")
	if err != nil {
		c.Fatal("Invalid credential")
	}

	//Create SAS query
	snapSASQueryParams, err := BlobSASSignatureValues{
		StartTime:     currentTime,
		ExpiryTime:    currentTime.Add(48 * time.Hour),
		SnapshotTime:  snapTime,
		Permissions:   "racwd",
		ContainerName: containerName,
		BlobName:      blobName,
		Protocol:      SASProtocolHTTPS,
	}.NewSASQueryParameters(credential)
	if err != nil {
		c.Fatal(err)
	}
	time.Sleep(time.Second * 2)

	//Attach SAS query to block blob URL
	p := NewPipeline(NewAnonymousCredential(), PipelineOptions{})
	snapParts := NewBlobURLParts(blobURL.URL())
	snapParts.SAS = snapSASQueryParams
	sburl := NewBlockBlobURL(snapParts.URL(), p)

	//Test the snapshot
	downloadResponse, err := sburl.Download(ctx, 0, 0, BlobAccessConditions{}, false, ClientProvidedKeyOptions{})
	if err != nil {
		c.Fatal(err)
	}

	downloadedData := &bytes.Buffer{}
	reader := downloadResponse.Body(RetryReaderOptions{})
	downloadedData.ReadFrom(reader)
	reader.Close()

	c.Assert(data, chk.Equals, downloadedData.String())

	//Try to delete snapshot -------------------------------------------------------------------------------------------
	_, err = sburl.Delete(ctx, DeleteSnapshotsOptionNone, BlobAccessConditions{})
	if err != nil { //This shouldn't fail.
		c.Fatal(err)
	}

	//Create a normal blob and attempt to use the snapshot SAS against it (assuming failure) ---------------------------
	//If this succeeds, it means a normal SAS token was created.

	fsburl := containerURL.NewBlockBlobURL("failsnap")
	_, err = fsburl.Upload(ctx, strings.NewReader(data), BlobHTTPHeaders{ContentType: "text/plain"}, Metadata{}, BlobAccessConditions{}, DefaultAccessTier, nil, ClientProvidedKeyOptions{})
	if err != nil {
		c.Fatal(err) //should succeed to create the blob via normal auth means
	}

	fsburlparts := NewBlobURLParts(fsburl.URL())
	fsburlparts.SAS = snapSASQueryParams
	fsburl = NewBlockBlobURL(fsburlparts.URL(), p) //re-use fsburl as we don't need the sharedkey version anymore

	resp, err := fsburl.Delete(ctx, DeleteSnapshotsOptionNone, BlobAccessConditions{})
	if err == nil {
		c.Fatal(resp) //This SHOULD fail. Otherwise we have a normal SAS token...
	}
}
