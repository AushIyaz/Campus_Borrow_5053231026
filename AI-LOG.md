# AI-LOG

Catat setiap penggunaan AI yang menghasilkan kode. Jika tidak memakai AI sama sekali, tulis satu baris: **tidak memakai AI**.

Log yang jujur tidak mengurangi nilai, sebanyak apa pun AI dipakai. Yang mengurangi nilai adalah log yang tidak sesuai dengan kode, atau isi log yang tidak bisa dijelaskan.

## Deteksi bentrok slot dan stok (B5)
- Prompt: "buat fungsi Go untuk mengecek irisan slot dan validasi stok maksimum per slot"
- Dipakai: fungsi checkBentrok di peminjaman.go
- Diubah: Memastikan pengecekan dilakukan iteratif per slot dalam rentang `slot_mulai` hingga `slot_selesai` dan mendeteksi total unit aktif terhadap stok sumber daya.

## Dashboard Ringkasan Harian (Change Request A)
- Prompt: "buat handler Go untuk GET /ringkasan dengan filter tanggal, menghitung total unit slot dan per sumber daya terurut ID"
- Dipakai: fungsi handleGetRingkasan di ringkasan.go
- Diubah: Menyesuaikan struktur JSON response agar persis dengan kontrak Change Request A serta menambahkan validasi `time.Parse`.