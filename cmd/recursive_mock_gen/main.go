package main

import (
	"context"
	"fmt"
	"github.com/SHOWROOM-inc/recursive_mock_gen/models"
	"github.com/SHOWROOM-inc/recursive_mock_gen/repositories"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"sync"
)

const input = "input"
const output = "output"
const cache = "cache"
const maxParallels = "max-parallels"

func main() {
	app := cli.NewApp()
	app.Name = "Recursive mock generator"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     input,
			Usage:    "Path to input root directory",
			Required: false,
			Value:    ".",
		},
		&cli.StringFlag{
			Name:     "output",
			Usage:    "Path to output root",
			Required: true,
		},
		&cli.StringFlag{
			Name:     cache,
			Usage:    "Path to cache file",
			Required: false,
			Value:    ".mockgen-cache.json",
		},
		&cli.Uint64Flag{
			Name:     maxParallels,
			Usage:    "Specify max number of parallels",
			Required: false,
			Value:    100,
		},
	}
	app.Action = action

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func action(c *cli.Context) error {
	cacheFilePath := c.String(cache)
	inputRootDir := c.String(input)
	outputRootDir := c.String(output)

	cacheRepo := repositories.NewCacheRepository()
	codeRepo := repositories.NewCodeRepository()

	prevEntries, err := cacheRepo.ReadCache(cacheFilePath)

	newEntries, err := codeRepo.LoadInterfaces(inputRootDir)
	if err != nil {
		panic(err)
	}

	changed, deleted := prevEntries.Diff(newEntries)

	// mockgenの実行
	runMockGenParallel(c.Context, changed, outputRootDir, c.Int64(maxParallels))

	// 不要になったモックファイルの削除
	for _, e := range deleted {
		fp := e.OutFilePath(outputRootDir)
		fmt.Println("delete ", fp)
		if err := os.Remove(fp); err != nil {
			fmt.Println("failed to delete ", e.FilePath)
		}
	}

	// キャッシュファイルの保存
	if err := cacheRepo.WriteCache(cacheFilePath, newEntries); err != nil {
		return err
	}

	return nil
}

func runMockGenParallel(ctx context.Context, target models.Cache, outputBasePath string, maxParallel int64) {
	var wg sync.WaitGroup

	// goroutine数に制約を持たせるためのチャンネルを作成
	ch := make(chan bool, maxParallel)

	for key, entry := range target {
		// 最大のgoroutine数制限のためにチャンネルのバッファを確保
		ch <- true

		wg.Add(1)
		go func(e models.Entry) {
			if err := runMockGen(ctx, e, outputBasePath); err != nil {
				fmt.Println(fmt.Sprintf("mockgen failed for %s", key), err)
			}
			// 処理終了でチャンネルのバッファを開放
			<-ch
			// goroutine終了
			wg.Done()
		}(entry)
	}
	wg.Wait()
}

func runMockGen(ctx context.Context, e models.Entry, outputBasePath string) error {
	cmd := exec.CommandContext(
		ctx,
		"mockgen",
		"-source", e.FilePath,
		"-destination", e.OutFilePath(outputBasePath),
		"-package", e.OutPackageName(),
	)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	fmt.Println("mockgen for ", e.PkgPath)
	return cmd.Run()
}
