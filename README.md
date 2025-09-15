# MI de Concorrência e Conectividade

Este projeto utiliza **Docker** e **Makefile** para automatizar a execução de um servidor, um cliente e testes de estresse em ambientes conteinerizados.

### Pré-requisitos

Certifique-se de que o Docker esteja instalado em sua máquina.

### Execução


O `Makefile` simplifica a execução das tarefas, eliminando a necessidade de digitar longos comandos Docker.

#### 1. Construir todas as imagens

Para compilar as imagens do servidor, cliente e teste de estresse, use o comando:
```bash
make build
```

#### 2. Executar o servidor e cliente na mesma rede local

Para rodar o servidor e o cliente em contêineres, use os comandos a seguir. O cliente e o servidor se conectarão usando a rede local do Docker.
Inicie o servidor:
```bash
make run-server
```

Em um novo terminal, inicie o cliente:
```bash
make run-client
```

#### 3. Executar o servidor e cliente em computadores diferentes

Para rodar os contêineres em máquinas distintas, siga estes passos:
Na máquina do servidor:

Obtenha o IP local da máquina (ex: 192.168.1.10).
Execute o servidor:
```bash
make run-server
```

Na máquina do cliente:
Edite o arquivo client.go para usar o IP local da máquina do servidor. Substitua localhost pelo IP:

```go
// Exemplo:
conn, err := net.Dial("tcp", "IP_LOCAL_DO_SERVIDOR:8080")
```

Construa a imagem do cliente novamente:
```bash
make build-client
```

Execute o cliente:
```bash
make run-client
```

#### 4. Executar os testes de estresse

Para rodar o teste de estresse contra o servidor, execute:
```Bash
make run-stress
```

#### 5. Limpeza

Para parar e remover todos os contêineres em execução, use o comando de limpeza:

```Bash
make clean
```
