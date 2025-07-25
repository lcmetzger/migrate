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
  ```
- Se nenhuma flag for informada, o programa perguntará interativamente se deseja gerar DDL e/ou DML:
  ```
  Deseja gerar DDL? (s/n):
  Deseja gerar DML? (s/n):
  ```
- Use `--dml` para criar apenas DML, `--ddl` para apenas DDL, ou `--all` para ambos sem perguntas.

### O que acontece ao executar o comando
- Será criada uma subpasta dentro de `DML` e/ou `DDL` com o nome:
  - Para DML: `<timestamp>_dml_<nome_da_migration>`
  - Para DDL: `<timestamp>_ddl_<nome_da_migration>`
- Dentro de cada subpasta, serão criados os arquivos:
  - `up.sql`
  - `down.sql`
- Os arquivos são criados com encoding ISO-8859-1
- Os arquivos criados contêm o seguinte comentário no início:
  - `-- Não esqueça de excluir este comentário, e verifique se o seu editor está definido para utilizar o encoding ISO-8859-1`
- Ao final, será exibida a lista dos arquivos criados:
  ```
  Arquivos criados:
    - DML/20240607180000000_dml_adicionar_usuario/up.sql
    - DML/20240607180000000_dml_adicionar_usuario/down.sql
    - DDL/20240607180000000_ddl_adicionar_usuario/up.sql
    - DDL/20240607180000000_ddl_adicionar_usuario/down.sql
  ```

### Exemplos de uso

```sh
# Modo totalmente interativo (nome e tipo)
./bin/migrate
Informe o nome para a migration: adicionar_usuario
Deseja gerar DDL? (s/n): s
Deseja gerar DML? (s/n): s

# Resultado:
Arquivos criados:
  - DML/20240607180000000_dml_adicionar_usuario/up.sql
  - DML/20240607180000000_dml_adicionar_usuario/down.sql
  - DDL/20240607180000000_ddl_adicionar_usuario/up.sql
  - DDL/20240607180000000_ddl_adicionar_usuario/down.sql
```

```sh
# Apenas DML
./bin/migrate adicionar_usuario --dml
# Resultado:
Arquivos criados:
  - DML/20240607180000000_dml_adicionar_usuario/up.sql
  - DML/20240607180000000_dml_adicionar_usuario/down.sql
```

```sh
# Apenas DDL
./bin/migrate adicionar_usuario --ddl
# Resultado:
Arquivos criados:
  - DDL/20240607180000000_ddl_adicionar_usuario/up.sql
  - DDL/20240607180000000_ddl_adicionar_usuario/down.sql
```

```sh
# Ambos (DDL e DML) sem perguntas
./bin/migrate adicionar_usuario --all
# Resultado:
Arquivos criados:
  - DML/20240607180000000_dml_adicionar_usuario/up.sql
  - DML/20240607180000000_dml_adicionar_usuario/down.sql
  - DDL/20240607180000000_ddl_adicionar_usuario/up.sql
  - DDL/20240607180000000_ddl_adicionar_usuario/down.sql
```

---

