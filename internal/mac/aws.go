package mac

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go/aws"
)

func AllServices(client *ecs.Client) (list []types.Service, err error) {
	ctx := context.Background()

	var clusterToken *string
	for {
		slog.Info("collecting clusters")
		allClusters, err0 := client.ListClusters(ctx, &ecs.ListClustersInput{NextToken: clusterToken})
		if err0 != nil {
			return list, err0
		}
		for _, eachCluster := range allClusters.ClusterArns {
			slog.Info("collecting services", "cluster", eachCluster)
			var taskToken *string
			for {
				allServices, err1 := client.ListServices(context.TODO(), &ecs.ListServicesInput{
					Cluster:    aws.String(eachCluster),
					LaunchType: types.LaunchTypeFargate,
					NextToken:  taskToken,
					MaxResults: aws.Int32(10), // max for describe
				})
				if err1 != nil {
					return list, err1
				}
				if len(allServices.ServiceArns) == 0 { // InvalidParameterException: Services cannot be empty
					return list, nil
				}
				slog.Debug("describing services", "cluster", eachCluster, "services.count", len(allServices.ServiceArns))
				allInfos, err2 := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{ // TODO paging
					Cluster:  aws.String(eachCluster),
					Services: allServices.ServiceArns,
					Include:  []types.ServiceField{types.ServiceFieldTags},
				})
				if err2 != nil {
					return list, err2
				}
				list = append(list, allInfos.Services...)
				taskToken = allServices.NextToken
				if taskToken == nil {
					break
				}
			}
		}
		clusterToken = allClusters.NextToken
		if clusterToken == nil {
			break
		}
	}
	return
}

func TagValue(service types.Service, tagKey string) string {
	for _, each := range service.Tags {
		if each.Key != nil && *each.Key == tagKey {
			if each.Value != nil {
				return *each.Value
			}
		}
	}
	return ""
}

func TasksForService(client *ecs.Client, clusterARN, shortServiceName string) ([]types.Task, error) {
	slog.Info("collecting tasks", "name", shortServiceName)
	ctx := context.Background()
	taskList, err := client.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:     aws.String(clusterARN),
		LaunchType:  types.LaunchTypeFargate,
		ServiceName: aws.String(shortServiceName),
	})
	if err != nil {
		return []types.Task{}, err
	}
	allInfos, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterARN),
		Tasks:   taskList.TaskArns,
	})
	if err != nil {
		return []types.Task{}, err
	}
	return allInfos.Tasks, nil
}

func StartService(client *ecs.Client, service Service, desiredTaskCount int) error {
	slog.Info("starting service", "arn", service.ARN)
	count := int32(desiredTaskCount)
	if count == 0 { //unspecified
		count = 1
	}
	return ChangeTaskCountOfService(client, service, int(count))
}

func ChangeTaskCountOfService(client *ecs.Client, service Service, desiredTaskCount int) error {
	slog.Info("changing tasks count of service", "arn", service.ARN, "count", desiredTaskCount)
	count := int32(desiredTaskCount)
	_, err := client.UpdateService(context.Background(), &ecs.UpdateServiceInput{
		Service:      aws.String(service.Name()),
		DesiredCount: aws.Int32(count),
		Cluster:      aws.String(service.ClusterARN()),
	})
	return err
}

func StopService(client *ecs.Client, service Service) error {
	slog.Info("stopping service", "arn", service.ARN)
	return ChangeTaskCountOfService(client, service, 0)
}

func ServiceStatus(client *ecs.Client, service Service) (int, string) {
	// at least one task must be running
	tasks, _ := TasksForService(client, service.ClusterARN(), service.Name())
	for _, each := range tasks {
		if each.LastStatus != nil {
			return len(tasks), *each.LastStatus
		}
	}
	return 0, Unknown
}
