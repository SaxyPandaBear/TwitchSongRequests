import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { LoginComponent } from './login/login.component';
import { LoginRoutingModule } from './login-routing.module';

import { NgWizardModule, NgWizardConfig, THEME } from 'ng-wizard';

const ngWizardConfig: NgWizardConfig = {
    theme: THEME.default,
};

@NgModule({
    declarations: [LoginComponent],
    imports: [
        CommonModule,
        LoginRoutingModule,
        NgWizardModule.forRoot(ngWizardConfig),
    ],
})
export class LoginModule {}
