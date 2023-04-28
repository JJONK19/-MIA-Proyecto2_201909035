import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ConsolaComponent } from './consola/consola.component';
import { LoginComponent } from './login/login.component';
import { ReportesComponent } from './reportes/reportes.component';
const routes: Routes = [{path: "", component:ConsolaComponent}, {path:"reportes", component:ReportesComponent}, {path:"login", component:LoginComponent}];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
