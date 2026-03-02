package dns

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

var (
	errNoWeightedRecords = errors.New("No weighted records found for this name. Nothing to delete.")
)

func upsertDNSRecord(client *route53.Client, record, publicIP, comment string) (*route53.ChangeResourceRecordSetsOutput, error) {
	change := &types.ChangeBatch{
		Changes: []types.Change{
			{
				Action: types.ChangeActionUpsert,
				ResourceRecordSet: &types.ResourceRecordSet{
					Name: aws.String(record),
					ResourceRecords: []types.ResourceRecord{
						{
							Value: aws.String(publicIP),
						},
					},
					TTL:  aws.Int64(60),
					Type: types.RRTypeA,
				},
			},
		},
	}

	if comment != "" {
		change.Comment = aws.String(comment)
	}

	return client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(os.Getenv("AWS_PRIVATE_HOSTED_ZONE_ID")),
		ChangeBatch:  change,
	})
}

type DNSWeightedRecord struct {
	Name          string
	SetIdentifier string
	Weight        int64
	IP            string
}

func upsertDNSWeightedRecord(client *route53.Client, record DNSWeightedRecord, comment string) (*route53.ChangeResourceRecordSetsOutput, error) {
	change := &types.ChangeBatch{
		Changes: []types.Change{
			{
				Action: types.ChangeActionUpsert,
				ResourceRecordSet: &types.ResourceRecordSet{
					Name:          aws.String(record.Name),
					Type:          types.RRTypeA,
					SetIdentifier: aws.String(record.SetIdentifier),
					TTL:           aws.Int64(60),
					Weight:        aws.Int64(record.Weight),
					ResourceRecords: []types.ResourceRecord{
						{
							Value: aws.String(record.IP),
						},
					},
				},
			},
		},
	}

	if comment != "" {
		change.Comment = aws.String(comment)
	}

	return client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(os.Getenv("AWS_PRIVATE_HOSTED_ZONE_ID")),
		ChangeBatch:  change,
	})
}

func resolveDNSRecord(record string) (string, error) {
	ips, err := net.LookupIP(record)
	if err != nil {
		return "", err
	}

	if len(ips) > 1 {
		log.Println("Notice! mutiple ips resolved for record - " + record)
	}

	ip := ips[0]

	return ip.String(), nil
}

func getPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func deleteExistingWeightedDNSRecords(client *route53.Client, record string) (*route53.ChangeResourceRecordSetsOutput, error) {
	// Fetch existing records starting at our target name
	listInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(os.Getenv("AWS_PRIVATE_HOSTED_ZONE_ID")),
		StartRecordName: aws.String(record),
	}

	listResp, err := client.ListResourceRecordSets(context.TODO(), listInput)
	if err != nil {
		e := errors.New("failed to list resource record sets")
		return nil, errors.Join(e, err)
	}

	// Filter the results and build the DELETE batch
	var changes []types.Change

	for _, rs := range listResp.ResourceRecordSets {
		// Because ListResourceRecordSets returns everything from the starting point alphabetically,
		// we break the loop as soon as we hit a record name that doesn't match our target.
		resultName := strings.TrimRight(*rs.Name, ".") // We need relative DNS names
		if resultName != record {
			break
		}

		// Check if it is a weighted record by verifying the Weight pointer is not nil
		if rs.Weight == nil {
			continue
		}

		// Append a DELETE action using the EXACT record set we just fetched
		changes = append(changes, types.Change{
			Action: types.ChangeActionDelete,
			// By passing the fetched 'rs' directly, we guarantee the TTL, Value,
			// Weight, and SetIdentifier match exactly what Route 53 expects.
			ResourceRecordSet: &rs,
		})
	}

	// If no weighted records were found, exit early
	if len(changes) == 0 {
		return nil, errNoWeightedRecords
	}

	// STEP 3: Execute the Delete operation
	deleteInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(os.Getenv("AWS_PRIVATE_HOSTED_ZONE_ID")),
		ChangeBatch: &types.ChangeBatch{
			Comment: aws.String("Deleting all weighted records for " + record),
			Changes: changes,
		},
	}

	deleteResp, err := client.ChangeResourceRecordSets(context.TODO(), deleteInput)
	if err != nil {
		e := errors.New("failed to delete resource record sets")
		return nil, errors.Join(e, err)
	}

	return deleteResp, nil
}
