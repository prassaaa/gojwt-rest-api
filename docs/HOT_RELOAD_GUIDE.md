# Hot Reload dengan Air

Air adalah live reload tool untuk Go yang akan otomatis rebuild dan restart aplikasi ketika ada perubahan file.

## 🚀 Cara Menggunakan Air

### 1. **Pastikan Air sudah terinstall**

Air sudah terinstall di `~/go/bin/air`. Pastikan `~/go/bin` ada di PATH Anda.

Check dengan:
```bash
~/go/bin/air -v
```

### 2. **Tambahkan ~/go/bin ke PATH (jika belum)**

Edit file `~/.bashrc` atau `~/.zshrc`:
```bash
export PATH=$PATH:~/go/bin
```

Kemudian reload:
```bash
source ~/.bashrc
```

### 3. **Jalankan dengan Air**

Di root project, jalankan:
```bash
~/go/bin/air
```

Atau jika sudah di PATH:
```bash
air
```

Atau menggunakan Makefile:
```bash
make dev
```

## 📋 Apa yang Terjadi?

Ketika Air running:

1. **Auto Build** - Air akan build aplikasi ke `tmp/main`
2. **Auto Run** - Aplikasi akan dijalankan otomatis
3. **Watch Changes** - Air akan watch semua file `.go`
4. **Auto Reload** - Jika ada perubahan, Air akan:
   - Stop aplikasi yang running
   - Rebuild code
   - Restart aplikasi

## 🎯 Konsep Hot Reload

```
┌─────────────────────────────────────────────┐
│  1. Anda edit file .go                      │
│  2. Save file                               │
│  3. Air detect perubahan                    │
│  4. Air kill process yang running           │
│  5. Air rebuild aplikasi                    │
│  6. Air jalankan binary baru                │
│  7. Server running dengan code baru         │
│  8. Kembali ke step 1                       │
└─────────────────────────────────────────────┘
```

## 🧪 Testing Hot Reload

### Step 1: Jalankan Air
```bash
~/go/bin/air
```

Output:
```
  __    _   ___
 / /\  | | | |_)
/_/--\ |_| |_| \_ v1.xx.x, built with Go 1.21.x

watching .
watching cmd
watching cmd/api
...
building...
running...
```

### Step 2: Edit File
Buka file `cmd/api/main.go` dan edit welcome message:

```go
"message": "Welcome to Go JWT REST API - HOT RELOAD WORKS!",
```

### Step 3: Save File
Save file (Ctrl+S)

### Step 4: Lihat Terminal
Air akan otomatis:
```
building...
running...
```

### Step 5: Test
```bash
curl http://localhost:8080/
```

Response sudah berubah tanpa restart manual! 🎉

## ⚙️ Konfigurasi Air (.air.toml)

File `.air.toml` sudah dibuat dengan konfigurasi:

- **Build command**: `go build -o ./tmp/main ./cmd/api/main.go`
- **Watch**: Semua file `.go`, `.html`, `.tpl`, `.tmpl`
- **Exclude**: `tmp/`, `vendor/`, `bin/`, test files
- **Delay**: 1 second sebelum rebuild
- **Clear screen**: Otomatis clear screen saat rebuild

## 🔧 Kustomisasi

Edit `.air.toml` untuk custom behavior:

```toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/api/main.go"  # Build command
  bin = "./tmp/main"                                 # Binary location
  delay = 1000                                       # Delay ms
  exclude_dir = ["assets", "tmp", "vendor"]          # Ignore folders
  include_ext = ["go", "tpl", "html"]               # Watch extensions
```

## 📊 Perbandingan

### **Tanpa Air (Manual Reload):**
```bash
1. Edit code
2. Ctrl+C untuk stop server
3. go run cmd/api/main.go
4. Wait...
5. Test
Repeat 😩
```

### **Dengan Air (Hot Reload):**
```bash
1. Edit code
2. Save (Ctrl+S)
3. Done! ✨
Air handle sisanya otomatis
```

## 💡 Tips

1. **Development**: Selalu gunakan Air untuk development
2. **Production**: JANGAN gunakan Air, build binary dengan `go build`
3. **Performance**: Air consume sedikit CPU untuk file watching
4. **Logs**: Air akan show build errors di terminal
5. **Clean**: Air otomatis clean tmp/ saat exit
6. **Test**: Test langsung di browser atau Postman tanpa restart

## 🐛 Troubleshooting

**Problem**: Air tidak detect perubahan
**Solution**: Check `.air.toml`, pastikan folder tidak di exclude

**Problem**: Build error terus muncul
**Solution**: Fix code error, Air akan auto rebuild ketika sudah benar

**Problem**: Port already in use
**Solution**: Kill process di port 8080:
```bash
lsof -ti:8080 | xargs kill -9
```

## ⚡ Shortcut

Tambahkan ke `.bashrc`:
```bash
alias air='~/go/bin/air'
alias dev='cd ~/Project/Golang/gojwt-rest-api && ~/go/bin/air'
```

Reload:
```bash
source ~/.bashrc
```

Sekarang tinggal ketik:
```bash
dev  # Langsung run project dengan hot reload!
```

---

Happy coding with hot reload! 🚀
