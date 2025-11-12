package api

import (
	"context"
	"core/queue"
	"fmt"
)

type TaskServer struct {
	Queue *queue.Queue
	UnimplementedTaskServiceServer
}

func (s *TaskServer) Start(ctx context.Context, message *TaskStartRequest) (*TaskStartResponse, error) {
	fmt.Println("Task Start Request | ", message.Task.Metric, " | ", message.Task.Groups)

	task := s.Queue.AddTask(FromGRPCTaskStartInfo(message.Task))

	if task == nil {
		fmt.Println("Task Start Response | Error | Task = nil | ", message.Task.Metric, " | ", message.Task.Groups)
		return &TaskStartResponse{Task: &TaskInfo{}}, nil
	}

	fmt.Println("Task Start Response | Success | ", message.Task.Metric, " | ", message.Task.Groups)
	return &TaskStartResponse{Task: ToGRPCTaskInfo(task)}, nil
}

func (s *TaskServer) Stop(ctx context.Context, message *TaskStopRequest) (*TaskStopResponse, error) {
	fmt.Println("Task Stop Request ", message.Metric, " | ", message.Groups)

	task, err := s.Queue.DeleteTask(message.Metric, message.Groups, message.DeleteSnapshots)

	if err != nil {
		fmt.Println("Task Stop Response | Error | ", message.Metric, " | ", message.Groups, " | ", err)
		return &TaskStopResponse{Task: &TaskInfo{}}, nil
	}

	if task == nil {
		fmt.Println("Task Stop Response | Error | Task = nil | ", message.Metric, " | ", message.Groups)
		return &TaskStopResponse{Task: &TaskInfo{}}, nil
	}

	fmt.Println("Task Stop Response | Success | ", message.Metric, " | ", message.Groups)
	return &TaskStopResponse{Task: ToGRPCTaskInfo(task)}, nil
}

func (s *TaskServer) Update(ctx context.Context, message *TaskStartRequest) (*TaskStartResponse, error) {
	fmt.Println("Update Request | ", message.Task.Metric, " | ", message.Task.Groups)

	task, err := s.Queue.UpdateTask(FromGRPCTaskStartInfo(message.Task))

	if err != nil {
		fmt.Println("Update Response | Error | ", message.Task.Metric, " | ", message.Task.Groups, " | ", err)
		return &TaskStartResponse{Task: &TaskInfo{}}, err
	}

	fmt.Println("Update Response | Success | ", message.Task.Metric, " | ", message.Task.Groups)
	return &TaskStartResponse{Task: ToGRPCTaskInfo(task)}, nil
}

func (s *TaskServer) List(ctx context.Context, message *TaskListRequest) (*TaskListResponse, error) {
	fmt.Println("Task List Request | ", message.Groups)

	tasks := s.Queue.ListTasks(message.Groups)

	result := []*TaskInfo{}

	for i := 0; i < len(tasks); i++ {
		result = append(result, ToGRPCTaskInfo(tasks[i]))
	}

	fmt.Println("Task List Response | Success | ", result)
	return &TaskListResponse{Tasks: result}, nil
}

func (s *TaskServer) ForceExecute(ctx context.Context, message *TaskForceExecuteRequest) (*TaskForceExecuteResponse, error) {
	fmt.Println("Task ForceExecute Request | ", message.Task, " | ", message.Groups)

	task, err := s.Queue.ForceExecute(FromGRPCTaskStartInfo(message.Task), message.Groups)

	if err != nil {
		fmt.Println("Task ForceExecute Response | Error | ", message.Task, " | ", message.Groups, " | ", err)
		return &TaskForceExecuteResponse{Task: &TaskInfo{}}, err
	}

	return &TaskForceExecuteResponse{Task: ToGRPCTaskInfo(task)}, nil
}

func (s *TaskServer) UpdateGroupName(ctx context.Context, message *UpdateGroupNameRequest) (*UpdateGroupNameResult, error) {
	fmt.Println("Task UpdateGroupName | ", message.Old, " | ", message.New)

	err := s.Queue.UpdateGroupName(message.Old, message.New)

	if err != nil {
		fmt.Println("Task UpdateGroupName Response | Error | ", message.Old, " | ", message.New, " | ", err)
		return &UpdateGroupNameResult{Old: "", New: ""}, err
	}

	fmt.Println("Task UpdateGroupName Response | Success | ", message.Old, " | ", message.New)
	return &UpdateGroupNameResult{Old: message.Old, New: message.New}, nil
}

func (s *TaskServer) UpdateByGroupName(ctx context.Context, message *UpdateByGroupNameRequest) (*UpdateByGroupNameResponse, error) {
	fmt.Println("Task UpdateByGroupName | ", message.Group, " | ", message.CreatedAt, " | ", message.DeletedAt)

	tasks := s.Queue.UpdateByGroupName(message.Group, message.CreatedAt.AsTime(), message.DeletedAt.AsTime())

	result := []*TaskInfo{}

	for i := 0; i < len(tasks); i++ {
		result = append(result, ToGRPCTaskInfo(tasks[i]))
	}

	fmt.Println("Task UpdateByGroupName Response | Success | ", result)
	return &UpdateByGroupNameResponse{Tasks: result}, nil
}
