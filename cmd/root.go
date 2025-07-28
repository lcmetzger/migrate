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

		// Determina quais roots devem ser processadas e se deve criar pastas
		roots, criarPastas := determinarRoots(allFlag, dmlFlag, ddlFlag)

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
			criarArquivosNaRaiz(root, timestamps[root], nome, &arquivos, criarPastas)
		}

		fmt.Println("Arquivos criados:")
		for _, arq := range arquivos {
			fmt.Println("  -", arq)
		}
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

// Função para criar os arquivos up.sql e down.sql em uma raiz (DML/DDL/CURRENT)
func criarArquivosNaRaiz(root, timestamp, nome string, arquivos *[]string, criarPastas bool) {
	switch root {
	case "DML":
		// Cria dentro da pasta DML/
		dirName := fmt.Sprintf("%s_%s", timestamp, nome)
		criarDiretorio("DML")
		dir := filepath.Join("DML", dirName)
		criarDiretorio(dir)
		upPath := filepath.Join(dir, "up.sql")
		downPath := filepath.Join(dir, "down.sql")
		criarArquivo(upPath)
		criarArquivo(downPath)
		*arquivos = append(*arquivos, upPath, downPath)
	case "DDL":
		// Cria dentro da pasta DDL/
		dirName := fmt.Sprintf("%s_%s", timestamp, nome)
		criarDiretorio("DDL")
		dir := filepath.Join("DDL", dirName)
		criarDiretorio(dir)
		upPath := filepath.Join(dir, "up.sql")
		downPath := filepath.Join(dir, "down.sql")
		criarArquivo(upPath)
		criarArquivo(downPath)
		*arquivos = append(*arquivos, upPath, downPath)
	case "CURRENT":
		// Cria na pasta atual
		dirName := fmt.Sprintf("%s_%s", timestamp, nome)
		dir := dirName
		criarDiretorio(dir)
		upPath := filepath.Join(dir, "up.sql")
		downPath := filepath.Join(dir, "down.sql")
		criarArquivo(upPath)
		criarArquivo(downPath)
		*arquivos = append(*arquivos, upPath, downPath)
	default:
		fmt.Printf("Erro: root inválido '%s'. Deve ser 'DML', 'DDL' ou 'CURRENT'.\n", root)
		os.Exit(1)
	}
}

// Função para perguntar interativamente sobre DDL, DML e criação de pastas
func perguntarTiposMigration() (bool, bool) {
	var ddl, dml bool
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Deseja gerar DDL? (s/n): ")
	resposta, _ := reader.ReadString('\n')
	resposta = strings.ToLower(strings.TrimSpace(resposta))
	ddl = (resposta == "s" || resposta == "sim")

	fmt.Print("Deseja gerar DML? (s/n): ")
	resposta, _ = reader.ReadString('\n')
	resposta = strings.ToLower(strings.TrimSpace(resposta))
	dml = (resposta == "s" || resposta == "sim")

	return ddl, dml
}

// Função para determinar quais roots devem ser processadas
func determinarRoots(allFlag, dmlFlag, ddlFlag bool) ([]string, bool) {
	var roots []string
	var criarPastas bool

	if allFlag {
		roots = []string{"DDL", "DML"}
		criarPastas = true // DDL e DML sempre vão para suas respectivas pastas
	} else if dmlFlag || ddlFlag {
		// Flags específicas informadas (incluindo modo interativo)
		if ddlFlag {
			roots = append(roots, "DDL")
		}
		if dmlFlag {
			roots = append(roots, "DML")
		}
		// Se flags específicas foram informadas, sempre criar nas pastas DDL/DML
		criarPastas = true
	} else {
		// Nenhuma flag informada - criar na pasta corrente
		roots = []string{"CURRENT"}
		criarPastas = false
	}

	return roots, criarPastas
}

// Função para obter o nome da migration (interativo ou via argumento)
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
