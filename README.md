# movie-api

Go REST API อ่านข้อมูลภาพยนตร์จาก PostgreSQL (Supabase) สำหรับหน้าบ้าน React — deploy บน Vercel

## Endpoints

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/api/health` | ตรวจสถานะ |
| GET | `/api/movies?q=&genre=&limit=20&offset=0` | รายการหนัง (ค้นหาจากชื่อ / กรอง genre / แบ่งหน้า) |
| GET | `/api/movies/{id}` | หนังรายตัว |

Response: `{ "data": [...] }` หรือ `{ "error": "..." }`

## 1. ตั้งค่า DB บน Supabase
1. สร้าง project → **SQL Editor** → วางและรัน `db/schema.sql`
2. ไปที่ **Project Settings → Database → Connection string → URI** (เลือก *Transaction pooler*, port 6543) คัดลอกเก็บไว้

## 2. รันในเครื่อง
```bash
cp .env.example .env      # แก้ DATABASE_URL
go mod tidy
go run ./cmd/server
curl localhost:8080/api/movies
```

## 3. Deploy บน Vercel
1. Push repo ขึ้น GitHub
2. Vercel → **Add New Project** → import repo (Framework: Other)
3. **Environment Variables** ใส่ `DATABASE_URL` (และ `CORS_ORIGINS` เป็น URL ของหน้าบ้าน ถ้าไม่อยากเปิด `*`)
4. Deploy → ทดสอบ `https://<app>.vercel.app/api/movies`

## ใช้จาก React
```js
const API = import.meta.env.VITE_API_URL; // https://<app>.vercel.app
const res = await fetch(`${API}/api/movies?q=dark`);
const { data } = await res.json();
```
