package mac

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go/aws"
)

type PlanFetcher struct {
	client *ecs.Client
	Plans  []*ServicePlan
}

func NewPlanFetcher(client *ecs.Client) *PlanFetcher {
	return &PlanFetcher{
		client: client,
	}
}

func (p *PlanFetcher) CheckServicePlans(plans []*ServicePlan) error {
	// given the service plans, check the AWS service, one-by-one because multi-cluster (optimization = group per cluster)
	for _, each := range plans {
		slog.Debug("describing service", "cluster", each.ClusterARN(), "service", each.ARN)
		infos, err := p.client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
			Cluster:  aws.String(each.ClusterARN()),
			Services: []string{each.ARN},
		})
		if err != nil || len(infos.Services) == 0 {
			slog.Warn("describe service fail or does not exist, plan will be disabled", "err", err)
			each.Disabled = true
		}
	}
	p.Plans = plans
	return nil
}

func (p *PlanFetcher) FetchServicePlans() error {
	allServices, err := AllServices(p.client)
	if err != nil {
		slog.Error("fetchServicesAndPlans fail", "err", err)
		return err
	}
	for _, each := range allServices {
		input := TagValue(each, serviceTagName)
		sp := new(ServicePlan)
		sp.ARN = *each.ServiceArn
		sp.TagValue = input // can be empty
		if IsTagValueReference(input) {
			slog.Debug("find tag value by service", "service", *each.ServiceArn, "moneypenny", input)
			input = ResolveTagValue(allServices, input)
			sp.ResolvedTagValue = input // can be empty
		}
		if input == "" {
			// skip this service plan
			continue
		}
		if err := sp.Validate(); err != nil {
			slog.Warn("invalid moneypenny tag value", "value", input, "err", err)
		}
		slog.Debug("adding service plan", "service", *each.ServiceArn, "crons", input)
		p.Plans = append(p.Plans, sp)
	}
	return nil
}
