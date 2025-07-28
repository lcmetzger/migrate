# Helper para a criação de migrations

## Documentação de Uso

### Pré-requisitos para compilar
- Go 1.24 ou superior instalado

### Instalação das dependências
Execute o comando abaixo na raiz do projeto:
```sh
go mod tidy
```

### Compilação usando Makefile

Para compilar os binários para diferentes sistemas operacionais, utilize os comandos abaixo na raiz do projeto:

```sh
# Compilar para todos os sistemas (Linux, macOS e Windows)
make

# Compilar apenas para Linux
make linux

# Compilar apenas para macOS
env GOARCH=amd64 make macos

# Compilar apenas para Windows
make windows

# Limpar os binários gerados
make clean
```

Os binários serão gerados na pasta `bin/`.

### Executando a CLI

Para criar uma nova migration, utilize:
```sh
go run main.go <nome_da_migration> [--dml|--ddl|--all]
```
Ou, se já compilado:
```sh
./bin/migrate <nome_da_migration> [--dml|--ddl|--all]
```

- Se o parâmetro `<nome_da_migration>` não for informado, o programa solicitará interativamente:
  ```
  Informe o nome para a migration:
  Deseja gerar DDL? (s/n):
  Deseja gerar DML? (s/n):
  ```
- Se nenhuma flag for informada, o programa criará os arquivos na pasta corrente
- Use `--dml` para criar apenas DML, `--ddl` para apenas DDL, ou `--all` para ambos sem perguntas.

**Nota:** O modo interativo funciona melhor quando executado diretamente no terminal. O uso de pipes (`echo | ./bin/migrate`) pode não funcionar corretamente devido a limitações na leitura do stdin.

### Comportamento dos parâmetros

O comportamento do aplicativo é o seguinte:

1. **Caso seja passado o parâmetro `--ddl`, ou no modo interativo indicado DDL como "s":**
   - Gera a pasta baseada em timestamp e os scripts `up.sql` e `down.sql` dentro da pasta `DDL/`

2. **Caso seja passado o parâmetro `--dml`, ou no modo interativo indicado DML como "s":**
   - Gera a pasta baseada em timestamp e os scripts `up.sql` e `down.sql` dentro da pasta `DML/`

3. **Caso não seja passado nenhum dos parâmetros:**
   - Gera a pasta baseada em timestamp e os scripts `up.sql` e `down.sql` dentro da pasta corrente

**Nota:** No modo interativo (quando nenhuma flag é informada), o programa pergunta se você deseja gerar DDL e/ou DML. Dependendo da sua escolha, os arquivos serão criados automaticamente nas pastas apropriadas (DDL/ ou DML/).

### O que acontece ao executar o comando

**Comportamento sem flags (pasta corrente):**
- Será criada uma subpasta na pasta atual com o nome:
  - `<timestamp>_<nome_da_migration>`
- Dentro da subpasta, serão criados os arquivos:
  - `up.sql`
  - `down.sql`

**Comportamento com flags específicas (`--ddl`, `--dml`, `--all`):**
- Será criada uma subpasta dentro de `DML` e/ou `DDL` com o nome:
  - Para DML: `<timestamp>_<nome_da_migration>`
  - Para DDL: `<timestamp>_<nome_da_migration>`
- Os diretórios DDL e DML serão criados automaticamente se não existirem
- Dentro de cada subpasta, serão criados os arquivos:
  - `up.sql`
  - `down.sql`

**Comportamento sem argumentos (modo interativo):**
- Se você selecionar DDL e/ou DML:
  - Será criada uma subpasta dentro de `DML` e/ou `DDL` com o nome:
    - Para DML: `<timestamp>_<nome_da_migration>`
    - Para DDL: `<timestamp>_<nome_da_migration>`
- Dentro de cada subpasta, serão criados os arquivos:
  - `up.sql`
  - `down.sql`

**Geral:**
- Os arquivos são criados com encoding ISO-8859-1
- Os arquivos criados contêm o seguinte comentário no início:
  - `-- Não esqueça de excluir este comentário, e verifique se o seu editor está definido para utilizar o encoding ISO-8859-1`
- A quebra de linha é definida automaticamente baseada no sistema operacional:
  - **Windows:** CRLF (`\r\n`)
  - **Unix/Linux/macOS:** LF (`\n`)
- Ao final, será exibida a lista dos arquivos criados

### Exemplos de uso

```sh
# Sem flags - cria na pasta corrente
./bin/migrate teste_2
# Resultado (cria na pasta atual):
Arquivos criados:
  - 20250728103214655_teste_2/up.sql
  - 20250728103214655_teste_2/down.sql
```

```sh
# Modo interativo (sem argumentos)
./bin/migrate
Informe o nome para a migration: adicionar_usuario
Deseja gerar DDL? (s/n): s
Deseja gerar DML? (s/n): s

# Resultado (cria dentro de DDL/ e DML/):
Arquivos criados:
  - DDL/20240607180000000_adicionar_usuario/up.sql
  - DDL/20240607180000000_adicionar_usuario/down.sql
  - DML/20240607180000001_adicionar_usuario/up.sql
  - DML/20240607180000001_adicionar_usuario/down.sql
```

```sh
# Modo interativo (sem argumentos)
./bin/migrate
Informe o nome para a migration: adicionar_usuario
Deseja gerar DDL? (s/n): s
Deseja gerar DML? (s/n): n

# Resultado (cria apenas dentro de DDL/):
Arquivos criados:
  - DDL/20240607180000000_adicionar_usuario/up.sql
  - DDL/20240607180000000_adicionar_usuario/down.sql
```

```sh
# Apenas DML
./bin/migrate adicionar_usuario --dml
# Resultado (cria dentro de DML/):
Arquivos criados:
  - DML/20240607180000000_adicionar_usuario/up.sql
  - DML/20240607180000000_adicionar_usuario/down.sql
```

```sh
# Apenas DDL
./bin/migrate adicionar_usuario --ddl
# Resultado (cria dentro de DDL/):
Arquivos criados:
  - DDL/20240607180000000_adicionar_usuario/up.sql
  - DDL/20240607180000000_adicionar_usuario/down.sql
```

```sh
# Ambos (DDL e DML) sem perguntas
./bin/migrate adicionar_usuario --all
# Resultado (cria dentro de DDL/ e DML/):
Arquivos criados:
  - DDL/20240607180000000_adicionar_usuario/up.sql
  - DDL/20240607180000000_adicionar_usuario/down.sql
  - DML/20240607180000001_adicionar_usuario/up.sql
  - DML/20240607180000001_adicionar_usuario/down.sql
```

---

