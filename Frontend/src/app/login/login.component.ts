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
      entrada: "",
    }
    /*
    this.analizarService.ejecutar(objeto).subscribe((res:any)=>{
      console.log(res)
      //ACA SE CAMBIA PARA ALMACENAR VARIABLES GLOBALES Y COSAS DE LA SALIDA
      
      this.analizarService.setErrores(res.errores.lista);
      this.analizarService.setSimbolos(res.simbolos.lista);
      this.analizarService.setMetodos(res.metodos.lista);
      this.analizarService.setDOT(res.ast);
      
    }, err=>{
      console.log(err)
    });
    */

    //ABRIR LA VENTANA DE LOS REPORTES
    //*Añador un mensaje de credenciales erroneas
    this.router.navigate(['/reportes']);
  }

}
