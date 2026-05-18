package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Invincib-e/gf-wrapper/config/render"
	"github.com/gogf/gf/cmd/gf/v2/gfcmd"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

const cliFolderName = "hack"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := gctx.New()

	if err := injectRenderedConfig(); err != nil {
		return err
	}

	command, err := gfcmd.GetCommand(ctx)
	if err != nil {
		return err
	}
	command.Run(ctx)
	return nil
}

func injectRenderedConfig() error {
	adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	if !ok {
		return fmt.Errorf("unexpected config adapter type %T", g.Cfg().GetAdapter())
	}

	if hackPath, _ := gfile.Search(cliFolderName); hackPath != "" {
		if err := adapter.SetPath(hackPath); err != nil {
			return err
		}
	}

	fileName := adapter.GetFileName()
	configPath, err := adapter.GetFilePath(fileName)
	if err != nil {
		return err
	}
	if configPath == "" {
		return nil
	}

	rawBytes, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config %s: %w", configPath, err)
	}
	raw := string(rawBytes)

	renderer := render.NewRenderer(
		render.EnvResolver{},
		render.FileResolver{},
		render.NewDockerSecretResolver(""),
	)
	rendered, err := renderer.Render(raw)
	if err != nil {
		return fmt.Errorf("render config %s: %w", configPath, err)
	}
	if rendered == raw {
		return nil
	}

	adapter.SetContent(rendered, fileName)
	adapter.SetContent(rendered)

	baseName := filepath.Base(configPath)
	if baseName != "" && baseName != fileName {
		adapter.SetContent(rendered, baseName)
	}

	adapter.Clear()
	return nil
}
