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
        .renderDot(this.analizarService.getDisk());

      d3graphviz.graphviz(this.tree.nativeElement)
        .renderDot(this.analizarService.getTree());

      d3graphviz.graphviz(this.sb.nativeElement)
        .renderDot(this.analizarService.getSB());
    });
  }

  enviarCodigo(): void {
    this.ruta = this.form.controls["ruta"].value; 
    var objeto = {
      ruta: this.ruta,
      particion: this.analizarService.getParticion()
    }
  
    this.analizarService.ejecutarFile(objeto).subscribe((res:any)=>{
      console.log(res)
      this.salida = (res.File);
    }, err=>{
      console.log(err)
    });
    
  }
}
