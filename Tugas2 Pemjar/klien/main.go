package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	alamat := "localhost:200"

	conn, err := net.Dial("tcp", alamat)
	if err != nil {
		fmt.Println("gagal tersambung ke server:", err)
		return
	}

	defer conn.Close()

	fmt.Println("berhasil terhubung ke", alamat)
	fmt.Println("Ketik pesan lalu tekan Enter.")
	fmt.Println("Ketik KELUAR untuk berhenti.")

	go func() {
		dariServer := bufio.NewScanner(conn)
		for dariServer.Scan() {
			fmt.Println(dariServer.Text())
		}

		fmt.Println("server menutup koneksi")
		os.Exit(0)
	}()

	keyboard := bufio.NewScanner(os.Stdin)

	for keyboard.Scan() {
		pesan := strings.TrimSpace(keyboard.Text())

		if strings.ToUpper(pesan) == "KELUAR" {
			return
		}

		_, err := fmt.Fprintf(conn, "%s\n", pesan)
		if err != nil {
			fmt.Println("gagal mengirim:", err)
			return
		}
	}
}
