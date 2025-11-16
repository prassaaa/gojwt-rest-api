# Fitur Refresh Token & Rotasi Token

## Ringkasan

API ini sekarang mengimplementasikan mekanisme refresh token yang aman dengan rotasi dan pencabutan token otomatis untuk keamanan yang lebih baik.

## Fitur Utama

### 1. **Sistem Token Ganda**
- **Access Token**: Berumur pendek (default 15 menit)
- **Refresh Token**: Berumur panjang (default 7 hari)
- Access token digunakan untuk permintaan API
- Refresh token digunakan untuk mendapatkan access token baru

### 2. **Rotasi Token Otomatis**
- Setiap proses refresh menghasilkan pasangan token baru
- Refresh token lama secara otomatis dicabut
- Mencegah serangan penggunaan kembali token

### 3. **Pelacakan Keluarga Token (Token Family)**
- Semua token dalam satu rantai refresh berbagi ID keluarga yang sama
- Mendeteksi penggunaan kembali token yang mencurigakan
- Mencabut seluruh keluarga token jika terdeteksi penggunaan kembali

### 4. **Pencabutan Token yang Aman**
- Logout secara eksplisit akan membatalkan refresh token
- Opsi untuk logout dari semua perangkat
- Daftar hitam (blacklist) token yang didukung oleh database

## Endpoint API

### 1. Login (Diperbarui)
```bash
POST /api/v1/auth/login
```

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "login berhasil",
  "data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "user@example.com",
      "is_admin": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "random_secure_token_here",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

### 2. Refresh Token (Baru)
```bash
POST /api/v1/auth/refresh
```

**Request:**
```json
{
  "refresh_token": "your_refresh_token_here"
}
```

**Response:**
```json
{
  "success": true,
  "message": "token berhasil diperbarui",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "new_random_secure_token_here",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

**Respons Error:**
- `401 Unauthorized`: Refresh token tidak valid atau kedaluwarsa
- `401 Unauthorized`: Terdeteksi penggunaan kembali token (pelanggaran keamanan)

### 3. Logout (Baru)
```bash
POST /api/v1/auth/logout
```

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "refresh_token": "your_refresh_token_here"
}
```

**Response:**
```json
{
  "success": true,
  "message": "logout berhasil",
  "data": null
}
```

## Fitur Keamanan

### Deteksi Penggunaan Kembali Token
Jika refresh token yang sudah pernah digunakan dikirimkan kembali:
1. Sistem mendeteksi upaya penggunaan kembali
2. Semua token dalam keluarga token tersebut segera dicabut
3. Pengguna harus login kembali

Ini melindungi dari skenario pencurian token.

### Pelacakan Keluarga Token
Setiap login membuat keluarga token baru. Semua proses refresh berikutnya mempertahankan ID keluarga yang sama, memungkinkan sistem untuk melacak dan mencabut token terkait jika terdeteksi aktivitas mencurigakan.

### Skema Database

**Tabel `refresh_tokens`:**
```sql
CREATE TABLE refresh_tokens (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  token VARCHAR(500) UNIQUE NOT NULL,
  token_family VARCHAR(100) NOT NULL,
  expires_at DATETIME NOT NULL,
  is_revoked BOOLEAN DEFAULT FALSE,
  revoked_at DATETIME,
  replaced_by VARCHAR(500),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_token (token),
  INDEX idx_token_family (token_family),
  INDEX idx_expires_at (expires_at)
);
```

**Tabel `token_blacklist`:**
```sql
CREATE TABLE token_blacklist (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  token VARCHAR(500) UNIQUE NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_token (token),
  INDEX idx_expires_at (expires_at)
);
```

## Konfigurasi

Tambahkan variabel lingkungan ini ke file `.env` Anda:

```bash
# Konfigurasi JWT
JWT_SECRET=your-super-secret-key
JWT_ACCESS_EXPIRATION=15m      # Masa berlaku access token
JWT_REFRESH_EXPIRATION=168h    # Masa berlaku refresh token (7 hari)
```

## Panduan Implementasi Klien

### Alur Aplikasi Web

```javascript
// 1. Login dan simpan token
const login = async (email, password) => {
  const response = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });

  const data = await response.json();

  // Simpan token dengan aman
  localStorage.setItem('access_token', data.data.access_token);
  localStorage.setItem('refresh_token', data.data.refresh_token);

  return data;
};

// 2. Permintaan API dengan refresh token otomatis
const apiRequest = async (url, options = {}) => {
  let accessToken = localStorage.getItem('access_token');

  options.headers = {
    ...options.headers,
    'Authorization': `Bearer ${accessToken}`
  };

  let response = await fetch(url, options);

  // Jika access token kedaluwarsa, perbarui
  if (response.status === 401) {
    const refreshToken = localStorage.getItem('refresh_token');

    const refreshResponse = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken })
    });

    if (refreshResponse.ok) {
      const refreshData = await refreshResponse.json();

      // Perbarui token yang disimpan
      localStorage.setItem('access_token', refreshData.data.access_token);
      localStorage.setItem('refresh_token', refreshData.data.refresh_token);

      // Coba lagi permintaan asli dengan token baru
      options.headers['Authorization'] = `Bearer ${refreshData.data.access_token}`;
      response = await fetch(url, options);
    } else {
      // Refresh gagal, alihkan ke halaman login
      window.location.href = '/login';
    }
  }

  return response;
};

// 3. Logout
const logout = async () => {
  const accessToken = localStorage.getItem('access_token');
  const refreshToken = localStorage.getItem('refresh_token');

  await fetch('/api/v1/auth/logout', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ refresh_token: refreshToken })
  });

  // Hapus dari local storage
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');

  window.location.href = '/login';
};
```

## Praktik Terbaik

1. **Simpan refresh token dengan aman**
   - Gunakan cookie httpOnly untuk aplikasi web
   - Gunakan penyimpanan aman untuk aplikasi seluler
   - Jangan pernah mengekspos refresh token di URL

2. **Tangani kedaluwarsa token dengan baik**
   - Implementasikan refresh token otomatis
   - Alihkan ke halaman login jika refresh gagal
   - Tampilkan pesan kesalahan yang sesuai

3. **Implementasikan logout yang benar**
   - Selalu panggil endpoint logout
   - Hapus semua token yang disimpan
   - Batalkan sesi pengguna

4. **Pantau aktivitas mencurigakan**
   - Catat upaya penggunaan kembali token
   - Beri peringatan pada beberapa upaya refresh yang gagal
   - Implementasikan pembatasan laju (rate limiting)

## Pemeliharaan

### Membersihkan Token Kedaluwarsa

Tambahkan tugas terjadwal untuk membersihkan token yang kedaluwarsa:

```go
// Jalankan setiap hari atau sesuai kebutuhan
tokenRepo.DeleteExpiredRefreshTokens()
tokenRepo.DeleteExpiredBlacklistTokens()
```

### Logout dari Semua Perangkat

Untuk mengimplementasikan "logout dari semua perangkat", hapus komentar pada baris ini di `user_service.go`:

```go
// Di dalam metode Logout
if err := s.tokenRepo.RevokeAllUserRefreshTokens(userID); err != nil {
    return err
}
```

## Pengujian

Jalankan skrip pengujian yang disediakan:

```bash
./test_refresh_token.sh
```

Ini akan menguji:
- Login dan pembuatan token
- Refresh token
- Rotasi token (token lama menjadi tidak valid)
- Logout
- Deteksi penggunaan kembali token

## Panduan Migrasi

Jika melakukan upgrade dari versi sebelumnya:

1. Perbarui file `.env` dengan konfigurasi JWT baru
2. Jalankan migrasi database (otomatis saat server dimulai)
3. Perbarui kode klien untuk menangani struktur token baru
4. Uji secara menyeluruh sebelum menerapkan ke produksi

## Pemecahan Masalah

### Error "Terdeteksi penggunaan kembali token"
- Ini berarti refresh token digunakan dua kali
- Semua token dalam keluarga tersebut sekarang dicabut
- Pengguna harus login kembali
- Ini adalah fitur keamanan, bukan bug

### Refresh token kedaluwarsa
- Pengguna harus login kembali
- Pertimbangkan untuk menambah `JWT_REFRESH_EXPIRATION` jika perlu
- Implementasikan fitur "ingat saya" jika diinginkan

### Access token ditolak
- Token mungkin kedaluwarsa (perbarui)
- Token mungkin salah format
- Rahasia JWT mungkin telah berubah
- Periksa log server untuk detailnya