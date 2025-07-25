package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/charmap"
)

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

		// Determina quais roots devem ser processadas
		roots := determinarRoots(allFlag, dmlFlag, ddlFlag)

		if len(roots) == 0 {
			fmt.Println("Nenhum tipo de migration selecionado.")
			os.Exit(1)
		}

		arquivos := []string{}
		timestamps := make(map[string]string)
		now := time.Now()
		for _, root := range roots {
			ts := now.Format("20060102150405") + fmt.Sprintf("%03d", now.Nanosecond()/1e6)
			timestamps[root] = ts
			now = now.Add(time.Millisecond)
		}
		for _, root := range roots {
			criarArquivosNaRaiz(root, timestamps[root], nome, &arquivos)
		}

		fmt.Println("Arquivos criados:")
		for _, arq := range arquivos {
			fmt.Println("  -", arq)
		}
	},
}

func init() {
	rootCmd.Flags().BoolVar(&dmlFlag, "dml", false, "Criar arquivo dml.sql")
	rootCmd.Flags().BoolVar(&ddlFlag, "ddl", false, "Criar arquivo ddl.sql")
	rootCmd.Flags().BoolVar(&allFlag, "all", false, "Criar arquivos up.sql e down.sql (padrão)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// Função para criar diretório se não existir
func criarDiretorio(path string) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("Erro ao criar diretório %s: %v\n", path, err)
		os.Exit(1)
	}
}

// Função para criar arquivo com o comentário padrão
func criarArquivo(path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Erro ao criar arquivo %s: %v\n", path, err)
		os.Exit(1)
	}
	writer := charmap.ISO8859_1.NewEncoder().Writer(f)
	writer.Write([]byte("-- Não esqueça de excluir este comentário, e verifique se o seu editor está definido para utilizar o encoding ISO-8859-1\n"))
	f.Close()
}

// Função para criar os arquivos up.sql e down.sql em uma raiz (DML/DDL)
func criarArquivosNaRaiz(root, timestamp, nome string, arquivos *[]string) {
	var sufixo string
	switch root {
	case "DML":
		sufixo = "dml"
	case "DDL":
		sufixo = "ddl"
	default:
		fmt.Printf("Erro: root inválido '%s'. Deve ser 'DML' ou 'DDL'.\n", root)
		os.Exit(1)
	}

	dirName := fmt.Sprintf("%s_%s_%s", timestamp, sufixo, nome)
	dir := filepath.Join(root, dirName)
	criarDiretorio(root)
	criarDiretorio(dir)
	upPath := filepath.Join(dir, "up.sql")
	downPath := filepath.Join(dir, "down.sql")
	criarArquivo(upPath)
	criarArquivo(downPath)
	*arquivos = append(*arquivos, upPath, downPath)
}

// Função para perguntar interativamente sobre DDL e DML
func perguntarTiposMigration() (bool, bool) {
	var ddl, dml bool

	fmt.Print("Deseja gerar DDL? (s/n): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	resposta := strings.ToLower(strings.TrimSpace(scanner.Text()))
	ddl = (resposta == "s" || resposta == "sim")

	fmt.Print("Deseja gerar DML? (s/n): ")
	scanner.Scan()
	resposta = strings.ToLower(strings.TrimSpace(scanner.Text()))
	dml = (resposta == "s" || resposta == "sim")

	return ddl, dml
}

// Função para determinar quais roots devem ser processadas
func determinarRoots(allFlag, dmlFlag, ddlFlag bool) []string {
	var roots []string

	if allFlag {
		roots = []string{"DDL", "DML"}
	} else if !dmlFlag && !ddlFlag {
		// Modo interativo
		ddl, dml := perguntarTiposMigration()
		if ddl {
			roots = append(roots, "DDL")
		}
		if dml {
			roots = append(roots, "DML")
		}
	} else {
		// Flags específicas informadas
		if ddlFlag {
			roots = append(roots, "DDL")
		}
		if dmlFlag {
			roots = append(roots, "DML")
		}
	}

	return roots
}

// Função para obter o nome da migration (interativo ou via argumento)
func obterNomeMigration(args []string) string {
	var nome string
	if len(args) == 0 {
		fmt.Print("Informe o nome para a migration: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		nome = strings.TrimSpace(scanner.Text())
		if nome == "" {
			fmt.Println("Nome não pode ser vazio.")
			os.Exit(1)
		}
	} else {
		nome = args[0]
	}
	return nome
}
