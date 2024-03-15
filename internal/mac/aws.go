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
	slog.Info("collecting tasks", "service", *service.ServiceName)
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

func TasksForService(client *ecs.Client, clusterARN, shortServiceName string) ([]types.Task, error) {
	slog.Info("collecting tasks", "service", shortServiceName)
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

func StartTask(client *ecs.Client, service Service, task types.Task) error {
	slog.Info("starting task", "arn", *task.TaskDefinitionArn)
	_, err := client.StartTask(context.Background(), &ecs.StartTaskInput{
		ContainerInstances: []string{*task.ContainerInstanceArn},
		TaskDefinition:     task.TaskDefinitionArn,
		Cluster:            task.ClusterArn,
	})
	return err
}

func StopTask(client *ecs.Client, task types.Task) error {
	slog.Info("stopping task", "arn", *task.TaskArn)
	_, err := client.StopTask(context.Background(), &ecs.StopTaskInput{
		Task:    task.TaskArn,
		Cluster: task.ClusterArn,
		Reason:  aws.String("moneypenny-aws-controls"),
	})
	return err
}

func StartService(client *ecs.Client, service Service) error {
	slog.Info("starting service", "arn", service.ARN)
	count := int32(service.DesiredTasksCount)
	if count == 0 { //unspecified
		count = 1
	}
	_, err := client.UpdateService(context.Background(), &ecs.UpdateServiceInput{
		Service:      aws.String(service.Name()),
		DesiredCount: aws.Int32(count),
		Cluster:      aws.String(service.ClusterARN()),
	})
	return err
}

func StopService(client *ecs.Client, service Service) error {
	slog.Info("stopping service", "arn", service.ARN)
	_, err := client.UpdateService(context.Background(), &ecs.UpdateServiceInput{
		Service:      aws.String(service.Name()),
		DesiredCount: aws.Int32(0),
		Cluster:      aws.String(service.ClusterARN()),
	})
	if err != nil {
		return err
	}
	all, err := TasksForService(client, service.ClusterARN(), service.Name())
	if err != nil {
		return err
	}
	for _, each := range all {
		if err := StopTask(client, each); err != nil {
			return err
		}
	}
	return nil
}

func IsServiceRunning(client *ecs.Client, service Service) bool {
	// at least one task must be running
	tasks, _ := TasksForService(client, service.ClusterARN(), service.Name())
	for _, each := range tasks {
		if each.LastStatus != nil && *each.LastStatus == Running {
			return true
		}
	}
	return false
}
