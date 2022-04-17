package storage

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v61/common"
	"github.com/oracle/oci-go-sdk/v61/objectstorage"
)

type BucketList struct {
	Buckets []Bucket
}
type Bucket struct {
	BucketName      string
	CompartmentOCID string
	CreationDate    common.SDKTime
	Namespace       string
	Visibility      string
	ObjectCount     int
	size            int
}

func GetBucketsIncompartment(config common.ConfigurationProvider, compartmentOCID string) BucketList {

	var buckets BucketList
	var bucket Bucket

	objClient, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		panic(err)
	}
	request := objectstorage.GetNamespaceRequest{}
	ctx := context.Background()

	resp, err := objClient.GetNamespace(ctx, request)
	nameSpace := resp.Value

	req := objectstorage.ListBucketsRequest{NamespaceName: nameSpace,
		CompartmentId: &compartmentOCID}

	listBucketsResp, err := objClient.ListBuckets(ctx, req)

	for listBucketsResp.OpcNextPage != nil {
		req.Page = listBucketsResp.OpcNextPage
		resp1, _ := objClient.ListBuckets(context.Background(), req)
		listBucketsResp.Items = append(listBucketsResp.Items, resp1.Items...)
		fmt.Println("Processing next page")
	}

	for _, c := range listBucketsResp.Items {
		bucket.BucketName = *c.Name
		bucket.CompartmentOCID = *c.CompartmentId
		bucket.Namespace = *c.Namespace
		bucket.CreationDate = *c.TimeCreated

		bucketReq := objectstorage.GetBucketRequest{NamespaceName: nameSpace,
			BucketName: c.Name}

		bucketResp, err := objClient.GetBucket(ctx, bucketReq)

		if err != nil {
			panic(err)
		}

		if bucketResp.PublicAccessType == "NoPublicAccess" {
			bucket.Visibility = "Private"
		} else {
			bucket.Visibility = "Public"
		}

		bucket.ObjectCount = int(*bucketResp.ApproximateCount)
		sizeingb := int64(*bucketResp.ApproximateSize) / int64(1024) / int64(1024) / int64(1024)
		bucket.size = int(sizeingb)

		buckets.Buckets = append(buckets.Buckets, bucket)
	}
	return buckets
}
