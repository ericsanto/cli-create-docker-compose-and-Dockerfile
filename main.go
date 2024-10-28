package main

import (
	"cli/config"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	appName     string
	dbUser      string
	dbPass      string
	dbName      string
	nodeVersion string
	outputDir   string
	ipServer    string
	framework   string
)

var rootCmd = &cobra.Command{Use: "tatu"}

var cmd = &cobra.Command{
	Use:   "setup-dev",
	Short: "Cria um Dockerfile e um docker-compose.yml para aplicação React Native com PostgreSQL",
	Run: func(cmd *cobra.Command, args []string) {
		generateDockerfile()
		generateDockerCompose()
		config.UpDockerCompose()
	},
}

func InitFunc() {
	cmd.Flags().StringVarP(&appName, "app", "a", "my-react-app", "Nome da aplicação")
	cmd.Flags().StringVarP(&dbUser, "dbuser", "u", "postgres", "Nome para usuário postgres")
	cmd.Flags().StringVarP(&dbPass, "dbpass", "p", "password", "Senha para usuário postgres")
	cmd.Flags().StringVarP(&dbName, "dbname", "n", "postgres", "Nome do database")
	cmd.Flags().StringVarP(&nodeVersion, "nodeVersion", "v", "20", "Versão do NodeJs")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "output dos arquivos Dockerfile e docker-compose.yml")
	cmd.Flags().StringVarP(&ipServer, "ipServer", "i", "127.0.0.1", "Define o endereço IP no qual a exposição deve ser executada dentro do contêiner")
	cmd.Flags().StringVarP(&framework, "framework", "f", "gin-gonic", "Framework para gerar o Dockerfile")

}

func generateDockerfile() {

	switch framework {

	case "gin-gonic":
		dockerfileContent := fmt.Sprintf(`FROM golang:latest AS builder 

		WORKDIR /app 
		COPY go.mod ./ 
		RUN go mod download 

		COPY *.go ./ 
		RUN go build -o main main.go 

		FROM alpine:latest AS runtime 

		WORKDIR /app 
		COPY --from=builder /app/main . 
		EXPOSE 8080 

		CMD ["chmod", "+x", "main"] 
		CMD ["./main"]`)

		WriteDockerfile(dockerfileContent)

	case "node":
		dockerfileContent := fmt.Sprintf(`
		
		FROM node:17.9.0-alpine3.15

		WORKDIR /usr/src/app

		COPY package*.json ./

		RUN npm install --only=production

		COPY --from=builder /usr/src/app/dist ./

		EXPOSE 3000

		ENTRYPOINT ["node","./app.js"]`)

		WriteDockerfile(dockerfileContent)

	case "django":
		dockerfileContent := fmt.Sprintf(`
		FROM python:3.8

		WORKDIR /app

		COPY requirements.txt /app
		RUN pip install -r requirements.txt
		
		COPY . /app
		RUN python manage.py collectstatic --no-input

		EXPOSE 8000
		CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]`)

		WriteDockerfile(dockerfileContent)
	}

}

func generateDockerCompose() {
	dockerCompose := fmt.Sprintf(`
services:

  db:
    image: postgres:17
    environment:
      POSTGRES_USER: %s
      POSTGRES_PASSWORD: %s
      POSTGRES_DB: %s
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  react-native:
    build:
      context: .
      dockerfile: Dockerfile
    ports: 
      - "19001:19001"
      - "19002:19002"
      - "19006:19006"
    depends_on:
      - db
    volumes:
      - ./:/opt/my-app

volumes:
  pgdata:
`, dbUser, dbPass, dbName)

	dockerComposePath := outputDir + "/docker-compose.yml"
	if err := os.WriteFile(dockerComposePath, []byte(dockerCompose), 0644); err != nil {
		fmt.Println("Não foi possível criar o arquivo docker-compose.yml no diretório especificado:", err)
		return
	}

	fmt.Println("docker-compose criado com sucesso!")
}

func ExecuteCommand() {
	InitFunc() // Chama a função InitFunc para configurar as flags
	rootCmd.AddCommand(cmd)
	rootCmd.Execute()
}

func main() {

	ExecuteCommand()
}

func WriteDockerfile(dockerfile string) {

	dockerfilePath := outputDir + "/Dockerfile"

	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		fmt.Println("Não foi possível criar o Dockerfile no diretório especificado")
		return
	}

	fmt.Println("Diretório criado com sucesso!")

}
