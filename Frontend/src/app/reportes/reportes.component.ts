import { Component, AfterViewInit } from '@angular/core';
import * as d3graphviz from 'd3-graphviz';
import { ElementRef, ViewChild } from '@angular/core';
import { AnalizarService } from 'src/app/servicios/analizar.service';
import { FormControl, FormGroup, Validators } from '@angular/forms';

@Component({
  selector: 'app-reportes',
  templateUrl: './reportes.component.html',
  styleUrls: ['./reportes.component.css']
})
export class ReportesComponent implements AfterViewInit {
  @ViewChild('disk') disk!: ElementRef;
  @ViewChild('tree') tree!: ElementRef;
  @ViewChild('sb') sb!: ElementRef;
  salida:  string = "";    //Texto que se va amostrar en consola
  ruta:  string = "";   //Codigo de entrada. Se envia al analizador
  form: FormGroup;

  constructor(private analizarService: AnalizarService) {
    this.form = new FormGroup({
        ruta: new FormControl('', [Validators.required])
      }
    )
   }
  
  ngAfterViewInit(): void {
    setTimeout(() => {
      // Inicializamos el gráfico
      d3graphviz.graphviz(this.disk.nativeElement)
        .renderDot('digraph {a -> disk}');

      d3graphviz.graphviz(this.tree.nativeElement)
        .renderDot('digraph {a -> tree}');

      d3graphviz.graphviz(this.sb.nativeElement)
        .renderDot('digraph {a -> sb}');
    });
  }

  enviarCodigo(): void {
    this.ruta = this.form.controls["ruta"].value; 
    var objeto = {
      entrada: this.ruta,
    }
    /*
    this.analizarService.ejecutar(objeto).subscribe((res:any)=>{
      console.log(res)
      this.salida = (res.salida);
      //ACA SE CAMBIA PARA ALMACENAR VARIABLES GLOBALES Y COSAS DE LA SALIDA
      this.analizarService.setErrores(res.errores.lista);
      this.analizarService.setSimbolos(res.simbolos.lista);
      this.analizarService.setMetodos(res.metodos.lista);
      this.analizarService.setDOT(res.ast);
    }, err=>{
      console.log(err)
    });
    */
  }
}
