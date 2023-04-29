import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from "@angular/common/http";
import { URL } from "../link/URL";
import { Observable } from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class AnalizarService {

  static particion: string;    //ID de la particion
  static tree:      string;    //DOT del reporte arbol
  static sb:        string;    //DOT del superbloque
  static disk:      string;    //DOT del disco

  constructor(private http: HttpClient) { }

  ejecutarConsola(entrada: any): Observable<any> {
    const httpOptions = {
      headers: new HttpHeaders({
        'Content-Type': 'application/json'
      }),
    };
    return this.http.post<any>(URL + 'consola', entrada);
  }

  ejecutarLogin(entrada: any): Observable<any> {
    const httpOptions = {
      headers: new HttpHeaders({
        'Content-Type': 'application/json'
      }),
    };
    return this.http.post<any>(URL + 'login', entrada);
  }

  ejecutarFile(entrada: any): Observable<any> {
    const httpOptions = {
      headers: new HttpHeaders({
        'Content-Type': 'application/json'
      }),
    };
    return this.http.post<any>(URL + 'file', entrada);
  }

  //VARIABLES GLOBALES
  setParticion(entrada:string):void{
    AnalizarService.particion = entrada;
  }

  getParticion(): string{
    return AnalizarService.particion;
  }

  setTree(entrada:string):void{
    AnalizarService.tree = entrada;
  }

  getTree(): string{
    return AnalizarService.tree;
  }

  setSB(entrada:string):void{
    AnalizarService.sb = entrada;
  }

  getSB(): string{
    return AnalizarService.sb;
  }

  setDisk(entrada:string):void{
    AnalizarService.disk = entrada;
  }

  getDisk(): string{
    return AnalizarService.disk;
  }

  
}
