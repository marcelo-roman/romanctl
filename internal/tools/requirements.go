package tools

import "romanctl/internal/platform"

// requirements diz o que cada ferramenta exige da máquina. O que não está aqui só depende
// do sistema, da arquitetura e do gerenciador de pacotes.
var requirements = map[string][]string{
	// Motores e o que só funciona com um motor local. Clientes como docker-cli, crictl,
	// dive e lazydocker ficam de fora, porque servem para uma máquina remota.
	"docker":          {platform.NeedContainers},
	"podman":          {platform.NeedContainers},
	"podman-compose":  {platform.NeedContainers},
	"buildah":         {platform.NeedContainers},
	"containerd":      {platform.NeedContainers},
	"nerdctl":         {platform.NeedContainers},
	"kind":            {platform.NeedContainers},
	"minikube":        {platform.NeedContainers},
	"k3d":             {platform.NeedContainers},
	"k3s":             {platform.NeedContainers},
	"k0s":             {platform.NeedContainers},
	"colima":          {platform.NeedVM},
	"lima":            {platform.NeedVM},
	"docker-desktop":  {platform.NeedVM, platform.NeedGUI},
	"podman-desktop":  {platform.NeedVM, platform.NeedGUI},
	"orbstack":        {platform.NeedVM, platform.NeedGUI},
	"rancher-desktop": {platform.NeedContainers, platform.NeedGUI},

	// O binário oficial é ligado à glibc: não roda no Alpine.
	"node": {platform.NeedGlibc}, "rust": {platform.NeedGlibc}, "dart": {platform.NeedGlibc}, "swift": {platform.NeedGlibc},
	"deno": {platform.NeedGlibc}, "bun": {platform.NeedGlibc}, "uv": {platform.NeedGlibc}, "delta": {platform.NeedGlibc},
	"eza": {platform.NeedGlibc}, "aws": {platform.NeedGlibc}, "gcloud": {platform.NeedGlibc}, "neovim": {platform.NeedGlibc},
	"ollama": {platform.NeedGlibc}, "code-server": {platform.NeedGlibc},
}
