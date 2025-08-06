package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// resetFlags redefine as flags globais e o estado para cada teste.
func resetFlags() {
	dmlFlag = false
	ddlFlag = false
	allFlag = false
	subDir = ""
	// Restaura o stdin, stdout, stderr originais
	rootCmd.SetIn(nil)
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
}

// setupTestEnvironment cria um diretório temporário e muda o diretório de trabalho para ele.
func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "migration-test-*")
	if err != nil {
		t.Fatalf("Falha ao criar diretório temporário: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Falha ao obter o diretório de trabalho: %v", err)
	}

	os.Chdir(tempDir)

	// Retorna uma função de limpeza para restaurar o estado original.
	return tempDir, func() {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}
}

// checkFiles verifica se a estrutura de arquivos criada corresponde ao esperado.
func checkFiles(t *testing.T, baseDir string, expectedFiles []string) {
	t.Helper()

	// Regex para encontrar e substituir o timestamp na verificação.
	re := regexp.MustCompile(`\d{17,}_`)

	foundFiles := make(map[string]bool)
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}
			// Normaliza o caminho removendo o timestamp e usando barras (/) como separador.
			normalizedPath := re.ReplaceAllString(relPath, "{timestamp}_")
			foundFiles[filepath.ToSlash(normalizedPath)] = true
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Erro ao percorrer o diretório de teste: %v", err)
	}

	if len(expectedFiles) != len(foundFiles) {
		t.Errorf("Número de arquivos incorreto. Esperado: %d, Encontrado: %d", len(expectedFiles), len(foundFiles))
		var foundPaths []string
		for k := range foundFiles {
			foundPaths = append(foundPaths, k)
		}
		t.Logf("Arquivos encontrados: %v", foundPaths)
	}

	expectedMap := make(map[string]bool)
	for _, f := range expectedFiles {
		expectedMap[filepath.ToSlash(f)] = true
	}

	for expected := range expectedMap {
		if !foundFiles[expected] {
			t.Errorf("Arquivo esperado não foi encontrado: %s", expected)
		}
	}

	for found := range foundFiles {
		if !expectedMap[found] {
			t.Errorf("Arquivo inesperado foi encontrado: %s", found)
		}
	}
}

// executeCommand é um helper para executar o comando rootCmd com args e stdin.
func executeCommand(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()

	resetFlags()

	var outBuf, errBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)

	if stdin != "" {
		rootCmd.SetIn(strings.NewReader(stdin))
	}

	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	return outBuf.String(), errBuf.String(), err
}

func TestCreateMigration(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedFiles []string
		expectError   bool
	}{
		{
			name:          "Criação DDL",
			args:          []string{"criar_tabela_usuarios", "--ddl"},
			expectedFiles: []string{"DDL/{timestamp}_criar_tabela_usuarios/up.sql", "DDL/{timestamp}_criar_tabela_usuarios/down.sql"},
		},
		{
			name:          "Criação DML",
			args:          []string{"inserir_usuario_padrao", "--dml"},
			expectedFiles: []string{"DML/{timestamp}_inserir_usuario_padrao/up.sql", "DML/{timestamp}_inserir_usuario_padrao/down.sql"},
		},
		{
			name: "Criação com --all",
			args: []string{"setup_inicial", "--all"},
			expectedFiles: []string{
				"DDL/{timestamp}_setup_inicial/up.sql", "DDL/{timestamp}_setup_inicial/down.sql",
				"DML/{timestamp}_setup_inicial/up.sql", "DML/{timestamp}_setup_inicial/down.sql",
			},
		},
		{
			name:          "Criação com subdiretório",
			args:          []string{"criar_indices", "--ddl", "--sub", "performance"},
			expectedFiles: []string{"DDL/performance/{timestamp}_criar_indices/up.sql", "DDL/performance/{timestamp}_criar_indices/down.sql"},
		},
		{
			name:          "Criação sem flags (diretório atual)",
			args:          []string{"migracao_local"},
			expectedFiles: []string{"{timestamp}_migracao_local/up.sql", "{timestamp}_migracao_local/down.sql"},
		},
		{
			name:        "Sem nome de migration",
			args:        []string{"--ddl"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, cleanup := setupTestEnvironment(t)
			defer cleanup()

			_, _, err := executeCommand(t, "", tc.args...)

			if tc.expectError {
				if err == nil {
					t.Error("Esperado um erro, mas nenhum ocorreu.")
				}
			} else if err != nil {
				t.Errorf("Erro inesperado: %v", err)
			}

			checkFiles(t, tempDir, tc.expectedFiles)
		})
	}
}

func TestInteractiveMode(t *testing.T) {
	t.Run("Modo interativo completo", func(t *testing.T) {
		tempDir, cleanup := setupTestEnvironment(t)
		defer cleanup()

		userInput := "teste_interativo\nsub_dir\ns\ns\n"
		// Para o modo interativo, não passamos argumentos.
		_, _, err := executeCommand(t, userInput, []string{}...)

		if err != nil {
			t.Fatalf("Erro inesperado no modo interativo: %v", err)
		}

		expected := []string{
			"DDL/sub_dir/{timestamp}_teste_interativo/up.sql",
			"DDL/sub_dir/{timestamp}_teste_interativo/down.sql",
			"DML/sub_dir/{timestamp}_teste_interativo/up.sql",
			"DML/sub_dir/{timestamp}_teste_interativo/down.sql",
		}

		checkFiles(t, tempDir, expected)
	})
}
