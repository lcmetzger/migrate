package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/charmap"
)

// MigrationType representa o tipo de migration
type MigrationType string

const (
	MigrationTypeDDL     MigrationType = "DDL"
	MigrationTypeDML     MigrationType = "DML"
	MigrationTypeCURRENT MigrationType = "CURRENT"
)

// MigrationConfig contém a configuração para criar uma migration
type MigrationConfig struct {
	Type      MigrationType
	Timestamp string
	Name      string
	BasePath  string
}

// MigrationResult contém o resultado da criação de uma migration
type MigrationResult struct {
	UpPath   string
	DownPath string
}

var (
	dmlFlag bool
	ddlFlag bool
	allFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "migrate [nome]",
	Short: "Helper para criação de migrations",
	Long:  `CLI para facilitar a criação de migrations com timestamp e arquivos SQL.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nome := obterNomeMigration(args)
		configs := determinarConfiguracoes(allFlag, dmlFlag, ddlFlag, nome)

		if len(configs) == 0 {
			fmt.Println("Nenhum tipo de migration selecionado.")
			os.Exit(1)
		}

		results := criarMigrations(configs)
		exibirResultados(results)
	},
}

func init() {
	rootCmd.Flags().BoolVar(&dmlFlag, "dml", false, "Criar scripts DML na pasta DML/")
	rootCmd.Flags().BoolVar(&ddlFlag, "ddl", false, "Criar scripts DDL na pasta DDL/")
	rootCmd.Flags().BoolVar(&allFlag, "all", false, "Criar scripts DDL e DML nas pastas DDL/ e DML/")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// criarDiretorio cria um diretório se não existir
func criarDiretorio(path string) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("Erro ao criar diretório %s: %v\n", path, err)
		os.Exit(1)
	}
}

// obterLineEnding retorna a quebra de linha apropriada para o sistema operacional
func obterLineEnding() string {
	if runtime.GOOS == "windows" {
		return "\r\n" // CRLF para Windows
	}
	return "\n" // LF para Unix/Linux/macOS
}

// criarArquivo cria um arquivo com o comentário padrão
func criarArquivo(path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Erro ao criar arquivo %s: %v\n", path, err)
		os.Exit(1)
	}
	defer f.Close()

	writer := charmap.ISO8859_1.NewEncoder().Writer(f)
	comment := "-- Não esqueça de excluir este comentário, e verifique se o seu editor está definido para utilizar o encoding ISO-8859-1" + obterLineEnding()
	writer.Write([]byte(comment))
}

// criarMigration cria uma migration baseada na configuração
func criarMigration(config MigrationConfig) MigrationResult {
	dirName := fmt.Sprintf("%s_%s", config.Timestamp, config.Name)

	var dir string
	switch config.Type {
	case MigrationTypeDDL:
		criarDiretorio("DDL")
		dir = filepath.Join("DDL", dirName)
	case MigrationTypeDML:
		criarDiretorio("DML")
		dir = filepath.Join("DML", dirName)
	case MigrationTypeCURRENT:
		dir = dirName
	default:
		fmt.Printf("Erro: tipo de migration inválido '%s'\n", config.Type)
		os.Exit(1)
	}

	criarDiretorio(dir)
	upPath := filepath.Join(dir, "up.sql")
	downPath := filepath.Join(dir, "down.sql")

	criarArquivo(upPath)
	criarArquivo(downPath)

	return MigrationResult{
		UpPath:   upPath,
		DownPath: downPath,
	}
}

// criarMigrations cria múltiplas migrations baseadas nas configurações
func criarMigrations(configs []MigrationConfig) []MigrationResult {
	results := make([]MigrationResult, 0, len(configs))

	now := time.Now()
	for _, config := range configs {
		// Gera timestamp único para cada migration
		timestamp := now.Format("20060102150405") + fmt.Sprintf("%03d", now.Nanosecond()/1e6)
		config.Timestamp = timestamp

		result := criarMigration(config)
		results = append(results, result)

		// Incrementa o tempo para garantir timestamps únicos
		now = now.Add(time.Millisecond)
	}

	return results
}

// exibirResultados exibe a lista de arquivos criados
func exibirResultados(results []MigrationResult) {
	fmt.Println("Arquivos criados:")
	for _, result := range results {
		fmt.Println("  -", result.UpPath)
		fmt.Println("  -", result.DownPath)
	}
}

// perguntarTiposMigration pergunta interativamente sobre DDL e DML
func perguntarTiposMigration() (bool, bool) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Deseja gerar DDL? (s/n): ")
	resposta, _ := reader.ReadString('\n')
	resposta = strings.ToLower(strings.TrimSpace(resposta))
	ddl := (resposta == "s" || resposta == "sim")

	fmt.Print("Deseja gerar DML? (s/n): ")
	resposta, _ = reader.ReadString('\n')
	resposta = strings.ToLower(strings.TrimSpace(resposta))
	dml := (resposta == "s" || resposta == "sim")

	return ddl, dml
}

// determinarConfiguracoes determina as configurações de migration baseadas nas flags
func determinarConfiguracoes(allFlag, dmlFlag, ddlFlag bool, nome string) []MigrationConfig {
	var configs []MigrationConfig

	if allFlag {
		configs = append(configs,
			MigrationConfig{Type: MigrationTypeDDL, Name: nome},
			MigrationConfig{Type: MigrationTypeDML, Name: nome},
		)
	} else if dmlFlag || ddlFlag {
		if ddlFlag {
			configs = append(configs, MigrationConfig{Type: MigrationTypeDDL, Name: nome})
		}
		if dmlFlag {
			configs = append(configs, MigrationConfig{Type: MigrationTypeDML, Name: nome})
		}
	} else {
		configs = append(configs, MigrationConfig{Type: MigrationTypeCURRENT, Name: nome})
	}

	return configs
}

// obterNomeMigration obtém o nome da migration (interativo ou via argumento)
func obterNomeMigration(args []string) string {
	var nome string
	if len(args) == 0 {
		// Modo interativo - pergunta o nome e os tipos
		fmt.Print("Informe o nome para a migration: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		nome = strings.TrimSpace(scanner.Text())
		if nome == "" {
			fmt.Println("Nome não pode ser vazio.")
			os.Exit(1)
		}

		// Pergunta sobre DDL e DML
		ddl, dml := perguntarTiposMigration()

		// Atualiza as flags globais baseado nas respostas
		if ddl {
			ddlFlag = true
		}
		if dml {
			dmlFlag = true
		}
	} else {
		nome = args[0]
	}
	return nome
}
