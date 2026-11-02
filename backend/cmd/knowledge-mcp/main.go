package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"knowledgegraph/backend/internal/content"
	"knowledgegraph/backend/internal/knowledgemcp"
)

func main() {
	root := flag.String("root", "./var/documents", "existing owner-readable document directory")
	workspaceURL := flag.String("workspace-url", "https://knowledge.henryweb.site", "authenticated workspace URL for source citations")
	check := flag.Bool("check", false, "validate read access and exit without starting MCP")
	flag.Parse()
	if *check {
		reader, err := content.OpenWorkspaceReader(*root)
		if err != nil {
			log.Fatal(err)
		}
		documents, err := reader.List("zh-CN")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Read-only knowledge store ready: %d documents", len(documents))
		return
	}
	server, err := knowledgemcp.New(*root, *workspaceURL)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err = server.Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}
