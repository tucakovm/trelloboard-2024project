import { Component } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { UserService } from '../services/user.service';

@Component({
  selector: 'app-change-password',
  templateUrl: './pass-recovery.component.html',
  styleUrls: ['./pass-recovery.component.css'],
})
export class PassRecoveryComponent {
  changePasswordForm: FormGroup;
  errorOccurred: boolean = false;
  errorMessage: string = '';
  successMessage: string = '';

  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private userService: UserService,
    private router: Router
  ) {
    this.changePasswordForm = this.fb.group({
      newPassword: [
        '',
        [
          Validators.required,
          Validators.minLength(8),
          Validators.pattern(
            /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/
          ),
        ],
      ],
      confirmNewPassword: ['', [Validators.required]],
      code:['', [Validators.required]]
    });
  }

  get newPassword() {
    return this.changePasswordForm.get('newPassword');
  }

  get confirmNewPassword() {
    return this.changePasswordForm.get('confirmNewPassword');
  }
  get code(){
    return this.changePasswordForm.get('code')
  }

  onSubmit() {
    if (this.changePasswordForm.invalid) {
      this.errorOccurred = true;
      this.errorMessage = 'Please fill out the form correctly.';
      return;
    }

    const { newPassword, confirmNewPassword, code } = this.changePasswordForm.value;

    if (newPassword !== confirmNewPassword) {
      this.errorOccurred = true;
      this.errorMessage = 'New password and confirm password do not match.';
      return;
    }

    const username = this.route.snapshot.queryParamMap.get('username');


    if (!username) {
      this.errorOccurred = true;
      this.errorMessage = 'Required parameters are missing in the URL.';
      return;
    }

<<<<<<< HEAD
    console.log(code)


    this.userService.recoverPassword(username, newPassword, code).subscribe({
=======

    this.userService.recoverPassword(username, newPassword).subscribe({
>>>>>>> c4b5a3c39341525041d46da04f5e479b7ee7b10a
      next: (res) => {
        this.successMessage = 'Password changed successfully!';
        this.router.navigate(['/login']);
      },
      error: (err) => {
        this.errorOccurred = true;
        this.errorMessage = err.error?.message ||  'Error while changing the password.';
        console.error(err);
      },
    });
  }
}
