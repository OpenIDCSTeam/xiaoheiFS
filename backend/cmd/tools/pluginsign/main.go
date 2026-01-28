package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type checksums struct {
	Algo  string            `json:"algo"`
	Files map[string]string `json:"files"`
}

func main() {
	dir := flag.String("dir", "", "plugin directory to sign (must contain manifest.json)")
	priv := flag.String("ed25519-priv", "", "base64 ed25519 private key (64 bytes). If empty, generate a new keypair and print it.")
	flag.Parse()

	if strings.TrimSpace(*dir) == "" {
		fmt.Fprintln(os.Stderr, "missing -dir")
		os.Exit(2)
	}
	pluginDir := strings.TrimSpace(*dir)
	if _, err := os.Stat(filepath.Join(pluginDir, "manifest.json")); err != nil {
		fmt.Fprintln(os.Stderr, "manifest.json not found:", err)
		os.Exit(2)
	}

	var privKey ed25519.PrivateKey
	var pubKey ed25519.PublicKey
	if strings.TrimSpace(*priv) == "" {
		pubKey, privKey, _ = ed25519.GenerateKey(rand.Reader)
		fmt.Println("Generated Ed25519 keypair (store the private key securely):")
		fmt.Println("PublicKey(base64):", base64.StdEncoding.EncodeToString(pubKey))
		fmt.Println("PrivateKey(base64):", base64.StdEncoding.EncodeToString(privKey))
		fmt.Println()
	} else {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(*priv))
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid -ed25519-priv:", err)
			os.Exit(2)
		}
		if len(b) != ed25519.PrivateKeySize {
			fmt.Fprintln(os.Stderr, "invalid -ed25519-priv length")
			os.Exit(2)
		}
		privKey = ed25519.PrivateKey(b)
		pubKey = privKey.Public().(ed25519.PublicKey)
		fmt.Println("PublicKey(base64):", base64.StdEncoding.EncodeToString(pubKey))
	}

	cs, err := buildChecksums(pluginDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "checksums:", err)
		os.Exit(1)
	}
	csBytes, _ := json.MarshalIndent(cs, "", "  ")
	if err := os.WriteFile(filepath.Join(pluginDir, "checksums.json"), csBytes, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write checksums.json:", err)
		os.Exit(1)
	}

	sig := ed25519.Sign(privKey, csBytes)
	if err := os.WriteFile(filepath.Join(pluginDir, "signature.sig"), sig, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write signature.sig:", err)
		os.Exit(1)
	}

	fmt.Println("OK: wrote checksums.json + signature.sig")
}

func buildChecksums(pluginDir string) (checksums, error) {
	files := map[string]string{}
	err := filepath.WalkDir(pluginDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.EqualFold(name, "signature.sig") || strings.EqualFold(name, "checksums.json") {
			return nil
		}
		rel, _ := filepath.Rel(pluginDir, path)
		rel = filepath.ToSlash(rel)
		sum, err := sha256File(path)
		if err != nil {
			return err
		}
		files[rel] = sum
		return nil
	})
	if err != nil {
		return checksums{}, err
	}
	// stable ordering in output (for better diffs)
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := map[string]string{}
	for _, k := range keys {
		ordered[k] = files[k]
	}
	return checksums{Algo: "sha256", Files: ordered}, nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
