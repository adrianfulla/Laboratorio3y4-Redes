# Laboratorio3y4-Redes
 
## Objetivos del laboratorio
- Conocer los algoritmos de enrutamiento utilizados en las implementaciones actuales de Internet.
- Comprender cómo funcionan las tablas de enrutamiento.
- Implementar los algoritmos de enrutamiento en una red simulada sobre el protocolo XMPP.
- Analizar el funcionamiento de los algoritmos de enrutamiento.

## Utilización del cliente
El programa fue escrito completamente utilizando Go con la librería de Fyne para la interfaz de usuario. El programa fue empaquetado para la conveniencia de distribución. 

Los requisitos para utilizar el cliente son descargar la aplicación específica para el sistema operativo y que se encuentren en el mismo directorio los archivos de nombres y topología con las convenciones:
- topo-*.txt
- names-*.txt

### Windows
Para Windows se debe utilizar el [Laboratorio3y4-Redes.exe](https://github.com/adrianfulla/Laboratorio3y4-Redes/blob/main/Laboratorio3y4-Redes.exe)

Este debería ser suficiente junto con los .txt, es posible que haya que dar permisos para correr la aplicación.

### MacOS
En MacOS se debe utilizar el paquete para MacOS [Laboratorio3y4-Redes-Grupo4-MacOS](https://github.com/adrianfulla/Laboratorio3y4-Redes/blob/main/Laboratorio3y4-Redes-Grupo4-MacOS)


En la terminal se debe navegar al directorio con el paquete y los .txt y ejecutar el comando:
```bash
    ./Laboratorio3y4-Redes-Grupo4-MacOS
```

### Ejecución del código directamente
Alternativamente se puede compilar el código utilizando Go y Fyne en línea de comando.


#### Requisitos de instalación
- GoLang -> https://go.dev/doc/install
- Fyne -> https://docs.fyne.io/started/

#### Instalación
1. Clonar repositorio:
```bash
    git clone https://github.com/adrianfulla/Laboratorio3y4-Redes.git
```
2. Acceder al directorio del repositorio:
```bash
    cd Proyecto1-Redes/
```

3. Ejecutar comando:
```bash
    go mod tidy
```
Este comando instalara todas las dependencias necesarias

4. Ejecutar el comando go run para compilar y ejecutar el cliente:
```bash
    go run ./server
```