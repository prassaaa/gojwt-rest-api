# Quick Start Guide - Generate JWT Secret

## Cara Termudah

Jalankan salah satu command berikut untuk generate JWT secret:

### 1. Menggunakan Make (Recommended)
```bash
make generate-secret
```

### 2. Menggunakan Go
```bash
go run cmd/tools/generate_secret.go
```

### 3. Menggunakan OpenSSL
```bash
openssl rand -base64 32
```

## Setelah Generate

1. Copy output secret yang dihasilkan
2. Buka file `.env`
3. Paste secret ke variable `JWT_SECRET`

Contoh:
```env
JWT_SECRET=Xxxxxxxxxx=
```

## PENTING! ⚠️

- **JANGAN** commit file `.env` ke Git (sudah ada di .gitignore)
- **JANGAN** share secret ke orang lain
- Gunakan secret yang **berbeda** untuk development, staging, dan production
- Minimal 32 karakter untuk keamanan yang baik

## Troubleshooting

**Q: Bagaimana jika lupa JWT_SECRET?**
A: Generate secret baru dengan command di atas. User yang sudah login harus login ulang.

**Q: Apakah harus generate ulang setiap kali run aplikasi?**
A: TIDAK! Generate sekali saja dan simpan di `.env`. Secret harus konsisten.

**Q: Bisakah pakai password biasa?**
A: Sangat tidak disarankan! Gunakan random generated secret untuk keamanan maksimal.
