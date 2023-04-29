package main

import(
	"archivos/programa"
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/rs/cors"
	"io"
	"log"
)

func main(){
	//VARIABLES
    var sesion programa.Usuario
    var discos []programa.Disco

	mult := http.NewServeMux()

	//INICIO
	mult.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		fmt.Println("Servidor en Linea")
	})

	//CONSOLA DE COMANDOS
	mult.HandleFunc("/consola", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		var contenidos [6]string
		contenidos[1] = "0"

		//Leer el JSON y desencriptarlo
		var entrada EntradaConsola
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &entrada)
	
		//Ejecutar el codigo
		programa.Consola(&sesion, &discos, &contenidos, &entrada.Comandos)
	
		//Enviar la respuesta
		respuesta := Salida{
			Consola: contenidos[0],
			Login:   contenidos[1],
			File:    contenidos[2],
			Tree:    contenidos[3],
			Sb:      contenidos[4],
			Disk:    contenidos[5]}
	
		jsonResponse, jsonError := json.Marshal(respuesta)
	
		if jsonError != nil {
			fmt.Println("ERROR: Error al codificar la salida.")
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})

	//LOGIN
	mult.HandleFunc("/login", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		var contenidos [6]string
		contenidos[1] = "0"

		//Leer el JSON y desencriptarlo
		var entrada EntradaLogin
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &entrada)
		
		//CREAR LOS COMANDOS A EJECUTAR
		/*
			-Se debe de crear un logout para cerrar cualquier sesion
			-Se debe de crear un login para verificar que el usuario existe
			-Se debe de crear los comandos de los reportes
		*/
		comandos := ""
		comandos += "logout \n"
		comandos += "login >user=" + entrada.Username + " >pass=" + entrada.Password + " >id=" + entrada.Particion + " \n"
		comandos += "rep >id=" + entrada.Particion +  " >Path=\"/home/jjonk19/Documentos/Ingenieria/Proyectos/MIA/disk.pdf\" >name=disk \n"
		comandos += "rep >id=" + entrada.Particion +  " >Path=\"/home/jjonk19/Documentos/Ingenieria/Proyectos/MIA/tree.pdf\" >name=tree \n"
		comandos += "rep >id=" + entrada.Particion +  " >Path=\"/home/jjonk19/Documentos/Ingenieria/Proyectos/MIA/sb.pdf\" >name=sb \n"

		//Ejecutar el codigo
		programa.Consola(&sesion, &discos, &contenidos, &comandos)
	
		//Enviar la respuesta
		respuesta := Salida{
			Consola: contenidos[0],
			Login:   contenidos[1],
			File:    contenidos[2],
			Tree:    contenidos[3],
			Sb:      contenidos[4],
			Disk:    contenidos[5]}
	
		jsonResponse, jsonError := json.Marshal(respuesta)
	
		if jsonError != nil {
			fmt.Println("ERROR: Error al codificar la salida.")
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})

	//FILE
	mult.HandleFunc("/file", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	
		var contenidos [6]string
		contenidos[1] = "0"

		//Leer el JSON y desencriptarlo
		var entrada EntradaFile
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &entrada)

		//CREAR LOS COMANDOS A EJECUTAR
		/*
			-El login en teoria ya existe
			-Se debe de crear el comando de file nada mas
		*/
		comandos := ""
		comandos += "rep >id=" + entrada.Particion +  " >Path=\"/home/jjonk19/Documentos/Ingenieria/Proyectos/MIA/file.txt\" >ruta=" + entrada.Ruta + " >name=file"

		//Ejecutar el codigo
		programa.Consola(&sesion, &discos, &contenidos, &comandos)
	
		//Enviar la respuesta
		respuesta := Salida{
			Consola: contenidos[0],
			Login:   contenidos[1],
			File:    contenidos[2],
			Tree:    contenidos[3],
			Sb:      contenidos[4],
			Disk:    contenidos[5]}
	
		jsonResponse, jsonError := json.Marshal(respuesta)
	
		if jsonError != nil {
			fmt.Println("ERROR: Error al codificar la salida.")
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})
	
	handler := cors.Default().Handler(mult)
	fmt.Println("Servidor en Puerto 3000")
	log.Fatal(http.ListenAndServe(":3000", handler))
}
