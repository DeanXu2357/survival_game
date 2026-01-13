package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"survival/internal/adapters/handler/websocket"

	"github.com/spf13/cobra"
)

var port string

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Start the WebSocket game server",
	Long:  `Start the WebSocket game server for handling multiplayer connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv := websocket.NewServer(ctx, port)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			log.Printf("Starting server on port %s...", port)
			if err := srv.Start(); err != nil {
				log.Fatalf("Server failed: %v", err)
			}
		}()

		<-sigChan
		log.Println("Received shutdown signal")
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}

		log.Println("Server shut down successfully")
	},
}

func init() {
	rootCmd.AddCommand(backendCmd)
	backendCmd.Flags().StringVarP(&port, "port", "p", "3033", "Port to run the server on")
}
