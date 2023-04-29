import { Component, OnInit } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { AnalizarService } from 'src/app/servicios/analizar.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {
  form: FormGroup;
  particion: string = "";
  username: string = "";
  password: string = "";

  constructor(private analizarService: AnalizarService, private router: Router) {
    this.form = new FormGroup({
        particion: new FormControl('', [Validators.required]),
        username: new FormControl('', [Validators.required]),
        password: new FormControl('', [Validators.required])
      }
    )
   }  

   ngOnInit(): void {
    this.form = new FormGroup({
      particion: new FormControl('', [Validators.required]),
      username: new FormControl('', [Validators.required]),
      password: new FormControl('', [Validators.required])
      }
    )
  }

  enviarCodigo(): void {
    this.particion = this.form.controls["particion"].value;
    this.username = this.form.controls["username"].value;
    this.password = this.form.controls["password"].value; 
    //PENDIENTE HACER LA REQUEST
    var objeto = {
      particion: this.particion,
      username: this.username,
      password: this.password
    }
    
    this.analizarService.ejecutarLogin(objeto).subscribe((res:any)=>{
      console.log(res)
      //ACA SE CAMBIA PARA ALMACENAR VARIABLES GLOBALES Y COSAS DE LA SALIDA
      if(res.Login == "0"){
        console.log("No se pudo iniciar sesion.")
      }else{
        this.analizarService.setParticion(this.particion);
        this.analizarService.setDisk(res.Disk)
        this.analizarService.setSB(res.Sb)
        this.analizarService.setTree(res.Tree)

        //Abrir la ventana de los reportes
        this.router.navigate(['/reportes']);
      }
      
    }, err=>{
      console.log(err)
    });
  }

}
