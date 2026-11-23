package tools

import "romanctl/internal/platform"

// requirements says what each tool demands from the machine. Anything not listed here depends only
// on the OS, the architecture and the package manager.
var requirements = map[string][]string{
	// Engines and what only works with a local engine. Clients like docker-cli, crictl,
	// dive and lazydocker are left out, because they serve a remote machine.
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

	// The official binary links against glibc: it does not run on Alpine.
	"node": {platform.NeedGlibc}, "rust": {platform.NeedGlibc}, "dart": {platform.NeedGlibc}, "swift": {platform.NeedGlibc},
	"deno": {platform.NeedGlibc}, "bun": {platform.NeedGlibc}, "uv": {platform.NeedGlibc}, "delta": {platform.NeedGlibc},
	"eza": {platform.NeedGlibc}, "aws": {platform.NeedGlibc}, "gcloud": {platform.NeedGlibc}, "neovim": {platform.NeedGlibc},
	"ollama": {platform.NeedGlibc}, "code-server": {platform.NeedGlibc},
}
