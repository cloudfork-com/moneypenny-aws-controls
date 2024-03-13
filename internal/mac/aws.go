package mac

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go/aws"
)

func AllServices(client *ecs.Client) (list []types.Service, err error) {
	ctx := context.Background()

	var clusterToken *string
	for {
		slog.Info("collecting clusters", "profile", os.Getenv("AWS_PROFILE"))
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
				slog.Info("describing services", "cluster", eachCluster, "tasks.count", len(allServices.ServiceArns))
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
func AllTasks(client *ecs.Client) (list []types.Task, err error) {
	ctx := context.Background()

	var clusterToken *string
	for {
		slog.Info("collecting clusters", "profile", os.Getenv("AWS_PROFILE"))
		allClusters, err0 := client.ListClusters(ctx, &ecs.ListClustersInput{NextToken: clusterToken})
		if err0 != nil {
			return list, err0
		}
		for _, eachCluster := range allClusters.ClusterArns {
			slog.Info("collecting tasks", "cluster", eachCluster)
			var taskToken *string
			for {
				allTasks, err1 := client.ListTasks(context.TODO(), &ecs.ListTasksInput{
					Cluster:    aws.String(eachCluster),
					LaunchType: types.LaunchTypeFargate,
					NextToken:  taskToken,
				})
				if err1 != nil {
					return list, err1
				}
				slog.Info("describing tasks", "cluster", eachCluster, "tasks.count", len(allTasks.TaskArns))
				allInfos, err2 := client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{ // TODO paging
					Cluster: aws.String(eachCluster),
					Tasks:   allTasks.TaskArns,
					Include: []types.TaskField{types.TaskFieldTags},
				})
				if err2 != nil {
					return list, err2
				}
				list = append(list, allInfos.Tasks...)
				taskToken = allTasks.NextToken
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

func NameOfTask(task types.Task) string {
	for _, each := range task.Tags {
		if each.Key != nil && *each.Key == "name" {
			if each.Value != nil {
				return *each.Value
			}
		}
	}
	return ""
}
func NameOfService(service types.Service) string {
	for _, each := range service.Tags {
		if each.Key != nil && *each.Key == "name" {
			if each.Value != nil {
				return *each.Value
			}
		}
	}
	return ""
}

func TaskForService(client *ecs.Client, service types.Service) ([]types.Task, error) {
	slog.Info("collecting tasks", "service", service.ServiceName)
	ctx := context.Background()
	taskList, err := client.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:     service.ClusterArn,
		LaunchType:  types.LaunchTypeFargate,
		ServiceName: service.ServiceName,
	})
	if err != nil {
		return []types.Task{}, err
	}
	allInfos, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: service.ClusterArn,
		Tasks:   taskList.TaskArns,
	})
	if err != nil {
		return []types.Task{}, err
	}
	return allInfos.Tasks, nil
}
