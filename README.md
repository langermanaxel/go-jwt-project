# Go JWT Project

Este proyecto es una implementación de autenticación JWT en Go utilizando el framework Gin.

## Descripción

El `go-jwt-project` es una API RESTful que proporciona autenticación y autorización utilizando JSON Web Tokens (JWT). Este proyecto incluye rutas de autenticación y de usuario, así como ejemplos de cómo proteger rutas con middleware JWT.

## Características

- Registro y autenticación de usuarios
- Generación y validación de tokens JWT
- Protección de rutas con middleware JWT
- Ejemplos de rutas protegidas y no protegidas

## Requisitos

- Go 1.16 o superior
- Gin Gonic
- Golang JWT (github.com/dgrijalva/jwt-go)
- Bcrypt (golang.org/x/crypto/bcrypt)

## Instalación

1. Clona el repositorio:

   ```sh
   git clone https://github.com/langermanaxel/go-jwt-project.git
   cd go-jwt-project
   ```

2. Instala las dependencias:

   ```sh
   go mod tidy
   ```

## Uso

1. Configura las variables de entorno necesarias. Por ejemplo, puedes usar un archivo `.env` para definir el puerto del servidor y la clave secreta para los tokens JWT.

   ```sh
   export PORT=8000
   export JWT_SECRET=tu_clave_secreta
   ```

2. Inicia el servidor:

   ```sh
   go run main.go
   ```

3. Usa herramientas como `curl` o Postman para interactuar con la API.

### Ejemplos de rutas

- **Registro de usuario:**

  ```sh
  POST /register
  {
      "username": "tu_usuario",
      "password": "tu_contraseña"
  }
  ```

- **Inicio de sesión:**

  ```sh
  POST /login
  {
      "username": "tu_usuario",
      "password": "tu_contraseña"
  }
  ```

- **Ruta protegida:**

  ```sh
  GET /api-1
  Authorization: Bearer <tu_token_jwt>
  ```

## Contribuciones

Las contribuciones son bienvenidas. Por favor, abre un issue o un pull request para discutir cualquier cambio que te gustaría hacer.

## Licencia

Este proyecto está licenciado bajo la Licencia MIT. Consulta el archivo `LICENSE` para más detalles.
