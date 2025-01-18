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
   cd voucher-backend
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
   CompileDaemon -command="./voucher-backend"
   ```

3. Aplikasi akan mulai berjalan, dan CompileDaemon akan memantau perubahan file secara otomatis. Setiap kali Anda menyimpan perubahan pada file, CompileDaemon akan mengompilasi ulang dan menjalankan aplikasi.

## Langkah 4: Menguji Aplikasi

1. Buka web browser atau alat pengujian API (seperti [Postman](https://www.postman.com/)).
2. Akses aplikasi yang berjalan di [http://localhost:PORT](http://localhost:PORT) (gantilah `PORT` sesuai dengan konfigurasi aplikasi Anda).

   Contoh: [http://localhost:3000](http://localhost:3000)

   Aplikasi Anda sekarang harus berjalan dan dapat diakses di alamat yang ditentukan.


## Langkah 5: Endpoint dan format input data

1. Add Brand
   Endpoint : [http://localhost:3000/brand]
   input json :
   {
      "name" : "Indomaret"
   }

2. Add Voucher
   Endpoint : [http://localhost:3000/voucher]
   input json :
   {
    "name" : "Voucher Indomaret",
    "discount" : 50,
    "point" : 200,
    "quantity" : 100,
    "brand_id" : 2
   }

3. Get Voucher by Voucher ID
   Endpoint : [http://localhost:3000/voucher/:id] 
   example Endpoint : [http://localhost:3000/voucher/2]

4. Get Voucher by Brand ID
   Endpoint : [http://localhost:3000/voucher/brand/:id] 
   example Endpoint : [http://localhost:3000/voucher/brand/2]

5. Redeem Voucher (Transactions)
   Endpoint : [http://localhost:3000/transaction/redemption]
   input json :
   {  
      "customer_id": 1,  
      "vouchers": [  
         {  
            "voucher_id": 1,  
            "quantity": 2  
         },  
         {  
            "voucher_id": 2,  
            "quantity": 1  
         }  
      ]  
   }  


6. Get Transactions by ID
   Endpoint : [http://localhost:3000/transaction/redemption/:id]
   example Endpoint : [http://localhost:3000/transaction/redemption/1]
