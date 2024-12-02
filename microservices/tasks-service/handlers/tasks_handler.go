package handlers

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strings"
	proto "tasks-service/proto/task"
	"tasks-service/service"
	//"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskHandler struct {
	service        *service.TaskService // Use a pointer here
	projectService proto.ProjectServiceClient
	proto.UnimplementedTaskServiceServer
	natsConn *nats.Conn
}

func NewTaskHandler(service *service.TaskService, projectService proto.ProjectServiceClient, natsConn *nats.Conn) *TaskHandler {
	return &TaskHandler{service: service,
		projectService: projectService,
		natsConn:       natsConn}
}

//func (h *TaskHandler) DoneTasksByProject(ctx context.Context, req *proto.DoneTasksByProjectReq) (*proto.DoneTasksByProjectRes, error) {
//	is, err := h.service.DoneTasksByProject(req.ProjId)
//	if err != nil {
//		return nil, status.Error(codes.InvalidArgument, "bad request ...")
//	}
//	doneTasksByProjectReq := &proto.DoneTasksByProjectRes{
//		IsDone: is,
//	}
//	return doneTasksByProjectReq, nil
//}

func (h *TaskHandler) Delete(ctx context.Context, req *proto.DeleteTaskReq) (*proto.EmptyResponse, error) {
	err := h.service.DeleteTask(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func (h *TaskHandler) Create(ctx context.Context, req *proto.CreateTaskReq) (*proto.EmptyResponse, error) {
	log.Println(req.Task)
	err := h.service.Create(req.Task)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	subject := "create-task"

	message := map[string]string{
		"TaskName":  req.Task.Name,
		"ProjectId": req.Task.ProjectId,
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification message: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create notification message...")
	}

	err = h.natsConn.Publish(subject, messageData)
	if err != nil {
		log.Printf("Error publishing notification: %v", err)
		return nil, status.Error(codes.Internal, "Failed to send notification...")
	}

	log.Printf("Notification sent: %s", string(messageData))

	return nil, nil
}

func (h *TaskHandler) GetById(ctx context.Context, req *proto.GetByIdReq) (*proto.TaskResponse, error) {
	task, err := h.service.GetById(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	response := &proto.TaskResponse{Task: task}
	return response, nil
}

func (h *TaskHandler) GetAllByProjectId(ctx context.Context, req *proto.GetAllTasksReq) (*proto.GetAllTasksRes, error) {
	allTasks, err := h.service.GetTasksByProjectId(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch tasks")
	}
	response := &proto.GetAllTasksRes{Tasks: allTasks}
	return response, nil
}

func (h *TaskHandler) AddMemberTask(ctx context.Context, req *proto.AddMemberTaskReq) (*proto.EmptyResponse, error) {
	// Provera da li je kontekst istekao pre bilo čega
	select {
	case <-ctx.Done():
		log.Printf("Handler detected context cancellation or timeout: %v", ctx.Err())
		return nil, status.Error(codes.DeadlineExceeded, "Request timed out or was canceled")
	default:
	}
	task, _ := h.service.GetById(req.TaskId)
	userOnProjectReq := &proto.UserOnOneProjectReq{
		UserId:    req.User.Username,
		ProjectId: task.ProjectId,
	}

	// Provera timeout-a pre pozivanja udaljenog servisa
	select {
	case <-ctx.Done():
		log.Printf("Context timeout before calling project service: %v", ctx.Err())
		return nil, status.Error(codes.DeadlineExceeded, "Request timed out or was canceled")
	default:

	}
	//time.Sleep(5 * time.Second) // test : Request timeout
	projServiceResponse, err := h.projectService.UserOnOneProject(ctx, userOnProjectReq)
	if err != nil {
		log.Printf("Error checking project: %v", err)
		return nil, status.Error(codes.Internal, "Error checking project")
	}

	if projServiceResponse.IsOnProj {
		taskId := req.TaskId

		// Provera timeout-a pre dodavanja člana
		select {
		case <-ctx.Done():
			log.Printf("Context timeout before adding member: %v", ctx.Err())
			return nil, status.Error(codes.DeadlineExceeded, "Request timed out or was canceled")
		default:
			// Nastavlja sa dodavanjem člana
		}
		subject := "add-to-task"
		err = h.service.AddMember(taskId, req.User)
		if err != nil {
			log.Printf("Error adding member to project: %v", err)
			return nil, status.Error(codes.InvalidArgument, "Error adding member...")
		}
		projectFromTaskReq := &proto.GetByIdReq{
			Id: taskId,
		}
		projectFromTask, _ := h.GetById(ctx, projectFromTaskReq)
		message := map[string]string{
			"UserId":    req.User.Id,
			"TaskId":    taskId,
			"ProjectId": projectFromTask.Task.ProjectId,
		}

		messageData, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling notification message: %v", err)
			return nil, status.Error(codes.Internal, "Failed to create notification message...")
		}

		err = h.natsConn.Publish(subject, messageData)
		if err != nil {
			log.Printf("Error publishing notification: %v", err)
			return nil, status.Error(codes.Internal, "Failed to send notification...")
		}

		log.Printf("Notification sent: %s", string(messageData))

		return nil, nil
	} else {
		return nil, status.Error(codes.Internal, "User is not assigned to a project.")
	}
}

func (h *TaskHandler) RemoveMemberTask(ctx context.Context, req *proto.RemoveMemberTaskReq) (*proto.EmptyResponse, error) {
	taskId := req.TaskId
	err := h.service.RemoveMember(taskId, req.UserId)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		return nil, status.Error(codes.InvalidArgument, "Error removing member...")
	}
	projectFromTaskReq := &proto.GetByIdReq{
		Id: taskId,
	}
	projectFromTask, _ := h.GetById(ctx, projectFromTaskReq)
	subject := "remove-from-task"
	message := map[string]string{
		"UserId":    req.UserId,
		"TaskId":    taskId,
		"ProjectId": projectFromTask.Task.ProjectId,
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification message: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create notification message...")
	}

	err = h.natsConn.Publish(subject, messageData)
	if err != nil {
		log.Printf("Error publishing notification: %v", err)
		return nil, status.Error(codes.Internal, "Failed to send notification...")
	}

	log.Printf("Notification sent: %s", string(messageData))

	return nil, nil
}
func (h *TaskHandler) UpdateTask(ctx context.Context, req *proto.UpdateTaskReq) (*proto.EmptyResponse, error) {
	log.Println("Received UpdateTask request for task ID:", req.Id)

	// Validate the task exists
	existingTask, err := h.service.GetById(req.Id)
	if err != nil {
		log.Printf("Error fetching task for update: %v", err)
		return nil, status.Error(codes.NotFound, "Task not found")
	}

	// Update the fields of the task
	updatedTask := existingTask
	updatedTask.Name = req.Name
	updatedTask.Description = req.Description
	updatedTask.Status = req.Status
	updatedTask.Members = req.Members

	// Call the service layer to save changes
	err = h.service.UpdateTask(updatedTask)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		return nil, status.Error(codes.Internal, "Failed to update task")
	}

	log.Println("Task updated successfully:", req.Id)

	getProjReq := &proto.GetByIdReq{
		Id: req.Id,
	}

	projId, _ := h.GetById(ctx, getProjReq)

	subject := "update-task"
	message := map[string]string{
		"TaskId":     req.Id,
		"TaskStatus": req.Status,
		"ProjectId":  projId.Task.ProjectId,
	}
	if len(req.Members) > 0 {
		var memberIds []string
		for _, member := range req.Members {
			memberIds = append(memberIds, member.Id)
		}
		message["MemberIds"] = strings.Join(memberIds, ",")
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification message: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create notification message...")
	}

	err = h.natsConn.Publish(subject, messageData)
	if err != nil {
		log.Printf("Error publishing notification: %v", err)
		return nil, status.Error(codes.Internal, "Failed to send notification...")
	}

	log.Printf("Notification sent: %s", string(messageData))
	return &proto.EmptyResponse{}, nil
}
