package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

func listInstances(ctx context.Context, project, zone string) error {
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	if zone != "" {
		req := &computepb.ListInstancesRequest{Project: project, Zone: zone}
		return printInstances(ctx, c, req)
	}

	zonesClient, err := compute.NewZonesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer zonesClient.Close()
	zr := &computepb.ListZonesRequest{Project: project}
	zit := zonesClient.List(ctx, zr)
	for {
		z, err := zit.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		req := &computepb.ListInstancesRequest{Project: project, Zone: z.GetName()}
		if err := printInstances(ctx, c, req); err != nil {
			return err
		}
	}

	return nil
}

func printInstances(ctx context.Context, c *compute.InstancesClient, req *computepb.ListInstancesRequest) error {
	it := c.List(ctx, req)
	fmt.Printf("ZONE: %s\n", req.GetZone())
	for {
		inst, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("%s\t%d\n", inst.GetName(), inst.GetId())
	}
	return nil
}

func listBuckets(ctx context.Context, project string) error {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	it := c.Buckets(ctx, project)
	fmt.Printf("BUCKETS:\n")
	for {
		b, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println(b.Name)
	}

	return nil
}

func main() {
	project := flag.String("project", "", "GCP project ID")
	zone := flag.String("zone", "", "GCP zone (optional)")
	buckets := flag.Bool("buckets", false, "List Cloud Storage buckets")
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "Error: project is required")
		os.Exit(1)
	}

	ctx := context.Background()
	if err := listInstances(ctx, *project, *zone); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *buckets {
		if err := listBuckets(ctx, *project); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
