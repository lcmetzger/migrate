package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/charmap"
)

// Tipos e constantes
type MigrationType string

const (
	MigrationTypeDDL     MigrationType = "DDL"
	MigrationTypeDML     MigrationType = "DML"
	MigrationTypeCURRENT MigrationType = "CURRENT"
)

// Estruturas de dados
type MigrationConfig struct {
	Type      MigrationType
	Timestamp string
	Name      string
	SubDir    string
}

type MigrationResult struct {
	UpPath   string
	DownPath string
}

// Template para arquivos DML
const templateDML = `BEGIN
  -- insira aqui seus scripts DML
  -- Não esqueça de retirar todos os comentários
  -- O arquivo está em charset ISO-8859-1, e deve seer enviado nesse charset



  COMMIT;
EXCEPTION WHEN OTHERS THEN
  DBMS_OUTPUT.PUT_LINE(SQLERRM);
  DBMS_OUTPUT.PUT_LINE(SQLCODE);
  ROLLBACK;
END;`

// Template para arquivos DDL
const templateDDL = `-- Cada script DDL deve ser terminado por ";" e abaixo de cada comando inserir uma "/"
-- Não esqueça de retirar todos os comentários
-- O arquivo está em charset ISO-8859-1, e deve seer enviado nesse charset`

// Variáveis para as flags do Cobra
var (
	dmlFlag bool
	ddlFlag bool
	allFlag bool
	subDir  string
)

// rootCmd representa o comando raiz da aplicação.
var rootCmd = &cobra.Command{
	Use:   "migrate [nome]",
	Short: "Helper para criação de migrations",
	Long:  `CLI para facilitar a criação de migrations com timestamp e arquivos SQL.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var migrationName string
		var isInteractive bool

		// Modo interativo é ativado se nenhum argumento for passado.
		if len(args) == 0 {
			isInteractive = true
			interactiveResults, err := runInteractiveMode(cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			migrationName = interactiveResults.Name
			subDir = interactiveResults.SubDir // Sobrescreve a flag --sub
			ddlFlag = interactiveResults.DDL
			dmlFlag = interactiveResults.DML
			allFlag = false // --all não é uma opção no modo interativo
		} else {
			migrationName = args[0]
		}

		if migrationName == "" {
			return fmt.Errorf("nome da migration não pode ser vazio")
		}

		// Determina as configurações com base nas flags.
		configs := determinarConfiguracoes(allFlag, dmlFlag, ddlFlag, migrationName, subDir)

		// Se nenhuma configuração foi gerada (e não estamos no modo interativo), o padrão é criar na pasta atual.
		if len(configs) == 0 && !isInteractive {
			configs = append(configs, MigrationConfig{Type: MigrationTypeCURRENT, Name: migrationName, SubDir: subDir})
		}

		if len(configs) == 0 {
			return fmt.Errorf("nenhum tipo de migration foi selecionado")
		}

		results, err := criarMigrations(configs)
		if err != nil {
			return err
		}

		exibirResultados(cmd.OutOrStdout(), results)
		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVar(&dmlFlag, "dml", false, "Criar scripts DML na pasta DML/")
	rootCmd.Flags().BoolVar(&ddlFlag, "ddl", false, "Criar scripts DDL na pasta DDL/")
	rootCmd.Flags().BoolVar(&allFlag, "all", false, "Criar scripts DDL e DML nas pastas DDL/ e DML/")
	rootCmd.Flags().StringVar(&subDir, "sub", "", "Especifica um subdiretório para a migration")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func criarDiretorio(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório %s: %w", path, err)
	}
	return nil
}

func processarTemplate(template string) string {
	lineEnding := "\n"
	if runtime.GOOS == "windows" {
		lineEnding = "\r\n"
	}

	// Substitui todas as quebras de linha \n por lineEnding apropriado para o sistema
	return strings.ReplaceAll(template, "\n", lineEnding)
}

func criarArquivo(path string, migrationType MigrationType) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo %s: %w", path, err)
	}
	defer f.Close()

	writer := charmap.ISO8859_1.NewEncoder().Writer(f)
	var content string
	switch migrationType {
	case MigrationTypeDML:
		content = processarTemplate(templateDML)
	case MigrationTypeDDL, MigrationTypeCURRENT:
		content = processarTemplate(templateDDL)
	default:
	}

	_, err = writer.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("erro ao escrever no arquivo %s: %w", path, err)
	}
	return nil
}

func criarMigration(config MigrationConfig) (*MigrationResult, error) {
	dirName := fmt.Sprintf("%s_%s", config.Timestamp, config.Name)
	basePath := string(config.Type)

	if config.SubDir != "" && config.Type != MigrationTypeCURRENT {
		basePath = filepath.Join(basePath, config.SubDir)
	}

	var dir string
	switch config.Type {
	case MigrationTypeDDL, MigrationTypeDML:
		dir = filepath.Join(basePath, dirName)
	case MigrationTypeCURRENT:
		dir = dirName
	default:
		return nil, fmt.Errorf("tipo de migration inválido: '%s'", config.Type)
	}

	if err := criarDiretorio(dir); err != nil {
		return nil, err
	}

	upPath := filepath.Join(dir, "up.sql")
	downPath := filepath.Join(dir, "down.sql")

	if err := criarArquivo(upPath, config.Type); err != nil {
		return nil, err
	}
	if err := criarArquivo(downPath, config.Type); err != nil {
		return nil, err
	}

	return &MigrationResult{UpPath: upPath, DownPath: downPath}, nil
}

func criarMigrations(configs []MigrationConfig) ([]MigrationResult, error) {
	results := make([]MigrationResult, 0, len(configs))
	now := time.Now()

	for i, config := range configs {
		// Adiciona um pequeno atraso para garantir um timestamp único para cada arquivo.
		migrationTime := now.Add(time.Duration(i) * time.Millisecond)
		timestamp := migrationTime.Format("20060102150405") + fmt.Sprintf("%03d", migrationTime.Nanosecond()/1e6)
		config.Timestamp = timestamp

		result, err := criarMigration(config)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}

	return results, nil
}

func exibirResultados(writer io.Writer, results []MigrationResult) {
	fmt.Fprintln(writer, "Arquivos criados:")
	for _, result := range results {
		fmt.Fprintln(writer, "  -", filepath.ToSlash(result.UpPath))
		fmt.Fprintln(writer, "  -", filepath.ToSlash(result.DownPath))
	}
}

type interactiveResult struct {
	Name   string
	SubDir string
	DDL    bool
	DML    bool
}

func runInteractiveMode(reader io.Reader, writer io.Writer) (*interactiveResult, error) {
	scanner := bufio.NewScanner(reader)
	result := &interactiveResult{}

	fmt.Fprint(writer, "Informe o nome para a migration: ")
	if !scanner.Scan() {
		return nil, fmt.Errorf("falha ao ler entrada para o nome da migration: %w", scanner.Err())
	}
	result.Name = strings.TrimSpace(scanner.Text())

	fmt.Fprint(writer, "Informe o subdiretório (opcional): ")
	if !scanner.Scan() {
		return nil, fmt.Errorf("falha ao ler entrada para o subdiretório: %w", scanner.Err())
	}
	result.SubDir = strings.TrimSpace(scanner.Text())

	fmt.Fprint(writer, "Deseja gerar DDL? (s/n): ")
	if !scanner.Scan() {
		return nil, fmt.Errorf("falha ao ler entrada para DDL: %w", scanner.Err())
	}
	result.DDL = strings.ToLower(strings.TrimSpace(scanner.Text())) == "s"

	fmt.Fprint(writer, "Deseja gerar DML? (s/n): ")
	if !scanner.Scan() {
		return nil, fmt.Errorf("falha ao ler entrada para DML: %w", scanner.Err())
	}
	result.DML = strings.ToLower(strings.TrimSpace(scanner.Text())) == "s"

	return result, nil
}

func determinarConfiguracoes(useAll, useDML, useDDL bool, name, subDir string) []MigrationConfig {
	var configs []MigrationConfig

	if useAll {
		configs = append(configs, MigrationConfig{Type: MigrationTypeDDL, Name: name, SubDir: subDir})
		configs = append(configs, MigrationConfig{Type: MigrationTypeDML, Name: name, SubDir: subDir})
	} else {
		if useDDL {
			configs = append(configs, MigrationConfig{Type: MigrationTypeDDL, Name: name, SubDir: subDir})
		}
		if useDML {
			configs = append(configs, MigrationConfig{Type: MigrationTypeDML, Name: name, SubDir: subDir})
		}
	}

	return configs
}
