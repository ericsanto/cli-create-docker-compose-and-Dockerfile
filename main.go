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

}

func generateDockerfile() {
	dockerfile := fmt.Sprintf(`
# pull base image
FROM node:23-alpine3.19

# set our node environment, either development or production
ARG NODE_ENV=production
ENV NODE_ENV $NODE_ENV

# default to port 19006 for node, and 19001 and 19002 (tests) for debug
ARG PORT=19006
ENV PORT $PORT
EXPOSE 19006 19001 19002
ENV REACT_NATIVE_PACKAGER_HOSTNAME=%s

# install global packages
ENV NPM_CONFIG_PREFIX=/home/node/.npm-global
ENV PATH /home/node/.npm-global/bin:$PATH
RUN npm i --unsafe-perm -g npm@latest expo-cli@latest
RUN apt-get update && apt-get install -y qemu-user-static
RUN yarn add @expo/ngrok

# install dependencies
RUN mkdir /opt/my-app && chown root:root /opt/my-app
WORKDIR /opt/my-app
COPY package.json package-lock.json ./
RUN yarn install

# copy in our source code last
COPY . /opt/my-app/

CMD ["npx","expo", "start", "--tunel"]`, ipServer)

	dockerfilePath := outputDir + "/Dockerfile"
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		fmt.Println("Não foi possível criar o Dockerfile no diretório mencionado:", err)
		return
	}

	fmt.Println("Dockerfile criado com sucesso!")
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
