# PROSES

## 1. Persona dan masalahnya
Persona B ditujukan untuk mahasiswa peminjam alat kampus. Masalah yang sering terjadi adalah beberapa mahasiswa memborong peminjaman alat secara berlebihan sehingga mahasiswa lain tidak kebagian. Fitur khas Persona B mengatasi hal ini dengan memberikan batasan durasi maksimal 3 slot per peminjaman serta kuota maksimum 2 peminjaman aktif per orang pada tanggal yang sama.

## 2. Rancangan struktur data
Menggunakan struct `SumberDaya` dan `Peminjaman` yang disimpan di dalam slice memori (`sumberDayas` dan `peminjamans`) di dalam struct `AppMemory`. Dipilih karena penyimpanan in-memory memenuhi syarat kecepatan tanpa database eksternal, serta dilindungi oleh `sync.Mutex` untuk menjamin keamanan *concurrent requests*.

## 3. Satu keputusan desain
Saat mengerjakan Change Request A, saya memutuskan untuk menggunakan penelusuran tunggal (*single-pass loop*) pada slice peminjaman untuk mengumpulkan total peminjaman dan menghitung `total_unit_slot`, lalu menggabungkannya dengan seluruh sumber daya menggunakan map pelacak sementara. Keputusan ini diambil agar performa tetap cepat dan efisien tanpa iterasi berulang ke memori.

## 4. Bug atau kendala (minimal dua)
### Bug 1
- Gejala: Terjadi *deadlock* saat memanggil fungsi helper kuota di dalam handler.
- Cara menemukan: Melakukan pengujian bersamaan (*concurrent test*) dan server berhenti merespons.
- Penyebab dan solusi: Memastikan bahwa pemanggilan `db.mu.Lock()` hanya dilakukan sekali di level handler utama sebelum fungsi helper dipanggil, bukan mengunci mutex ganda di dalam fungsi helper.

### Bug 2
- Gejala: Perhitungan jumlah slot pada change request salah selisih 1 jam.
- Cara menemukan: Menguji peminjaman dari slot 2 sampai 4 dan mendapati total unit slot kurang.
- Penyebab dan solusi: Memperbaiki rumus hitung slot dari `slot_selesai - slot_mulai` menjadi `slot_selesai - slot_mulai + 1`.

## 5. Change request
Menambahkan endpoint `GET /ringkasan?tanggal=YYYY-MM-DD` untuk menyediakan data dashboard ringkasan harian bagi pengelola kampus. Bagian kode yang diubah meliputi penambahan rute di `handler.go` serta pembuatan file `ringkasan.go` yang menghitung `jumlah_peminjaman`, `total_unit_slot`, dan daftar `per_sumber_daya` secara terurut. Keputusan desainnya adalah menyertakan semua sumber daya meski bernilai 0 peminjaman pada hari tersebut sesuai aturan kontrak.