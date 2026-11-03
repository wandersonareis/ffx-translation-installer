// Package service reúne os serviços de domínio do instalador expostos ao
// frontend via Wails (application.NewService).
package service

import (
	"os"
	"path/filepath"
	"strings"

	"ffxinstaller/backend/loggingService"

	"github.com/ApolloF/gamekit/steam"
)

// GameName é o nome do único jogo suportado pelo instalador. É isso que
// aparece em logs e mensagens: o appId do Steam é detalhe interno de busca.
const GameName = "FINAL FANTASY X/X-2 HD Remaster"

// ffxAppID é o appId Steam do FINAL FANTASY X/X-2 HD Remaster. Uso interno:
// só serve para achar o jogo nos manifests do Steam, nunca deve aparecer em
// mensagem para o usuário.
const ffxAppID = 359870

// gameDataFiles são os dados obrigatórios dentro da pasta do jogo: é neles
// que a tradução será aplicada. Se algum não existir, a pasta é inválida.
var gameDataFiles = []string{
	filepath.Join("data", "FFX_Data.vbf"),
	filepath.Join("data", "FFX2_Data.vbf"),
}

// GameService localiza a instalação do FINAL FANTASY X/X-2 HD Remaster no
// Steam usando o pacote gamekit/steam (lê o registro do Windows, as
// bibliotecas e os manifests).
//
// Substituiu o pacote monguse, cuja busca por appId não funcionava: a função
// monguse comparava o id com `string(appId)`, que no Go converte o inteiro
// para um rune (um caractere), e nunca casava com o id em texto — nenhum jogo
// era encontrado. O gamekit compara os ids como texto.
type GameService struct{}

// NewGameService cria um novo serviço de jogos.
func NewGameService() *GameService {
	loggingService.Init()
	return &GameService{}
}

// GetGamePath retorna a pasta de instalação do FINAL FANTASY X/X-2 HD
// Remaster — o único jogo suportado. O frontend concatena FFX.exe para formar
// o caminho do executável do modal. "" se o jogo não estiver instalado: ter
// manifest com caminho não garante instalação, só o gamekit com Installed
// true confirma.
func (s *GameService) GetGamePath() string {
	dir := lookup()
	if dir == "" {
		loggingService.Warn("GameService: %s não encontrado nas bibliotecas do Steam", GameName)
		return ""
	}
	loggingService.Info("GameService: %s localizado em %s", GameName, dir)
	return dir
}

// IsValid informa se a pasta do jogo contém os dados obrigatórios
// (data\FFX_Data.vbf e data\FFX2_Data.vbf). Se algum não existir, a pasta é
// inválida — mesmo com manifest e executáveis presentes. É a validação usada
// pelo frontend para o valor exibido no campo do executável.
func (s *GameService) IsValid(dir string) bool {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		loggingService.Warn("GameService: validação sem diretório informado para %s", GameName)
		return false
	}
	if !isDir(dir) {
		loggingService.Warn("GameService: pasta de %s não encontrada: %s", GameName, dir)
		return false
	}
	valid := true
	for _, rel := range gameDataFiles {
		if !fileExists(filepath.Join(dir, rel)) {
			loggingService.Warn("GameService: dado obrigatório de %s ausente em %s: %s", GameName, dir, rel)
			valid = false
		}
	}
	return valid
}

// lookup devolve a pasta do jogo ou "" se não houver manifest. Sem Installed
// (gamekit) a pasta não é considerada encontrada: manifest com caminho pode
// existir para um jogo desinstalado/incompleto.
func lookup() string {
	root := steam.Dir()
	if root == "" {
		return ""
	}
	app, ok := steam.Apps(root)[ffxAppID]
	if !ok || !app.Installed || app.Dir == "" {
		return ""
	}
	return app.Dir
}

// fileExists informa se o caminho existe e é um arquivo.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// isDir informa se o caminho existe e é um diretório.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
