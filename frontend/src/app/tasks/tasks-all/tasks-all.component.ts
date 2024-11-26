  import { Component, OnInit } from '@angular/core';
  import { TaskService } from '../../services/task.service';
  import { ActivatedRoute } from '@angular/router';
  import { Task } from '../../model/task';
  import {Router} from '@angular/router';
  import { Status } from '../../model/status';

  @Component({
    selector: 'app-tasks-all',
    templateUrl: './tasks-all.component.html',
    styleUrl: './tasks-all.component.css'
  })
  export class TasksAllComponent implements OnInit{
    id: string | null = null;
    constructor(private tasksService:TaskService,private route: ActivatedRoute, private router: Router){}
    tasks?:Task[];

    ngOnInit(): void {
        this.getAll();
    }

    getAll() {
      this.id = this.route.snapshot.paramMap.get('projectId');
      console.log("id:" + this.id);
      if (this.id) {
        this.tasksService.getAllTasksByProjectId(this.id).subscribe(
          (response) => {
            this.tasks = response.tasks;
            console.log(this.tasks);
          },
          (error) => {
            console.error('Error fetching tasks:', error);
          }
        );
      }
    }
    viewDetails(task: any) {
      console.log('Viewing details for:', task);
      // Navigate to the task details page with the task ID
      this.router.navigate(['/task-details', task.id]);
    }
  }
