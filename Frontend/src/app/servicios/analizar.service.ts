import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from "@angular/common/http";
import { URL } from "../link/URL";
import { Observable } from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class AnalizarService {

  static el: Array<any>;  //ERROES
  static sl: Array<any>; //SINBOLOS
  static sm: Array<any>; //METODOS
  static dot: string; //CODIGO DOT

  constructor(private http: HttpClient) { }

  ejecutar(codigo: any): Observable<any> {
    const httpOptions = {
      headers: new HttpHeaders({
        'Content-Type': 'application/json'
      }),
    };
    return this.http.post<any>(URL + 'analizar', codigo);
  }

  setErrores(lista:Array<any>):void{
    AnalizarService.el = lista;
  }

  getErrores(): Array<any>{
    return AnalizarService.el;
  }

  setSimbolos(lista:Array<any>):void{
    AnalizarService.sl = lista;
  }

  getSimbolos(): Array<any>{
    return AnalizarService.sl;
  }

  setMetodos(lista:Array<any>):void{
    AnalizarService.sm = lista;
  }

  getMetodos(): Array<any>{
    return AnalizarService.sm;
  }

  setDOT(lista:string):void{
    AnalizarService.dot = lista;
  }

  getDOT(): string{
    return AnalizarService.dot;
  }
}
