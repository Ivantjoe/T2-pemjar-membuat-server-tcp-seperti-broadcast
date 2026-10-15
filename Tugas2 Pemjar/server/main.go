package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type klien struct {
	nama   string
	keluar chan string
}

type ruang struct {
	mu     sync.Mutex
	daftar map[*klien]bool
}

func (r *ruang) tambah(k *klien) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.daftar[k] = true
}

func (r *ruang) hapus(k *klien) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.daftar[k] {
		delete(r.daftar, k)
		close(k.keluar)
	}
}

func (r *ruang) siarkan(pengirim *klien, pesan string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k := range r.daftar {
		if k == pengirim {
			continue
		}

		select {
		case k.keluar <- pesan:
		default:
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":200")
	if err != nil {
		log.Fatal("gagal membuka port 200:", err)
	}

	defer ln.Close()

	log.Println("server broadcast berjalan di port 200")

	r := &ruang{daftar: make(map[*klien]bool)}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("gagal menerima koneksi:", err)
			continue
		}

		go layani(conn, r)
	}
}

func layani(conn net.Conn, r *ruang) {
	defer conn.Close()

	alamat := conn.RemoteAddr().String()
	log.Println("client terhubung:", alamat)

	scanner := bufio.NewScanner(conn)
	fmt.Fprintln(conn, "Masukkan nama Anda:")
	if !scanner.Scan() {
		return
	}

	nama := strings.TrimSpace(scanner.Text())
	if nama == "" {
		nama = alamat
	}

	k := &klien{nama: nama, keluar: make(chan string, 16)}
	selesai := make(chan struct{})
	go func() {
		defer close(selesai)
		for pesan := range k.keluar {
			if _, err := fmt.Fprintln(conn, pesan); err != nil {
				return
			}
		}
	}()

	r.tambah(k)
	fmt.Fprintf(conn, "Selamat datang, %s! Pesan Anda akan dilihat semua client lain.\n", nama)
	r.siarkan(k, fmt.Sprintf("*** %s bergabung ***", nama))

	
	for scanner.Scan() {
		pesan := strings.TrimSpace(scanner.Text())
		if pesan == "" {
			continue
		}

		log.Printf("dari %s (%s): %q", nama, alamat, pesan)
		r.siarkan(k, fmt.Sprintf("%s: %s", nama, pesan))
	}

	r.hapus(k)
	<-selesai

	r.siarkan(k, fmt.Sprintf("*** %s keluar ***", nama))
	log.Println("client terputus:", alamat)
}
