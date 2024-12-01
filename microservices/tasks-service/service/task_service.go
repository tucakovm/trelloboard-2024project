package service

import (
	"context"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strings"
	"tasks-service/domain"
	proto "tasks-service/proto/task"
	"tasks-service/repository"
)

type TaskService struct {
	repo   repository.TaskRepo
	Tracer trace.Tracer
}

func NewTaskService(repo repository.TaskRepo, tracer trace.Tracer) *TaskService {
	return &TaskService{repo: repo, Tracer: tracer}
}

func (s *TaskService) Create(taskReq *proto.Task, ctx context.Context) error {
	newTask := &domain.Task{
		Name:        taskReq.Name,
		Description: taskReq.Description,
		Status:      0,
		ProjectID:   taskReq.ProjectId,
		Members:     make([]domain.User, 0),
	}
	log.Println(newTask)
	return s.repo.Create(*newTask, ctx)
}

func (s *TaskService) DeleteTask(id string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.deleteTask")
	defer span.End()
	return s.repo.Delete(id, ctx)
}

func (s *TaskService) DoneTasksByProject(id string, ctx context.Context) (bool, error) {
	ctx, span := s.Tracer.Start(ctx, "s.doneTasksByProject")
	defer span.End()
	return s.repo.HasIncompleteTasksByProject(id, ctx)
}

func (s *TaskService) GetById(id string, ctx context.Context) (*proto.Task, error) {
	ctx, span := s.Tracer.Start(ctx, "s.verify")
	defer span.End()
	task, err := s.repo.GetById(id, ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "DB exception.")
	}
	var protoMembers []*proto.User
	for _, member := range task.Members {
		protoMembers = append(protoMembers, &proto.User{
			Id:       member.Id,
			Username: member.Username,
			Role:     member.Role,
		})
	}
	protoTask := &proto.Task{
		Id:          task.Id.Hex(),
		Name:        task.Name,
		Description: task.Description,
		Status:      task.Status.String(),
		ProjectId:   task.ProjectID,
		Members:     protoMembers,
	}

	return protoTask, nil
}

func (s *TaskService) DeleteTasksByProjectId(id string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.deleteTasksByProjectId")
	defer span.End()
	return s.repo.DeleteAllByProjectID(id, ctx)
}

func (s *TaskService) GetTasksByProjectId(id string, ctx context.Context) ([]*proto.Task, error) {
	ctx, span := s.Tracer.Start(ctx, "s.getTasksByProjectId")
	defer span.End()
	tasks, err := s.repo.GetAllByProjectID(id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "DB exception.")
	}
	log.Println("SErvice tasks")
	log.Println(tasks)
	var protoTasks []*proto.Task
	for _, dp := range tasks {
		var protoMembers []*proto.User
		for _, member := range dp.Members {
			protoMembers = append(protoMembers, &proto.User{
				Id:       member.Id,
				Username: member.Username,
				Role:     member.Role,
			})
		}
		protoTasks = append(protoTasks, &proto.Task{
			Id:          dp.Id.Hex(),
			Name:        dp.Name,
			Description: dp.Description,
			Status:      dp.Status.String(),
			ProjectId:   dp.ProjectID,
			Members:     protoMembers,
		})
	}
	log.Println(protoTasks)
	return protoTasks, err
}

func (t *TaskService) AddMember(projectId string, protoUser *proto.User, ctx context.Context) error {
	ctx, span := t.Tracer.Start(ctx, "s.addMember")
	defer span.End()
	task, err := t.repo.GetById(projectId, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.NotFound, "Project not found")
	}

	user := &domain.User{
		Id:       protoUser.Id,
		Username: protoUser.Username,
		Role:     protoUser.Role,
	}
	log.Println("TASK SERVICE gettask repo->: ", task)
	for _, member := range task.Members {
		if strings.EqualFold(strings.TrimSpace(member.Username), strings.TrimSpace(user.Username)) {
			return status.Error(codes.AlreadyExists, "Member already part of the task")
		}
	}
	return t.repo.AddMember(projectId, *user, ctx)
}

func (t *TaskService) RemoveMember(projectId string, userId string, ctx context.Context) error {
	ctx, span := t.Tracer.Start(ctx, "s.removeMamber")
	defer span.End()
	return t.repo.RemoveMember(projectId, userId, ctx)
}
func (s *TaskService) UpdateTask(taskReq *proto.Task, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.updateTask")
	defer span.End()
	// Fetch the existing task
	existingTask, err := s.repo.GetById(taskReq.Id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.NotFound, "Task not found")
	}

	existingTask.Name = taskReq.Name
	existingTask.Description = taskReq.Description

	statusEnum, err := domain.ParseTaskStatus2(taskReq.Status)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.InvalidArgument, "Invalid task status")
	}

	existingTask.Status = statusEnum

	// Update members if provided
	if len(taskReq.Members) > 0 {
		var updatedMembers []domain.User
		for _, member := range taskReq.Members {
			updatedMembers = append(updatedMembers, domain.User{
				Id:       member.Id,
				Username: member.Username,
				Role:     member.Role,
			})
		}
		existingTask.Members = updatedMembers
	}

	// Call the repository to persist the changes
	err = s.repo.Update(*existingTask, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return status.Error(codes.Internal, "Failed to update task in the database")
	}

	return nil
}
