# Backend for Carport

## Langkah 1: Clone Repositori

1. Buka terminal atau command prompt.
2. Pindah ke direktori di mana Anda ingin menyimpan proyek Golang.
3. Jalankan perintah berikut untuk melakukan clone repositori:

   ```bash
   git clone [URL_REPOSITORI]
   ```

   Gantilah `[URL_REPOSITORI]` dengan URL repositori Golang yang ingin Anda clone.

## Langkah 2: Install Library dan Setup env

1. Pindah ke direktori proyek yang baru saja di-clone:

   ```bash
   cd carport-backend
   ```

2. Jalankan perintah berikut untuk menginstal library dan dependensi:

   ```bash
   go mod tidy
   ```

   Perintah ini akan menginstal semua library dan dependensi yang didefinisikan dalam file `go.mod`.

3. Buat file .env berisi

   ```
   PORT=[Untuk menjalankan aplikasi]
   BD="host=[host] user=[username_db] password=[password_db] dbname=[name_db] port=[port_db] sslmode=disable"
   SECRET=[jwt_scret]
   ```

## Langkah 3: Menjalankan Aplikasi dengan CompileDaemon

1. Install CompileDaemon jika belum terpasang:

   ```bash
   go get github.com/githubnemo/CompileDaemon
   ```

   ```bash
   go install github.com/githubnemo/CompileDaemon
   ```

2. Jalankan aplikasi menggunakan CompileDaemon:

   ```bash
   CompileDaemon -command="./carport-backend"
   ```

3. Aplikasi akan mulai berjalan, dan CompileDaemon akan memantau perubahan file secara otomatis. Setiap kali Anda menyimpan perubahan pada file, CompileDaemon akan mengompilasi ulang dan menjalankan aplikasi.

## Langkah 4: Menguji Aplikasi

1. Buka web browser atau alat pengujian API (seperti [Postman](https://www.postman.com/)).
2. Akses aplikasi yang berjalan di [http://localhost:PORT](http://localhost:PORT) (gantilah `PORT` sesuai dengan konfigurasi aplikasi Anda).

   Contoh: [http://localhost:3000](http://localhost:3000)

   Aplikasi Anda sekarang harus berjalan dan dapat diakses di alamat yang ditentukan.
